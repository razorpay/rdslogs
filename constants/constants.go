package constants

const (
	OutputStdOut = "stdout"

	OutputFile = "file"

	TrackerRedis = "redis"

	// Fortunately for us, the RDS team has diligently ignored requests to make
	// RDS Postgres's `log_line_prefix` customizable for years
	// (https://forums.aws.amazon.com/thread.jspa?threadID=143460).
	// So we can hard-code this prefix format for Postgres log lines.
	RdsPostgresLinePrefix = "%t:%r:%u@%d:[%p]:"

	// DBTypePostgreSQL postgresql db type
	DBTypePostgreSQL = "postgresql"

	// DBTypeMySQL mysql db type
	DBTypeMySQL = "mysql"
)
