package publisher

import (
	"io"
	"os"
)

// STDOUTPublisher implements Publisher and sends the data to stdout
type STDOUTPublisher struct {
}

func (s *STDOUTPublisher) Write(line string) {
	_, _ = io.WriteString(os.Stdout, line)
}

func (s *STDOUTPublisher) Close() {}
