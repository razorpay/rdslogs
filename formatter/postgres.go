package formatter

import (
	"fmt"
)

type PostgresFormatter struct {}

func (f *PostgresFormatter) Format(log string) []string {
	var str []string
	str = append(str, fmt.Sprintf("DATA: %s", log))

	return str
}
