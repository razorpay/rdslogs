package publisher

import (
	"io"
	"os"
	"time"

	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/parsers"
)

// STDOUTPublisher implements Publisher and sends the data to stdout
type STDOUTPublisher struct {
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

func (s *STDOUTPublisher) Write(line string) {
	_, _ = io.WriteString(os.Stdout, line)
}

func (s *STDOUTPublisher) Close() {}
