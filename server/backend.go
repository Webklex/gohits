package server

import (
	"../utils/log"
	"github.com/go-web/httplog"
	"github.com/go-web/httpmux"
	"github.com/rs/cors"
	"net/http"
	"strings"

	"github.com/fiorix/go-listener/listener"
)

type writerFunc func(w http.ResponseWriter, r *http.Request)

func (s *Server) NewHandler() (http.Handler, error) {
	s.initCors()
	mc := httpmux.DefaultConfig
	if err := s.initMiddlewares(&mc); err != nil {
		return nil, err
	}

	mux := httpmux.NewHandler(&mc)

	mux.HandleFunc("GET", "/", func(w http.ResponseWriter, req *http.Request) {
		if err := s.template.ExecuteTemplate(w, "index", s.indexResponse(req)); err != nil {
			log.Error(err)
		}
	})

	mux.GET("/svg/:username/:repository", s.registerHandler(s.badgeResponse))
	mux.GET("/json/:username/:repository", s.registerHandler(s.jsonResponse))
	mux.GET("/xml/:username/:repository", s.registerHandler(s.xmlResponse))
	mux.GET("/csv/:username/:repository", s.registerHandler(s.csvResponse))

	mux.GET("/ws", s.registerSocketHandler())

	return mux, nil
}

func (s *Server) registerHandler(writer writerFunc) http.HandlerFunc {
	return s.Api.cors.Handler(s.handleRequest(writer)).ServeHTTP
}

func (s *Server) registerSocketHandler() http.HandlerFunc {
	return s.Api.cors.Handler(s.socketHandler()).ServeHTTP
}

func (s *Server) socketHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := s.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error("upgrade:", err)
			return
		}

		client := NewClient(s, ws)
		s.register <- client

		// Allow collection of memory referenced by the caller by doing all work in
		// new goroutines.

		client.Listen()
		client.SendString("")
	}
}

func (s *Server) initCors() {
	s.Api.cors = cors.New(cors.Options{
		AllowedOrigins:   strings.Split(s.Config.CORSOrigin, ","),
		AllowedMethods:   []string{"GET"},
		AllowCredentials: true,
	})
}

func (s *Server) listenerOpts() []listener.Option {
	var opts []listener.Option
	if s.Config.FastOpen {
		opts = append(opts, listener.FastOpen())
	}
	if s.Config.Naggle {
		opts = append(opts, listener.Naggle())
	}
	return opts
}

func (s *Server) runServer(f http.Handler) {
	if !s.Config.Silent {
		log.Info("http server starting on", s.Config.ServerAddr)
	}
	ln, err := listener.New(s.Config.ServerAddr, s.listenerOpts()...)
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{
		Handler:      f,
		ReadTimeout:  s.Config.ReadTimeout,
		WriteTimeout: s.Config.WriteTimeout,
		ErrorLog:     s.Config.ErrorLogger(),
	}
	log.Fatal(srv.Serve(ln))
}

func (s *Server) runTLSServer(f http.Handler) {
	log.Info("https server starting on", s.Config.TLSServerAddr)
	opts := s.listenerOpts()
	if s.Config.HTTP2 {
		opts = append(opts, listener.HTTP2())
	}
	if s.Config.LetsEncrypt {
		if s.Config.LetsEncryptHosts == "" {
			log.Fatal("must set at least one host using --letsencrypt-hosts")
		}
		opts = append(opts, listener.LetsEncrypt(
			s.Config.LetsEncryptCacheDir,
			s.Config.LetsEncryptEmail,
			strings.Split(s.Config.LetsEncryptHosts, ",")...,
		))
	} else {
		opts = append(opts, listener.TLS(s.Config.TLSCertFile, s.Config.TLSKeyFile))
	}
	ln, err := listener.New(s.Config.TLSServerAddr, opts...)
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{
		Addr:         s.Config.TLSServerAddr,
		Handler:      f,
		ReadTimeout:  s.Config.ReadTimeout,
		WriteTimeout: s.Config.WriteTimeout,
		ErrorLog:     s.Config.ErrorLogger(),
		TLSConfig:    ln.TLSConfig(),
	}
	log.Fatal(srv.Serve(ln))
}

func (s *Server) initMiddlewares(mc *httpmux.Config) error {
	mc.Prefix = s.Config.APIPrefix
	mc.NotFound = s.guiMiddleware(s.Config.GuiDir)
	if s.Config.UseXForwardedFor {
		mc.UseFunc(httplog.UseXForwardedFor)
	}
	if !s.Config.Silent {
		mc.UseFunc(httplog.ApacheCombinedFormat(s.Config.AccessLogger()))
	}
	if s.Config.HSTS != "" {
		mc.UseFunc(hstsMiddleware(s.Config.HSTS))
	}
	if s.Config.RateLimitLimit > 0 {
		mc.Use(s.rateLimitMiddleware)
	}
	return nil
}

func (s *Server) messageHandler(message *Message) {
	cmd := &Command{}
	message.Decode(cmd)
	switch cmd.Name {
	case "unsubscribe":
		sectionKey := ""
		if cmd.Payload == "all" {
			sectionKey = cmd.Payload
		}else{
			if section := s.Counter.GetSectionByKey(sectionKey); section != nil {
				sectionKey = section.GetKey()
			}
		}
		if len(sectionKey) > 0 {

			s.mx.Lock()
			if clients, ok := s.subscriptions[sectionKey]; ok {
				if _, ok := clients[message.Client]; ok {
					delete(s.subscriptions[sectionKey], message.Client)
				}
				if len(s.subscriptions[sectionKey]) == 0 {
					delete(s.subscriptions, sectionKey)
				}
			}
			s.mx.Unlock()

			message.Client.SendString("successfully unsubscribed from " + sectionKey)
		}else{
			message.Client.SendString("invalid command")
		}
	case "subscribe":
		sectionKey := ""
		if cmd.Payload == "all" {
			sectionKey = cmd.Payload
		}else{
			if section := s.Counter.GetSectionByKey(sectionKey); section != nil {
				sectionKey = section.GetKey()
			}
		}
		if len(sectionKey) > 0 {

			s.mx.Lock()
			if _, ok := s.subscriptions[sectionKey]; !ok {
				s.subscriptions[sectionKey] = make(map[*Client]bool)
			}
			s.subscriptions[sectionKey][message.Client] = true
			s.mx.Unlock()

			message.Client.SendString("successfully subscribed to " + sectionKey)
		}else{
			message.Client.SendString("invalid command")
		}
	default:
		message.Client.SendString("invalid command")
	}
}

func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := s.RateLimit.GetLimiter(r.RemoteAddr)
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func hstsMiddleware(policy string) httpmux.MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.TLS == nil {
				return
			}
			w.Header().Set("Strict-Transport-Security", policy)
			next(w, r)
		}
	}
}

func (s *Server) guiMiddleware(path string) http.Handler {
	handler := http.NotFoundHandler()
	if path != "" {
		handler = http.FileServer(s.assets)
	}

	return handler
}
