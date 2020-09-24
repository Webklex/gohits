package server

import (
	"../utils/counter"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/go-web/httpmux"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func (s *Server) handleRequest(writer writerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writer(w, r)
	}
}

func (s *Server) indexResponse(r *http.Request) interface{} {
	return s.Config
}

func (s *Server) jsonResponse(w http.ResponseWriter, r *http.Request) {
	content, err := json.MarshalIndent(s.getSection(r), "", "\t")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if n, err := io.WriteString(w, string(content)); err != nil || n <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func (s *Server) xmlResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")

	x := xml.NewEncoder(w)
	x.Indent("", "\t")
	if err := x.Encode(s.getSection(r)); err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if n, err := w.Write([]byte{'\n'}); err != nil || n <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func (s *Server) csvResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	if n, err := io.WriteString(w, s.getSection(r).String()); err != nil || n <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func (s *Server) badgeResponse(w http.ResponseWriter, r *http.Request) {

	SetHeaders(w)
	counterStr := s.count(r)
	badge := fmt.Sprintf(`<?xml version="1.0"?>
	<svg xmlns="http://www.w3.org/2000/svg" width="80" height="20">
		<rect width="30" height="20" fill="#555"/>
		<rect x="30" width="50" height="20" fill="#4c1"/>
	
		<rect rx="3" width="80" height="20" fill="transparent"/>
		<g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
			<text x="15" y="14">hits</text>
			<text x="54" y="14">%s</text>
		</g>
		<!-- If you have time to help us add descriptive comments please PR! -->
	</svg>`, counterStr)

	if n, err := io.WriteString(w, badge); err != nil || n <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func (s *Server) badgeHeadResponse(w http.ResponseWriter, r *http.Request) {
	SetHeaders(w)
	_ = s.count(r)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getSection(r *http.Request) *counter.Section {
	username := sanitize(httpmux.Params(r).ByName("username"))
	repository := sanitize(httpmux.Params(r).ByName("repository"))

	return s.Counter.GetSection(username, repository)
}

func (s *Server) count(r *http.Request) string {

	section := s.getSection(r)
	userAgent := r.Header.Get("User-Agent")
	if strings.Contains(userAgent, "camo"){
		// Treat camouflaged request as new hit - is likely cached anyways
		s.mx.Lock()
		section.Increment()
		s.mx.Unlock()
		s.activities <- section
	}else{
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		h := sha256.New()
		h.Write([]byte(host + userAgent))
		token := fmt.Sprintf("%x", h.Sum(nil))

		entry := counter.NewEntry(token)

		s.mx.Lock()
		if s.Counter.AddEntry(section, entry) {
			s.activities <- section
		}
		s.mx.Unlock()
	}

	total := float64(section.Total)

	counterStr := fmt.Sprintf("%.0f", total)
	if total > 1000000 {
		counterStr = fmt.Sprintf("%.2fm", total/1000000)
	} else if total > 10000 {
		counterStr = fmt.Sprintf("%.0fk", total/1000)
	} else if total > 1000 {
		counterStr = fmt.Sprintf("%.2fk", total/1000)
	}

	return counterStr
}

func sanitize(in string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9\\-_.]+")
	return reg.ReplaceAllString(in, "_")
}

func SetHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/svg+xml;charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=60, s-maxage=60")

	loc, _ := time.LoadLocation("UTC")
	t := time.Now().In(loc)
	w.Header().Set("date", t.Format("Mon, 2 Jan 2006 15:04:05 MST"))
	w.Header().Set("expires", t.Add(time.Minute).Format("Mon, 2 Jan 2006 15:04:05 MST"))
}
