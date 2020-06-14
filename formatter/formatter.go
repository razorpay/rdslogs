package formatter

type Formatter interface {
	Format(string) string
}

type JsonData struct {
	Time string
	User string
	Host string
	ConnectionId int64
	QueryTime float64
	LockTime float64
	RowsSent int64
	RowsExamined int64
	DatabaseName string
	Timestamp int64
	Query string
}
