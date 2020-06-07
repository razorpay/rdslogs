package cli

import (
	"fmt"
	"strings"
	"time"
)

// LogFile wraps the returned structure from AWS
// "Size": 2196,
// "LogFileName": "slowquery/mysql-slowquery.log.7",
// "LastWritten": 1474959300000
type LogFile struct {
	Size            int64     `json:"Size"` // in bytes?
	LogFileName     string    `json:"LogFileName"`
	LastWritten     int64     `json:"LastWritten"` // arrives as msec since epoch
	LastWrittenTime time.Time `json:"LastWrittenTime"`
	Path            string    `json:"Path"`
}

func (l LogFile) String() string {
	return fmt.Sprintf("%-35s (date: %s, size: %d)", l.LogFileName, l.LastWrittenTime, l.Size)
}

//MatchFileWithMarker ....
func (l LogFile) MatchFileWithMarker(marker string) (string, bool) {
	splitMarker := strings.Split(marker, ":")
	splitFile := strings.Split(l.LogFileName, ".")

	if splitMarker[0] == splitFile[len(splitFile)-1] {
		return splitMarker[1], true
	}
	return "", false
}
