package cli

// Usage info for --help
var Usage = `rdslogs --identifier my-rds-instance

rdslogs streams a log file from Amazon RDS and prints it to STDOUT or File

AWS credentials are required and can be provided via IAM roles, AWS shared
config (~/.aws/config), AWS shared credentials (~/.aws/credentials), or
the environment variables AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.

Passing --download triggers Download Mode, in which rdslogs will download the
specified logs to the directory specified by --download_dir. Logs are specified
via the --log_file flag, which names an active log file as well as the past 24
hours of rotated logs. (For example, specifying --log_file=foo.log will download
foo.log as well as foo.log.0, foo.log.2, ... foo.log.23.)

When --output is set to "file", will download the specified logs to the directory
specified by --download_dir instead of being printed to STDOUT.

When --tracker is enabled, it will store the marker by default to redis or we can
set the tracker type by passing value to --tracker_type. Tracker backfills the data
in stream mode only according to marker stored in tracker.
`
