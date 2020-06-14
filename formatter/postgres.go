package formatter

import (
	"fmt"
)

type PostgresFormatter struct {}

func (f *PostgresFormatter) Format(log string) []string {
	return fmt.Sprintf("DATA: %s", log)
}
