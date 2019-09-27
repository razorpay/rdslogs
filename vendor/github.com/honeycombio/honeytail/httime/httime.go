package httime

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	StrftimeChar        = "%"
	UnixTimestampFmt    = "%s(%L)?"
	UnixTimestampFmtAlt = "%s(.%L)?"
	UnixTimestampFmtTxt = "unixepoch"
)

// TODO UnixTimestampFmt is not a regex - it shouldn't look like one. Also even
// if it were a regex it would be wrong because it's expecting a dot between the
// seconds and fractional seconds, eg `%s(.%L)?`. But just changing it would
// break anybody that is using the magical  string `%s(%L)?` as their timestamp
// in order to enable this magical behavior. Is anybody doing that? Who can
// know? Introducing UnixTimestampFmtAlt gets it closer to what it's actually
// representing, except that it's not actually _just_ milliseconds, it's
// actually any fraction. Which can't really be represented using either
// strftime or go's time format. Ugh.  Time!! Everything is terrible. Maybe this
// magical string should just be "UnixEpochDecimal" instead to make clear it's
// something special?

var (
	// DefaultNower returns current time when called with Now() unless overridden
	DefaultNower Nower = &RealNower{}
	// Location defaults to UTC unless overridden
	Location *time.Location = time.UTC

	warnedAboutTime        = false
	possibleTimeFieldNames = []string{
		"time", "Time",
		"timestamp", "Timestamp", "TimeStamp",
		"date", "Date",
		"datetime", "Datetime", "DateTime",
		"msec",
	}
	// reference: http://man7.org/linux/man-pages/man3/strftime.3.html
	convertMapping = map[string]string{
		"%a": "Mon",
		"%A": "Monday",
		"%b": "Jan",
		"%B": "January",
		"%c": "", // locale not supported
		"%C": "06",
		"%d": "02",
		"%D": "01/02/06",
		"%e": "_2",
		"%E": "", // modifiers not supported
		"%f": "999",
		"%F": "2006-01-02",
		"%G": "", // week-based year not supported
		"%g": "", // week-based year not supported
		"%h": "Jan",
		"%H": "15",
		"%I": "03",
		"%j": "",   // day of year not supported
		"%k": "15", // same case as %H but accepts leading space instead of 0
		"%l": "_3",
		"%L": "999", // milliseconds
		"%m": "01",
		"%M": "04",
		"%n": "\n",
		"%O": "", // modifiers not supported
		"%p": "PM",
		"%P": "pm",
		"%r": "03:04:05 PM",
		"%R": "15:04",
		"%S": "05",
		"%t": "\t",
		"%T": "15:04:05",
		"%u": "", // day of week not supported
		"%U": "", // week number of the current year not supported
		"%V": "", // ISO 8601 week number not supported
		"%w": "", // day of week not supported
		"%W": "", // day of week not supported
		"%x": "", // date-only not supported
		"%X": "", // date-only not supported
		"%y": "06",
		"%Y": "2006",
		"%z": "-0700",
		"%Z": "MST",
		"%+": "Mon Jan _2 15:04:05 MST 2006",
	}
)

type Nower interface {
	Now() time.Time
}

type RealNower struct{}

func (r *RealNower) Now() time.Time {
	return time.Now().UTC()
}

func Now() time.Time {
	return DefaultNower.Now()
}

// GetTimestamp looks through the event map for something that looks like a
// timestamp.
//
// It will guess at the key name or use the specified one if it is not an empty
// string.  If unable to parse the timestamp, it will return the current time.
// The time field will be deleted from the map if found.
func GetTimestamp(m map[string]interface{}, timeFieldName, timeFieldFormat string) time.Time {
	var (
		ts                        time.Time
		foundFieldName            string
		timeFoundImproperTypeMsg  = "Found time field but type is not string or int"
		timeFoundInvalidFormatMsg = "found time field but failed to parse using specified format"
		timeFieldNotFoundMsg      = "Couldn't find specified time field"
	)
	if timeFieldName != "" {
		if t, found := m[timeFieldName]; found {
			timeStr := ""
			switch v := t.(type) {
			case string:
				timeStr = v
			case int:
				timeStr = strconv.Itoa(v)
			case float64:
				timeStr = strconv.FormatFloat(v, 'f', -1, 64)
			case time.Time:
				// it's a time.Time struct - we can just return it
				return v
			default:
				warnAboutTime(timeFieldName, t, timeFoundImproperTypeMsg)
				ts = Now()
			}
			if timeStr != "" {
				ts = tryTimeFormats(timeStr, timeFieldFormat)
				if ts.IsZero() {
					warnAboutTime(timeFieldName, t, timeFoundInvalidFormatMsg)
					ts = Now()
				}
			}
		} else {
			warnAboutTime(timeFieldName, nil, timeFieldNotFoundMsg)
			ts = Now()
		}
		// we were told to look for a specific field;
		// let's return what we found instead of continuing to look.
		delete(m, timeFieldName)
		return ts
	}
	// go through all the possible fields that might have a timestamp
	// for the first one we find, if it's a string field, try and parse it
	// if we succeed, stop looking. Otherwise keep trying
	for _, timeField := range possibleTimeFieldNames {
		if t, found := m[timeField]; found {
			timeStr, found := t.(string)
			if found {
				foundFieldName = timeField
				ts = tryTimeFormats(timeStr, timeFieldFormat)
				if !ts.IsZero() {
					break
				}
				warnAboutTime(timeField, t, timeFoundInvalidFormatMsg)
			}
		}
	}
	if ts.IsZero() {
		ts = Now()
	}
	delete(m, foundFieldName)
	return ts
}

// Parse wraps time.ParseInLocation to use httime's Location from parsers
func Parse(format, timespec string) (time.Time, error) {
	return time.ParseInLocation(format, timespec, Location)
}

// convertTimeFormat tries to handle C-style time formats alongside Go's
// existing time.Parse behavior.
func convertTimeFormat(layout string) string {
	for format, conv := range convertMapping {
		layout = strings.Replace(layout, format, conv, -1)
	}
	return layout
}

func tryTimeFormats(t, intendedFormat string) time.Time {
	// golang can't parse times with decimal fractional seconds marked by a comma
	// hack it by just replacing all commas with periods and hope it works out.
	// https://github.com/golang/go/issues/6189
	t = strings.Replace(t, ",", ".", -1)
	if (intendedFormat == UnixTimestampFmt) ||
		(intendedFormat == UnixTimestampFmtAlt) ||
		(intendedFormat == UnixTimestampFmtTxt) {
		if unix, err := strconv.ParseInt(t, 0, 64); err == nil {
			return time.Unix(unix, 0)
		}
		// fractional seconds eg 12345678.890123
		if unix, err := strconv.ParseFloat(t, 64); err == nil {
			sec, dec := math.Modf(unix)
			return time.Unix(int64(sec), int64(dec*(1e9)))
		}
	}
	if intendedFormat != "" {
		format := strings.Replace(intendedFormat, ",", ".", -1)
		if strings.Contains(format, StrftimeChar) {
			if ts, err := Parse(convertTimeFormat(format), t); err == nil {
				return ts
			}
		}

		// Still try Go style, just in case
		if ts, err := Parse(format, t); err == nil {
			return ts
		}
	}

	var ts time.Time
	if tOther, err := Parse("2006-01-02 15:04:05.999999999 -0700 MST", t); err == nil {
		ts = tOther
	} else if tOther, err := Parse(time.RFC3339Nano, t); err == nil {
		ts = tOther
	} else if tOther, err := Parse(time.RubyDate, t); err == nil {
		ts = tOther
	} else if tOther, err := Parse(time.UnixDate, t); err == nil {
		ts = tOther
	} else {
		// the defaults didn't catch it, let's try a few other things
		// is it all numeric? then try unix epoch times
		epochInt, err := strconv.ParseInt(t, 0, 64)
		if err == nil {
			// it might be seconds or it might be milliseconds! Who can know!
			// 10-digit numbers are seconds, 13-digit milliseconds, 16 microseconds
			if len(t) == 10 {
				ts = time.Unix(epochInt, 0)
			} else if len(t) > 10 {
				// turn it into seconds and fractional seconds
				fractionalTime := t[:10] + "." + t[10:]
				// then chop it into the int part and the fractional part
				if epochFloat, err := strconv.ParseFloat(fractionalTime, 64); err == nil {
					sec, dec := math.Modf(epochFloat)
					ts = time.Unix(int64(sec), int64(dec*(1e9)))
				}

			}
		} else {
			epochFloat, err := strconv.ParseFloat(t, 64)
			if err == nil {
				sec, dec := math.Modf(epochFloat)
				ts = time.Unix(int64(sec), int64(dec*(1e9)))
			}
		}
	}
	return ts
}

func warnAboutTime(fieldName string, foundTimeVal interface{}, msg string) {
	if warnedAboutTime {
		return
	}
	logrus.WithField("time_field", fieldName).WithField("time_value", foundTimeVal).Warn(msg + "\n  Please refer to https://honeycomb.io/docs/json#timestamp-parsing")
	warnedAboutTime = true
}
