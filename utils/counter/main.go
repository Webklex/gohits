package counter

import (
	"sync"
	"time"
)

func NewCounter(duration time.Duration) *Counter {
	c := &Counter{
		Duration: duration,
		mx:        &sync.RWMutex{},
	}
	if c.Sections == nil {
		c.Sections = make(map[string]*Section)
	}
	for k, _ := range c.Sections {
		c.Sections[k].Entries = make(map[string]*Entry)
	}
	return c
}

func NewEntry(hash string) *Entry {
	c := &Entry{
		Hash:      hash,
		Timestamp: time.Now(),
	}
	return c
}

func (c *Counter) GetSection(username string, repository string) *Section {
	sectionKey := username + "/" + repository
	c.mx.Lock()
	if _, ok := c.Sections[sectionKey]; !ok {
		c.Sections[sectionKey] = NewSection(username, repository)
	}
	c.mx.Unlock()
	return c.Sections[sectionKey]
}

func (c *Counter) GetSectionByKey(sectionKey string) *Section {
	if _, ok := c.Sections[sectionKey]; !ok {
		return nil
	}
	return c.Sections[sectionKey]
}

func (c *Counter) Run() {
	t := time.NewTicker(c.Duration)
	defer func() {
		t.Stop()
	}()

	for {
		select {
		case <-t.C:
			for sectionKey, section := range c.Sections {
				// Delete possible junk sections
				if section.Total == 1 && time.Now().After(section.CreatedAt.Add(24*time.Hour)) {
					c.RemoveSection(sectionKey)
				}else{
					_ = section.Save()
					if time.Now().After(section.UpdatedAt.Add(c.Duration)) {
						c.RemoveSection(sectionKey)
					}
				}
				for hash, entry := range section.Entries {
					if time.Now().After(entry.Timestamp.Add(c.Duration)) {
						c.RemoveEntry(section, hash)
					}
				}
			}
		}
	}
}

func (c *Counter) RemoveEntry(section *Section, hash string) {
	c.mx.Lock()
	if _, ok := section.Entries[hash]; !ok {
		delete(section.Entries, hash)
	}
	c.mx.Unlock()
}

func (c *Counter) RemoveSection(sectionKey string) {
	c.mx.Lock()
	if _, ok := c.Sections[sectionKey]; !ok {
		delete(c.Sections, sectionKey)
	}
	c.mx.Unlock()
}

func (c *Counter) AddEntry(section *Section, entry *Entry) bool {
	sectionKey := section.GetKey()

	c.mx.Lock()
	if _, ok := c.Sections[sectionKey]; !ok {
		c.Sections[sectionKey] = section
	}
	result := c.Sections[sectionKey].AddEntry(entry)
	c.mx.Unlock()

	return result
}