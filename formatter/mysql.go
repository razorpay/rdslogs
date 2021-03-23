package formatter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type MySQLFormatter struct {}

func (f *MySQLFormatter) Format(log string) []string {
	logSlice := strings.Split(log, "\n")

	dbName := ""
	counter := 0
	data := JsonData{}
	var QueryStrings []string

	for _, line := range logSlice {
		if counter == 0 {
			data = JsonData{}
		}

		counter++

		line = strings.Trim(line, " ")

		if line == "" {
			continue
		}

		if strings.Contains(line, "# Time") {
			data.Time = getQueryTime(line)
		} else if strings.Contains(line, "# User@Host") {
			data.User = getUser(line)

			data.Host = getHost(line)

			data.ConnectionId = getConnectionId(line)
		} else if strings.Contains(line, "# Query_time") {
			data.QueryTime = getQueryTimes(line, "Query_time")

			data.LockTime = getQueryTimes(line, "Lock_time")

			data.RowsSent = getRowsCount(line, "Rows_sent")

			data.RowsExamined = getRowsCount(line, "Rows_examined")
		} else if strings.Index(strings.ToLower(line), "use ") == 0 {
			dbName = getDatabaseName(line)
			data.DatabaseName = dbName
		} else if strings.Index(strings.ToLower(line), "set timestamp") == 0 {
			data.Timestamp = getTimestamp(line)
		} else {
			data.Query = removeSensitiveData(line)

			if data.Time == "" {
				continue
			}

			if data.DatabaseName == "" && dbName != "" {
				data.DatabaseName = dbName
			}

			jsonData, err := json.Marshal(data)

			if err != nil {
				continue
			}

			QueryStrings = append(QueryStrings, string(jsonData))

			counter = 0
		}
	}

	return QueryStrings
}

func getQueryTime(str string) string {
	regex := "([0-9]{4})[-]([0-9]{2})[-]([0-9]{2})T([0-9]{2})[:]([0-9]{2})[:]([0-9]{2})[./]([0-9]{6})Z$"
	match := regexp.MustCompile(regex).FindStringSubmatch(str)

	if len(match) > 0 {
		return match[0]
	}
	return ""
}

func getUser(str string) string {
	regex := "Host: ([a-z0-9_-]*)"
	match := regexp.MustCompile(regex).FindStringSubmatch(str)

	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func getHost(str string) string {
	regex := "(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}"
	match := regexp.MustCompile(regex).FindStringSubmatch(str)

	if len(match) > 0 {
		return match[0]
	}
	return ""
}

func getConnectionId(str string) int64 {
	regex := "Id: ([0-9]*)"
	match := regexp.MustCompile(regex).FindStringSubmatch(str)

	if len(match) > 1 {
		connectionId, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return 0
		}
		return connectionId
	}
	return 0
}

func getQueryTimes(str string, queryType string) float64 {
	regex := fmt.Sprintf("%s: (([0-9]*).([0-9]*))", queryType)

	match := regexp.MustCompile(regex).FindStringSubmatch(str)
	if len(match) > 1 {
		queryTime, err := strconv.ParseFloat(match[1], 64)

		if err != nil {
			return 0.0
		}
		return queryTime
	}
	return 0.0
}

func getRowsCount(str string, queryType string) int64 {
	regex := fmt.Sprintf("%s: (([0-9]*).([0-9]*))", queryType)

	match := regexp.MustCompile(regex).FindStringSubmatch(str)
	if len(match) > 1 {
		count, err := strconv.ParseInt(match[1], 10, 64)

		if err != nil {
			return 0
		}
		return count
	}
	return 0
}

func getDatabaseName(str string) string {
	str = strings.Replace(str, "use ", "", -1)
	str = strings.Replace(str, "USE ", "", -1)

	return strings.Trim(str, ";")
}

func getTimestamp(str string) int64 {
	regex := "([0-9]*)"
	match := regexp.MustCompile(regex).FindStringSubmatch(str)

	if len(match) > 0 {
		queryTime, err := strconv.ParseInt(match[0], 10, 64)

		if err != nil {
			return 0
		}
		return queryTime
	}
	return 0
}
