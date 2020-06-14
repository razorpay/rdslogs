package publisher

import (
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/parsers"
)

// FILEPublisher implements Publisher and saves data to file
type FILEPublisher struct {
	FileName string
	Path     string
	Suffix   *string
	Writekey       string
	Dataset        string
	APIHost        string
	ScrubQuery     bool
	SampleRate     int
	Parser         parsers.Parser
	AddFields      map[string]string
	initialized    bool
	lines          chan string
	eventsToSend   chan event.Event
	eventsSent     uint
	lastUpdateTime time.Time
}

func (s *FILEPublisher) Write(line string) {
	suffix := ""
	if *s.Suffix != "" {
		splitMarker := strings.Split(*s.Suffix, ":")
		suffix = "." + splitMarker[0]
	}

	filename := s.Path + suffix
	if err := os.MkdirAll(path.Dir(filename), os.ModePerm); err != nil {
		log.Fatal(err)
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(line)); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func (s *FILEPublisher) Close() {}