package server

import (
	"../utils/config"
	"../utils/counter"
	"../utils/filesystem"
	"../utils/log"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"io/ioutil"
	olog "log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
)

// Create a custom visitor struct which holds the rate limiter for each
// visitor and the last time that the visitor was seen.

type Server struct {
	Config  *config.Config
	Counter *counter.Counter

	Host string
	Port int

	// Registered clients.
	clients       map[*Client]bool
	subscriptions map[string]map[*Client]bool

	// Register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister chan *Client

	activities chan *counter.Section

	RateLimit *RateLimit
	Visitors  map[string]*Visitor
	mx        *sync.RWMutex

	template *template.Template
	assets *assetfs.AssetFS

	Upgrader websocket.Upgrader

	Api *ApiHandler
}

type ApiHandler struct {
	cors *cors.Cors
}

func NewServerConfig(c *config.Config, assets *assetfs.AssetFS) *Server {
	parts := strings.Split(c.ServerAddr, ":")
	host := parts[0]
	port, err := strconv.Atoi(parts[1])

	if err != nil || port <= 0 {
		print("Invalid Socket provided")
		os.Exit(1)
	}

	defaultConfig := config.DefaultConfig()
	if c.PingPeriod <= 0 {
		c.PingPeriod = defaultConfig.PingPeriod
	}
	if c.PongWait <= 0 {
		c.PongWait = defaultConfig.PongWait
	}
	if c.PongWait <= c.PingPeriod {
		c.PongWait = c.PingPeriod + defaultConfig.PongWait
	}

	s := &Server{
		Config:  c,
		Counter: counter.NewCounter(c.SessionLifetime),

		Host: host,
		Port: port,

		register:      make(chan *Client),
		unregister:    make(chan *Client),
		activities:    make(chan *counter.Section),
		clients:       make(map[*Client]bool),
		subscriptions: make(map[string]map[*Client]bool),

		Upgrader: websocket.Upgrader{
			ReadBufferSize:    4096,
			WriteBufferSize:   4096,
			EnableCompression: true,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},

		RateLimit: NewRateLimit(c.RateLimitLimit, c.RateLimitBurst, c.RateLimitInterval),
		Api:       &ApiHandler{},
		mx:        &sync.RWMutex{},
	}

	s.assets = assets
	s.template = ParseTemplates("htdocs/template")

	c.LogOutput = os.Stdout

	if c.LogToStdout {
		olog.SetOutput(c.LogOutput)
	} else if c.LogOutputFile != "" {
		_, _ = filesystem.MakeDir(c.LogOutputFile)
		_ = ioutil.WriteFile(c.LogOutputFile, []byte(""), 0644)
		c.LogOutput, _ = os.OpenFile(c.LogOutputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		olog.SetOutput(c.LogOutput)
	}
	if !c.LogTimestamp {
		olog.SetFlags(0)
	}

	return s
}

func (s *Server) listen() {
	for {
		select {
		// Register new clients
		case client := <-s.register:
			s.clients[client] = true
			log.Info("Client connected!")
		// Register new clients
		case section := <-s.activities:
			sectionKey := section.GetKey()
			if subscribers, ok := s.subscriptions[sectionKey]; ok {
				for client, state := range subscribers {
					if state {
						client.SendString(sectionKey)
					}
				}
			}
			if subscribers, ok := s.subscriptions["all"]; ok {
				for client, state := range subscribers {
					if state {
						client.SendString(sectionKey)
					}
				}
			}
		// Unregister an existing client
		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				_ = client.Close()
				log.Info("Client disconnected!")
			}
		}
	}
}

func (s *Server) Start() {
	f, err := s.NewHandler()
	if err != nil {
		log.Fatal(err)
	}
	go s.Counter.Run()
	go s.listen()
	if s.Config.ServerAddr != "" {
		go s.runServer(f)
	}
	if s.Config.TLSServerAddr != "" {
		go s.runTLSServer(f)
	}
	select {}
}

func ParseTemplates(_path string) *template.Template {
	tpl := template.New("")
	err := filepath.Walk(_path, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".tmpl") {
			_, err = tpl.ParseFiles(path)
			if err != nil {
				log.Error(err)
			}
		}

		return err
	})

	if err != nil {
		panic(err)
	}

	return tpl
}
