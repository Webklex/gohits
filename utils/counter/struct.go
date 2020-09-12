package counter

import (
	"encoding/xml"
	"sync"
	"time"
)

type Counter struct {
	Sections map[string]*Section `json:"sections"`
	File     string              `json:"-"`
	Duration time.Duration       `json:"-"`
	mx       *sync.RWMutex       `json:"-"`
}

type Section struct {
	XMLName    xml.Name          `xml:"Section" json:"-"`
	Username   string            `json:"username"`
	Repository string            `json:"repository"`
	Total      int64             `json:"total"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Entries    map[string]*Entry `xml:"-" json:"-"`
	File       string            `xml:"-" json:"-"`
}

type Entry struct {
	Hash      string
	Timestamp time.Time
}
