package formatter

import (
	"regexp"
	"strings"

	"github.com/razorpay/rdslogs/constants"
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
	regexps := map[string]string{
		constants.RegexEmail:     constants.RegexDefaultReplace,
		constants.RegexIpAddress: constants.RegexDefaultReplace,
		constants.RegexMobile:    constants.RegexDefaultReplace,
		constants.RegexName:      constants.RegexNameReplace,
	}

	data = strings.Replace(data, "`", "", -1)

	for reg, repl := range regexps {
		data = regexp.MustCompile(reg).ReplaceAllString(data, repl)
	}

	return data
}
