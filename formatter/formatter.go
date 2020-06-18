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
	var regexps []string

	// regex for email, vpa
	regexps = append(regexps, "[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*")

	// regex for credit cards, mobile
	regexps = append(regexps, `(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\\\/]?)?((?:\(?\d{1,}\)?[\-\.\\\/]?){0,})(?:[\-\.\\\/]?(?:#|ext\.?|extension|x)[\-\.\\\/]?(\d+))?`)

	for _, reg := range regexps {
		m1 := regexp.MustCompile(reg)
		data = m1.ReplaceAllString(data, "")
	}

	return data
}
