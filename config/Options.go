package config

// Options contains all the CLI flags
type Options struct {
	Region             string `long:"region" description:"AWS region to use" default:"us-east-1"`
	InstanceIdentifier string `short:"i" long:"identifier" description:"RDS instance identifier"`
	DBType             string `long:"dbtype" description:"RDS database type. Accepted values are mysql and postgresql." default:"mysql"`
	LogFile            string `short:"f" long:"log_file" description:"RDS log file to retrieve"`
	Download           bool   `short:"d" long:"download" description:"Download old logs instead of tailing the current log"`
	DownloadDir        string `long:"download_dir" description:"directory in to which log files are downloaded" default:"./"`
	NumLines           int64  `long:"num_lines" description:"number of lines to request at a time from AWS. Larger number will be more efficient, smaller number will allow for longer lines" default:"10000"`
	BackoffTimer       int64  `long:"backoff_timer" description:"how many seconds to pause when rate limited by AWS." default:"5"`
	Output             string `short:"o" long:"output" description:"output for the logs: stdout or file" default:"stdout"`
	Formatter          bool   `long:"formatter" description:"To format the logs in json"`
	Tracker            bool   `long:"tracker" description:"To store the marker information"`
	TrackerType        string `long:"tracker_type" description:"To store the marker information to some database" default:"redis"`
	Version            bool   `short:"v" long:"version" description:"Output the current version and exit"`
	ConfigFile         string `short:"c" long:"config" description:"config file" no-ini:"true"`
	WriteDefaultConfig bool   `long:"write_default_config" description:"Write a default config file to STDOUT" no-ini:"true"`
	Debug              bool   `long:"debug" description:"turn on debugging output"`
}
