package publisher

import (
	"log"
	"os"
	"path"
	"strings"
)

// FILEPublisher implements Publisher and saves data to file
type FILEPublisher struct {
	FileName string
	Path     string
	Suffix   *string
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
