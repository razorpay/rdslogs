package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/razorpay/rdslogs/config"
	"github.com/razorpay/rdslogs/constants"
	"github.com/razorpay/rdslogs/formatter"
	"github.com/razorpay/rdslogs/publisher"
	"github.com/razorpay/rdslogs/tracker"
	"github.com/sirupsen/logrus"
)

// StreamPos represents a log file and marker combination
type StreamPos struct {
	logFile LogFile
	marker  string
}

//PreviousMarker ...
type PreviousMarker struct {
	LogFile LogFile
	Marker  string
}

// CLI contains handles to the provided Options + aws.RDS struct
type CLI struct {
	// Options is for command line options
	Options *config.Options
	// RDS is an initialized session connected to RDS
	RDS *rds.RDS
	// Abort carries a true message when we catch CTRL-C so we can clean up
	Abort chan bool

	// target to which to send output
	output publisher.Publisher
	// allow changing the time for tests
	fakeNower Nower

	PreviousMarker PreviousMarker `json:"PreviousMarker"`

	Tracker tracker.Tracker
}

// Stream polls the RDS log endpoint forever to effectively tail the logs and
// spits them out to either stdout or to file.
func (c *CLI) Stream() error {
	trackerEnabled := false
	var logFilePath string

	// Enabling Tracker
	if c.Options.Tracker {
		data := c.Tracker.ReadLatestMarker(c.Options.InstanceIdentifier)

		if data != "" {
			trackerEnabled = true
			_ = json.Unmarshal([]byte(data), &c.PreviousMarker)
		}
	}

	// make sure we have a valid log file from which to stream
	latestFile, err := c.GetLatestLogFile()
	logFilePath = c.CreateFilePath(latestFile)
	if err != nil {
		return err
	}

	// forever, download the most recent entries
	sPos := StreamPos{
		logFile: latestFile,
	}

	// create the chosen output publisher target
	if c.Options.Output == constants.OutputStdOut {
		c.output = &publisher.STDOUTPublisher{}
	} else if c.Options.Output == constants.OutputFile {
		c.output = &publisher.FILEPublisher{
			FileName: latestFile.LogFileName,
			Path:     &logFilePath,
			Suffix:   &sPos.marker,
		}
	}

	for {
		// check for signal triggered exit
		select {
		case <-c.Abort:
			return fmt.Errorf("signal triggered exit")
		default:
		}

		// get recent log entries
		resp, err := c.getRecentEntries(sPos)
		if err != nil {
			if strings.HasPrefix(err.Error(), "Throttling: Rate exceeded") {
				logrus.Warnf("AWS Rate limit hit; sleeping for %d seconds.\n", c.Options.BackoffTimer)
				c.waitFor(time.Duration(c.Options.BackoffTimer) * time.Second)
				continue
			}

			if strings.HasPrefix(err.Error(), "InvalidParameterValue: This file contains binary data") {
				logrus.Warnf("binary data at marker %s, skipping 1000 in marker position\n", sPos.marker)
				// skip over inaccessible data
				newMarker, err := sPos.Add(1000)
				if err != nil {
					return err
				}
				sPos.marker = newMarker
				c.PreviousMarker = PreviousMarker{
					LogFile: sPos.logFile,
					Marker:  sPos.marker,
				}
				c.updateTracker()
				continue
			}

			if strings.HasPrefix(err.Error(), "DBLogFileNotFoundFault") {
				logrus.WithError(err).
					Warn("log does not appear to exist (rotation ongoing?) - waiting and retrying")
				c.waitFor(time.Second * 5)
				continue
			}

			return err
		}

		if !*resp.AdditionalDataPending || (resp.Marker != nil && *resp.Marker == "0") {
			if c.Options.DBType == constants.DBTypePostgreSQL {
				// If that's all we've got for now, see if there's a newer file to
				// start tailing. This logic is only relevant for postgres: the
				// newest postgres log file will be named
				// error/postgresql.log.YYYY-MM-DD,
				// but the newest mysql log
				// will always be named
				// slowquery/mysql-slowquery.log.
				newestFile, err := c.GetLatestLogFile()
				logFilePath = c.CreateFilePath(newestFile)
				if err != nil {
					return err
				}
				if newestFile.LogFileName != sPos.logFile.LogFileName {
					logrus.WithFields(logrus.Fields{
						"oldFile": sPos.logFile.LogFileName,
						"newFile": newestFile.LogFileName}).Info("Found newer file")
					sPos.logFile = newestFile
					continue
				}
			}

			// Wait for a few seconds and try again.
			c.waitFor(5 * time.Second)
		}

		newMarker := c.getNextMarker(sPos, resp)

		if sPos.marker != newMarker {
			logrus.WithFields(logrus.Fields{
				"prevMarker": sPos.marker,
				"newMarker":  newMarker,
				"file":       sPos.logFile.LogFileName}).
				Info("Got new marker")
		}

		if newMarker == "0" {
			latestFile, err := c.GetLatestLogFile()
			logFilePath = c.CreateFilePath(latestFile)
			if err != nil {
				return err
			}
			sPos.logFile = latestFile
		}

		// In tracker is enabled, will download the previous file written less than an hour ago
		if trackerEnabled {
			splitMarker := strings.Split(c.PreviousMarker.Marker, ":")
			splitNewMarker := strings.Split(newMarker, ":")
			newMarkerInt, _ := strconv.Atoi(splitNewMarker[1])
			newMarkerInt = newMarkerInt - len(*resp.LogFileData)

			c1 := make(chan LogFile)
			flag := false
			if sPos.logFile.LastWritten-c.PreviousMarker.LogFile.LastWritten < 3600000 {
				if splitMarker[0] == splitNewMarker[0] {
					flag = true
					suffix := "." + splitNewMarker[0] + "." + splitMarker[1] + "-" + strconv.Itoa(newMarkerInt)
					sPos.logFile.Path = c.CreateFilePath(sPos.logFile, suffix)
					go c.downloadFile(sPos.logFile, c1, c.PreviousMarker.Marker, splitMarker[1], strconv.Itoa(newMarkerInt))
				}
			}
			trackerEnabled = false
			if flag {
				<-c1
			} else {
				suffix := "." + splitNewMarker[0] + ".0-" + strconv.Itoa(newMarkerInt)
				sPos.logFile.Path = c.CreateFilePath(sPos.logFile, suffix)
				go c.downloadFile(sPos.logFile, c1, "0", "0", strconv.Itoa(newMarkerInt))
				<-c1
			}

		}

		sPos.marker = newMarker
		c.PreviousMarker = PreviousMarker{
			LogFile: sPos.logFile,
			Marker:  sPos.marker,
		}
		c.updateTracker()

		// Writing data to Publisher
		if resp.LogFileData != nil && *resp.LogFileData != "" {
			formattedData := c.formatLogFileData(*resp.LogFileData)

			for _, jsonData := range formattedData {
				if jsonData != "" {
					c.output.Write(jsonData)
				}
			}
		}
	}
}

// getNextMarker takes in to account the current and next reported markers and
// decides whether to believe the resp.Marker or calculate its own next marker.
func (c *CLI) getNextMarker(sPos StreamPos, resp *rds.DownloadDBLogFilePortionOutput) string {
	// if resp is nil, we're up a creek and should return sPos' marker, but at
	// least we shouldn't try and dereference it and panic.
	if resp == nil {
		logrus.Warn("resp was nil, returning previous marker")
		return sPos.marker
	}

	if resp.Marker == nil {
		logrus.Warn("resp marker is nil, returning previous marker")
		return sPos.marker
	}

	// when we get to the end of a log segment, the marker in resp is "0".
	// if it's not "0", we should trust it's correct and use it.
	if *resp.Marker != "0" {
		return *resp.Marker
	}

	// ok, we've hit the end of a segment, but did we get any data? If we got
	// data, then it's not really the end of the segment and we should calculate a
	// new marker and use that.
	if resp.LogFileData != nil && len(*resp.LogFileData) != 0 {
		newMarkerStr, err := sPos.Add(len(*resp.LogFileData))
		if err != nil {
			logrus.WithError(err).
				Warn("failed to get next marker. Reverting to no marker.")
			return "0"
		}
		return newMarkerStr
	}
	// we hit the end of a segment but we didn't get any data. we should try again
	// during the 00-05 minutes past the hour time, and roll over once we get to 6
	// minutes past the hour
	var now time.Time
	if c.fakeNower != nil {
		now = c.fakeNower.Now().UTC()
	} else {
		now = time.Now().UTC()
	}
	curMin, _ := strconv.Atoi(now.Format("04"))
	if curMin > 5 {
		logrus.WithField("newMarker", *resp.Marker).
			Infof("no log data received but it's %d minutes (> 5) past "+
				"the hour, returning resp marker", curMin)
		return *resp.Marker
	}
	logrus.WithField("prevMarker", sPos.marker).
		Infof("no log data received but it's %d minutes (< 5) past "+
			"the hour, returning previous marker", curMin)
	// let's try again from where we did the last time.
	return sPos.marker
}

// Add returns a new marker string that is the current marker + dataLen offset
func (s *StreamPos) Add(dataLen int) (string, error) {
	splitMarker := strings.Split(s.marker, ":")
	if len(splitMarker) != 2 {
		// something's wrong. marker should have been #:#
		// TODO provide a better value
		return "", fmt.Errorf("marker didn't split into two pieces across a colon")
	}
	mHour, _ := strconv.Atoi(splitMarker[0])
	mOffset, _ := strconv.Atoi(splitMarker[1])
	mOffset += dataLen
	return fmt.Sprintf("%d:%d", mHour, mOffset), nil
}

// getRecentEntries fetches the most recent lines from the log file, starting
// from marker or the end of the file if marker is nil
// returns the downloaded data
func (c *CLI) getRecentEntries(sPos StreamPos) (*rds.DownloadDBLogFilePortionOutput, error) {
	params := &rds.DownloadDBLogFilePortionInput{
		DBInstanceIdentifier: aws.String(c.Options.InstanceIdentifier),
		LogFileName:          aws.String(sPos.logFile.LogFileName),
		NumberOfLines:        aws.Int64(c.Options.NumLines),
	}
	// if we have a marker, download from there. otherwise get the most recent line
	if sPos.marker != "" {
		params.Marker = &sPos.marker
	} else {
		params.NumberOfLines = aws.Int64(1)
	}
	return c.RDS.DownloadDBLogFilePortion(params)
}

// Download downloads RDS logs and reads them all in
func (c *CLI) Download() error {
	// get a list of RDS instances, return the one to use.
	// if one's user supplied, verify it exists.
	// if not user supplied and there's only one, use that
	// else ask
	logFiles, err := c.GetLogFiles()
	if err != nil {
		return err
	}

	logFiles, err = c.DownloadLogFiles(logFiles)
	if err != nil {
		logrus.Error("Error downloading log files:", err)
		return err
	}

	return nil
}

// DownloadLogFiles returns a new copy of the logFile list because it mutates the contents.
func (c *CLI) DownloadLogFiles(logFiles []LogFile) ([]LogFile, error) {
	logrus.Infof("Downloading log files to %s\n", c.Options.DownloadDir)
	downloadedLogFiles := make([]LogFile, 0, len(logFiles))
	for i := range logFiles {
		// returned logFile has a modified Path
		c1 := make(chan LogFile)
		go c.downloadFile(logFiles[i], c1)
		select {
		case logFile := <-c1:
			downloadedLogFiles = append(downloadedLogFiles, logFile)
		}
	}
	return downloadedLogFiles, nil
}

// downloadFile fetches an individual log file. Note that AWS's RDS
// DownloadDBLogFilePortion only returns 1MB at a time, and we have to manually
// paginate it ourselves.
func (c *CLI) downloadFile(logFile LogFile, ch chan LogFile, customPathOptional ...string) (LogFile, error) {
	logFileData := ""
	var err error
	var output publisher.Publisher
	params := &rds.DownloadDBLogFilePortionInput{
		DBInstanceIdentifier: aws.String(c.Options.InstanceIdentifier),
		LogFileName:          aws.String(logFile.LogFileName),
	}

	resp := &rds.DownloadDBLogFilePortionOutput{
		AdditionalDataPending: aws.Bool(true),
		Marker:                aws.String("0"),
	}

	if len(customPathOptional) < 1 {
		logFile.Path = path.Join(c.Options.DownloadDir, path.Base(logFile.LogFileName))
	} else {
		params.Marker = aws.String(customPathOptional[0])
		resp.Marker = aws.String(customPathOptional[0])
	}

	if c.Options.Output == constants.OutputStdOut {
		output = &publisher.STDOUTPublisher{}
	} else if c.Options.Output == constants.OutputFile {
		output = &publisher.FILEPublisher{
			FileName: logFile.LogFileName,
			Path:     &logFile.Path,
		}
	}

	if c.Options.Download {
		// open the out file for writing
		logrus.Infof("Downloading %s to %s ... ", logFile.LogFileName, logFile.Path)
		output = &publisher.FILEPublisher{
			FileName: logFile.LogFileName,
			Path:     &logFile.Path,
		}
	} else {
		logrus.Infof("Downloading previous file %s in %s mode", logFile.LogFileName, c.Options.Output)
	}
	defer logrus.Infof("done\n")

	for aws.BoolValue(resp.AdditionalDataPending) {
		// check for signal triggered exit
		select {
		case <-c.Abort:
			return logFile, fmt.Errorf("signal triggered exit")
		default:
		}

		params.Marker = resp.Marker // support pagination
		resp, err = c.RDS.DownloadDBLogFilePortion(params)
		if err != nil {
			return logFile, err
		}

		if len(customPathOptional) > 2 {
			endMarker, _ := strconv.Atoi(customPathOptional[2])
			startMarker, _ := strconv.Atoi(customPathOptional[1])
			logFileData = logFileData + aws.StringValue(resp.LogFileData)
			end := endMarker - startMarker
			if len(logFileData) >= end {
				formattedData := c.formatLogFileData(logFileData[0:end])

				for _, jsonData := range formattedData {
					if jsonData != "" {
						output.Write(jsonData)
					}
				}
				break
			}
		} else {
			formattedData := c.formatLogFileData(aws.StringValue(resp.LogFileData))

			for _, jsonData := range formattedData {
				if jsonData != "" {
					output.Write(jsonData)
				}
			}
		}
	}

	ch <- logFile
	logrus.Infof("file: %s is successfully downloaded", logFile.LogFileName)
	return logFile, nil
}

// GetLogFiles returns a list of all log files based on the Options.LogFile pattern
func (c *CLI) GetLogFiles() ([]LogFile, error) {
	// get a list of all log files.
	// prune the list so that the log file option is the prefix for all remaining files
	// return the list of as-yet unread files
	logFiles, err := c.getListRDSLogFiles()
	if err != nil {
		return nil, err
	}

	var matchingLogFiles []LogFile
	for _, lf := range logFiles {
		if strings.HasPrefix(lf.LogFileName, c.Options.LogFile) {
			matchingLogFiles = append(matchingLogFiles, lf)
		}
	}
	// matchingLogFiles now contains a list of eligible log files,
	// eg slow.log, slow.log.1, slow.log.2, etc.

	if len(matchingLogFiles) == 0 {
		errParts := []string{"No log file with the given prefix found. Available log files:"}

		for _, lf := range logFiles {
			errParts = append(errParts, fmt.Sprint("\t", lf.String()))
		}
		return nil, fmt.Errorf(strings.Join(errParts, "\n"))

	}

	return matchingLogFiles, nil
}

// GetLatestLogFile ...
func (c *CLI) GetLatestLogFile() (LogFile, error) {

	logFiles, err := c.GetLogFiles()

	if err != nil {
		return LogFile{}, err
	}

	if len(logFiles) == 0 {
		return LogFile{}, errors.New("No log files found")
	}

	sort.SliceStable(logFiles, func(i, j int) bool { return logFiles[i].LastWritten < logFiles[j].LastWritten })
	if c.PreviousMarker.LogFile.LastWritten > 0 && len(logFiles) > 1 {
		c.DownloadPreviousFiles(logFiles[:len(logFiles)-1])
	}
	return logFiles[len(logFiles)-1], nil
}

// Gets a list of all available RDS log files for an instance.
func (c *CLI) getListRDSLogFiles() ([]LogFile, error) {
	var output *rds.DescribeDBLogFilesOutput
	var err error
	var logFiles []LogFile
	for {
		if output == nil {
			params := &rds.DescribeDBLogFilesInput{
				DBInstanceIdentifier: &c.Options.InstanceIdentifier,
			}
			if c.PreviousMarker.LogFile.LastWritten > 0 {
				params.FileLastWritten = aws.Int64(c.PreviousMarker.LogFile.LastWritten)
			}
			output, err = c.RDS.DescribeDBLogFiles(params)
			logFiles = make([]LogFile, 0, len(output.DescribeDBLogFiles))
		} else {
			output, err = c.RDS.DescribeDBLogFiles(&rds.DescribeDBLogFilesInput{
				DBInstanceIdentifier: &c.Options.InstanceIdentifier,
				Marker:               output.Marker,
			})
		}
		if err != nil {
			return nil, err
		}

		// assign go timestamp from msec epoch time, rebuild as a list
		for _, lf := range output.DescribeDBLogFiles {
			logFiles = append(logFiles, LogFile{
				LastWritten:     *lf.LastWritten,
				LastWrittenTime: time.Unix(*lf.LastWritten/1000, 0),
				LogFileName:     *lf.LogFileName,
				Size:            *lf.Size,
			})
		}
		if output.Marker == nil {
			break
		}
	}
	return logFiles, nil
}

// ValidateRDSInstance validates that you have a valid RDS instance to talk to.
// If an instance isn't specified and your credentials contain more than one RDS
// instance, asks you to specify which instance you'd like to use.
func (c *CLI) ValidateRDSInstance() error {
	rdsInstances, err := c.getListRDSInstances()
	if err != nil {
		return err
	}

	if len(rdsInstances) == 0 {
		// we didn't get any instances back from RDS. not sure what to do next...
		return fmt.Errorf("The list of instances we got back from RDS is empty. Check the region and authentication?")
	}

	if c.Options.InstanceIdentifier != "" {
		for _, instance := range rdsInstances {
			if c.Options.InstanceIdentifier == instance {
				// the user asked for an instance and we found it in the list. \o/
				return nil
			}
		}
		// the user asked for an instance but we didn't find it.
		return fmt.Errorf("Instance identifier %s not found in list of instances:\n\t%s",
			c.Options.InstanceIdentifier,
			strings.Join(rdsInstances, "\n\t"))
	}

	// user didn't ask for an instance.
	// complain with a list of avaialable instances and exit.
	errStr := fmt.Sprintf(`No instance identifier specified. Available RDS instances:
	%s
Please specify an instance identifier using the --identifier flag
`, strings.Join(rdsInstances, "\n\t"))
	return fmt.Errorf(errStr)
}

// gets a list of all avaialable RDS instances
func (c *CLI) getListRDSInstances() ([]string, error) {
	out, err := c.RDS.DescribeDBInstances(nil)
	if err != nil {
		return nil, err
	}
	instances := make([]string, len(out.DBInstances))
	for i, instance := range out.DBInstances {
		instances[i] = *instance.DBInstanceIdentifier
	}
	return instances, nil
}

func (c *CLI) waitFor(d time.Duration) {
	select {
	case <-c.Abort:
		return
	case <-time.After(d):
		return
	}
}

// Nower interface abstracts time for testing
type Nower interface {
	Now() time.Time
}

//DownloadPreviousFiles ...
func (c *CLI) DownloadPreviousFiles(logFiles []LogFile) {
	for _, logFile := range logFiles {
		c1 := make(chan LogFile)
		logFile.Path = c.CreateFilePath(logFile)
		if size, ok := logFile.MatchFileWithMarker(c.PreviousMarker.Marker); ok {
			logFile.Path = logFile.Path + "." + size
			go c.downloadFile(logFile, c1, c.PreviousMarker.Marker)
		} else {
			go c.downloadFile(logFile, c1, "0")
		}
		<-c1
	}
}

//CreateFilePath ....
func (c *CLI) CreateFilePath(logFiles LogFile, suffix ...string) string {
	currentTime := time.Now()
	TimeFormat := currentTime.Format("01-02-2006")
	splitFile := strings.Split(logFiles.LogFileName, "/")
	if len(suffix) > 0 {
		splitFile[1] = splitFile[1] + suffix[0]
	}
	return strings.Join([]string{
		c.Options.DownloadDir,
		splitFile[0],
		c.Options.InstanceIdentifier,
		TimeFormat,
		splitFile[1],
	}, "/")
}

func (c *CLI) updateTracker() {
	if c.Options.Tracker {
		e, _ := json.Marshal(c.PreviousMarker)
		c.Tracker.WriteLatestMarker(c.Options.InstanceIdentifier, string(e))
	}
}

func (c *CLI) formatLogFileData(logFileData string) []string {
	var formattedData []string

	if c.Options.Formatter {
		if c.Options.DBType == constants.DBTypeMySQL {
			formatter := &formatter.MySQLFormatter{}

			formattedData = formatter.Format(logFileData)
		} else if c.Options.DBType == constants.DBTypePostgreSQL {
			formatter := &formatter.PostgresFormatter{}

			formattedData = formatter.Format(logFileData)
		}
	} else {
		formattedData = []string{logFileData}
	}

	return formattedData
}
