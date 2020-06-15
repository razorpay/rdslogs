package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	flag "github.com/jessevdk/go-flags"
	"github.com/razorpay/rdslogs/cli"
	"github.com/razorpay/rdslogs/config"
	"github.com/razorpay/rdslogs/constants"
	"github.com/razorpay/rdslogs/tracker"
	log "github.com/sirupsen/logrus"
)

// BuildID is set by Travis CI
var BuildID string

func main() {
	appEnv := os.Getenv("APP_ENV")
	if len(appEnv) > 0 && (appEnv == "dev" || appEnv == "development") {
		log.Infof("Running application in development mode...")
		config.LoadConfigFromFile()
	}
	config.InitilizeLogging()

	options, err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	sigs := make(chan os.Signal, 1)
	abort := make(chan bool, 0)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Fprintf(os.Stderr, "Aborting! Caught Signal \"%s\"\n", sig)
		fmt.Fprintf(os.Stderr, "Cleaning up...\n")
		select {
		case abort <- true:
			close(abort)
		case <-time.After(10 * time.Second):
			fmt.Fprintf(os.Stderr, "Taking too long... Aborting.\n")
			os.Exit(1)
		}
	}()

	c := &cli.CLI{
		Options: options,
		RDS: rds.New(session.New(), &aws.Config{
			Region: aws.String(options.Region),
		}),
		Abort: abort,
	}

	// Loading config based on tracker
	if options.Tracker {
		if options.TrackerType == constants.TrackerRedis {
			config.InitilizeRedisConfig()
			c.Tracker = &tracker.RedisTracker{
				Pool: tracker.NewPool(),
			}
		}
	}

	if options.Debug {
		log.SetLevel(log.DebugLevel)
	}

	if options.Output == constants.OutputStdOut {
		fmt.Fprintln(os.Stderr, "Sending output to STDOUT")
	} else if options.Output == constants.OutputFile {
		fmt.Fprintln(os.Stderr, "Sending output to FILE")
	} else {
		log.Fatal("output target not recognized. use --help for usage info")
	}

	// make sure we can talk to an RDS instance.
	err = c.ValidateRDSInstance()
	if err == credentials.ErrNoValidProvidersFoundInChain {
		log.Fatal(awsCredsFailureMsg())
	}
	if err != nil {
		log.Fatal(err)
	}

	if options.Download {
		fmt.Fprintln(os.Stderr, "Running in download mode - downloading old logs")
		err = c.Download()
	} else {
		fmt.Fprintln(os.Stderr, "Running in tail mode - streaming logs from RDS")
		err = c.Stream()
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(os.Stderr, "OK")
}

// getVersion returns the internal version ID
func getVersion() string {
	if BuildID == "" {
		return "dev"
	}
	return fmt.Sprintf("%s", BuildID)
}

// parse all the flags, exit if anything's amiss
func parseFlags() (*config.Options, error) {
	var options config.Options
	flagParser := flag.NewParser(&options, flag.Default)
	flagParser.Usage = cli.Usage

	// parse flags and check for extra command line args
	if extraArgs, err := flagParser.Parse(); err != nil || len(extraArgs) != 0 {
		if err != nil {
			if err.(*flag.Error).Type == flag.ErrHelp {
				// user specified --help
				os.Exit(0)
			}
			fmt.Fprintln(os.Stderr, "Failed to parse the command line. Run with --help for more info")
			return nil, err
		}
		return nil, fmt.Errorf("Unexpected extra arguments: %s\n", strings.Join(extraArgs, " "))
	}

	// if all we want is the config file, just write it in and exit
	if options.WriteDefaultConfig {
		ip := flag.NewIniParser(flagParser)
		ip.Write(os.Stdout, flag.IniIncludeDefaults|flag.IniCommentDefaults|flag.IniIncludeComments)
		os.Exit(0)
	}

	// spit out the version if asked
	if options.Version {
		fmt.Println("Version:", getVersion())
		os.Exit(0)
	}
	// read the config file if specified
	if options.ConfigFile != "" {
		ini := flag.NewIniParser(flagParser)
		ini.ParseAsDefaults = true
		if err := ini.ParseFile(options.ConfigFile); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("config file %s doesn't exist", options.ConfigFile)
			}
			return nil, err
		}
	}

	if options.LogFile == "" {
		if options.DBType == constants.DBTypeMySQL {
			options.LogFile = "slowquery/mysql-slowquery.log"
		} else if options.DBType == constants.DBTypePostgreSQL {
			options.LogFile = "error/postgresql.log"
		} else {
			log.Fatal(fmt.Sprintf("unsupported dbtype: `%s`", options.DBType))
		}
	}

	return &options, nil
}

func awsCredsFailureMsg() string {
	// check for AWS binary
	_, err := exec.LookPath("aws")
	if err == nil {
		return `Unable to locate credentials. You can configure credentials by running "aws configure".`
	}
	return `Unable to locate AWS credentials. You have a few options:
- Create an IAM role for the host machine with the permissions to access RDS
- Use an AWS shared config file (~/.aws/config)
- Configure credentials on a development machine (via ~/.aws/credentials)
- Or set the AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables

You can read more at this security blog post:
http://blogs.aws.amazon.com/security/post/Tx3D6U6WSFGOK2H/A-New-and-Standardized-Way-to-Manage-Credentials-in-the-AWS-SDKs

Or read more about IAM roles and RDS at:
http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.IAM.AccessControl.IdentityBased.html`
}
