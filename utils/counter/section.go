package counter

import (
	"../filesystem"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

func NewSection(username string, repository string) *Section {
	c := &Section{
		Username:   username,
		Repository: repository,
		Total:      0,
		CreatedAt:  time.Now(),
		Entries:    make(map[string]*Entry),
	}
	_ = c.Load("")
	return c
}

func (s *Section) GetKey() string {
	return s.Username + "/" + s.Repository
}

func (s *Section) String() string {
	dateFormat := "2006-01-02 15:04:05"
	return strings.Join([]string{
		s.Username,
		s.Repository,
		fmt.Sprintf("%d", s.Total),
		s.CreatedAt.Format(dateFormat),
		s.UpdatedAt.Format(dateFormat),
	}, ",")
}

func (s *Section) GetToken() string {
	h := sha256.New()
	h.Write([]byte(s.GetKey()))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (s *Section) AddEntry(entry *Entry) bool {
	if _, ok := s.Entries[entry.Hash]; !ok {
		s.Entries[entry.Hash] = entry
		s.Increment()
		return true
	}
	return false
}

func (s *Section) Increment() {
	s.Total += 1
	s.UpdatedAt = time.Now()
}

func (s *Section) initFile(filename string) {
	filesystem.CreateDirectory("data")
	if len(filename) == 0 {
		dir, _ := os.Getwd()
		filename = path.Join(dir, "data", s.GetToken() + ".json")
		_ = s.Load(filename)
	}
	s.File = filename
}


func (s *Section) Load(filename string) error {
	s.initFile(filename)

	if _, err := os.Stat(filename); err == nil {

		content, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		err = json.Unmarshal(content, s)
		if err != nil {
			return err
		}

	} else {
		_ = s.Save()
	}
	return nil
}

func (s *Section) Save() error {
	if len(s.File) == 0 {
		s.initFile("")
	}

	file, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(s.File, file, 0644)
	if err != nil {
		return err
	}

	return nil
}