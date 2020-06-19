package formatter

import (
	"regexp"
)

type Formatter interface {
	Format(string) []string
}

type JsonData struct {
	Time         string
	User         string
	Host         string
	ConnectionId int64
	QueryTime    float64
	LockTime     float64
	RowsSent     int64
	RowsExamined int64
	DatabaseName string
	Timestamp    int64
	Query        string
}

func removeSensitiveData(data string) string {
	regexps := [...]string{
		// regex for credit cards, mobile
		"[0-9+]{10,21}",
		// regex for email, vpa
		"[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*",
	}

	for _, reg := range regexps {
		data = regexp.MustCompile(reg).ReplaceAllString(data, "*")
	}

	return data
}
