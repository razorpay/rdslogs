package cli

// Usage info for --help
var Usage = `rdslogs --identifier my-rds-instance

rdslogs streams a log file from Amazon RDS and prints it to STDOUT or sends it
up to Honeycomb.io.

AWS credentials are required and can be provided via IAM roles, AWS shared
config (~/.aws/config), AWS shared credentials (~/.aws/credentials), or
the environment variables AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.

Passing --download triggers Download Mode, in which rdslogs will download the
specified logs to the directory specified by --download_dir. Logs are specified
via the --log_file flag, which names an active log file as well as the past 24
hours of rotated logs. (For example, specifying --log_file=foo.log will download
foo.log as well as foo.log.0, foo.log.2, ... foo.log.23.)

When --output is set to "honeycomb", the --writekey and --dataset flags are
required. Instead of being printed to STDOUT, database events from the log will
be transmitted to Honeycomb. --scrub_query and --sample_rate also only apply to
honeycomb output.
`
