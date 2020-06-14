package formatter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type MySQLFormatter struct {}

func (f *MySQLFormatter) Format(log string) string {
	data := JsonData{}
	logSlice := strings.Split(log, "\n")

	for _, line := range logSlice {
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
			data.DatabaseName = getDatabaseName(line)
		} else if strings.Index(strings.ToLower(line), "set timestamp") == 0 {
			data.Timestamp = getTimestamp(line)
		} else {
			data.Query = line
		}
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		return ""
	}

	return string(jsonData)
}

func getQueryTime(str string) string {
	regex := "([0-9]{4})[-]([0-9]{2})[-]([0-9]{2})T([0-9]{2})[:]([0-9]{2})[:]([0-9]{2})[./]([0-9]{6})Z$"

	return regexp.MustCompile(regex).FindStringSubmatch(str)[0]
}

func getUser(str string) string {
	regex := "Host: ([a-z0-9_-]*)"

	return regexp.MustCompile(regex).FindStringSubmatch(str)[1]
}

func getHost(str string) string {
	re := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)

	return re.FindStringSubmatch(str)[0]
}

func getConnectionId(str string) int64 {
	regex := "Id: ([0-9]*)"

	id := regexp.MustCompile(regex).FindStringSubmatch(str)[1]

	connectionId, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		return 0
	}

	return connectionId
}

func getQueryTimes(str string, queryType string) float64 {
	regex := fmt.Sprintf("%s: (([0-9]*).([0-9]*))", queryType)

	timePeriod := regexp.MustCompile(regex).FindStringSubmatch(str)[1]

	queryTime, err := strconv.ParseFloat(timePeriod, 64)

	if err != nil {
		return 0.0
	}

	return queryTime
}

func getRowsCount(str string, queryType string) int64 {
	regex := fmt.Sprintf("%s: (([0-9]*).([0-9]*))", queryType)

	rowCount := regexp.MustCompile(regex).FindStringSubmatch(str)[1]

	count, err := strconv.ParseInt(rowCount, 10, 64)

	if err != nil {
		return 0
	}

	return count
}

func getDatabaseName(str string) string {
	str = strings.Replace(str, "use ", "", -1)
	str = strings.Replace(str, "USE ", "", -1)

	return strings.Trim(str, ";")
}

func getTimestamp(str string) int64 {
	regex := "([0-9]*)"

	timestamp := regexp.MustCompile(regex).FindStringSubmatch(str)[0]

	queryTime, err := strconv.ParseInt(timestamp, 10, 64)

	if err != nil {
		return 0
	}

	return queryTime
}
