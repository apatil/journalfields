// This program reads journalctl json output from stdin
// Arguments are treated as field selectors.
// When they are given, only those fields will be shown
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-systemd/journal"

	"github.com/sirupsen/logrus"
)

// ignoredFields indicates the non _ fields that should not be printed
var ignoredFields = [...]string{
	"MESSAGE",
	"MESSAGE_ID",
	"PRIORITY",
	"SYSLOG_FACILITY",
	"SYSLOG_IDENTIFIER",
	"SYSLOG_PID",
}

// isIn returns true if `str` matches one of the strings in `strArr``
func isIn(str string, strArr []string) bool {
	for _, v := range strArr {
		if v == str {
			return true
		}
	}
	return false
}

// usToTime converts number of microseconds since epoch to a proper time.Time
func usToTime(timestampUs uint64) time.Time {
	tsUs := time.Microsecond * time.Duration(timestampUs)
	tsS := tsUs.Truncate(time.Second)
	tsUs -= tsS
	return time.Unix(int64(tsS.Seconds()), int64(tsUs.Nanoseconds()))
}

func main() {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	logfmt := logrus.TextFormatter{
		FullTimestamp: true,
	}

	// Search for help flag
	if isIn("-h", os.Args) || isIn("--help", os.Args) {
		fmt.Println("Usage: journalctl -o json | journalfields [SELECT_FIELD_TO_SHOW]...")
		fmt.Println("JournalFields is a simple program that reprints journalctl log entries with their hidden fields")
		os.Exit(0)
	}

	// Grab/parse selected fields to show from arguments
	var showFields []string
	if len(os.Args) > 1 {
		// Fields are uppercase only, so might as well just fix any given
		// lowecase fields
		showFields = make([]string, len(os.Args)-1)
		for i := 1; i < len(os.Args); i++ {
			showFields[i-1] = strings.ToUpper(os.Args[i])
		}
	}

	// Start processing log entries
	lines := bufio.NewScanner(os.Stdin)
	for lines.Scan() {
		var jFields map[string]interface{} // could be bytes
		line := lines.Bytes()
		if err := json.Unmarshal(line, &jFields); err != nil {
			log.Fatalf("Failed to unmarshal JSON (%s): %v", string(line), err)
		}
		// PRIORITY can actually be nil -- use default level
		if jFields["PRIORITY"] == nil {
			jFields["PRIORITY"] = "-1"
		}
		priority, err := strconv.Atoi(jFields["PRIORITY"].(string))
		if err != nil {
			log.Fatalf("Failed to parse priority: %v", err)
		}

		command := jFields["_COMM"]
		message := fmt.Sprint(jFields["MESSAGE"])

		timestampUsStr := jFields["__REALTIME_TIMESTAMP"].(string)
		timestampUs, err := strconv.ParseUint(timestampUsStr, 10, 64)
		if err != nil {
			log.Fatalf("Failed to parse timestamp to uint64: %v", err)
		}
		timestamp := usToTime(timestampUs)

		lFields := make(logrus.Fields)
		if showFields == nil {
			// Show default fields
			for k, v := range jFields {

				if len(k) > 0 && k[0] == '_' {
					continue
				}

				if isIn(k, ignoredFields[:]) {
					continue
				}

				lFields[k] = v
			}
		} else {
			// Only show select fields
			for _, k := range showFields {
				lFields[k] = jFields[k]
			}
		}

		// We need to link the logentry to a real Logger to allow checking
		// for a terminal and pretty printing
		logentry := log.WithFields(lFields)
		logentry.Time = timestamp
		logentry.Message = message

		/*
			From the journalhook library:
			logrus.DebugLevel: journal.PriDebug,
			logrus.InfoLevel:  journal.PriInfo,
			logrus.WarnLevel:  journal.PriWarning,
			logrus.ErrorLevel: journal.PriErr,
			logrus.FatalLevel: journal.PriCrit,
			logrus.PanicLevel: journal.PriEmerg,
		*/
		switch journal.Priority(priority) {
		case journal.PriDebug:
			logentry.Level = logrus.DebugLevel
		case journal.PriInfo:
			logentry.Level = logrus.InfoLevel
		case journal.PriWarning:
			logentry.Level = logrus.WarnLevel
		case journal.PriErr:
			logentry.Level = logrus.ErrorLevel
		case journal.PriCrit:
			logentry.Level = logrus.FatalLevel
		case journal.PriEmerg:
			logentry.Level = logrus.PanicLevel
		// default is for the items that do not have PRIORITY fields
		default:
			logentry.Level = logrus.InfoLevel
		}

		fmtBuf, err := logfmt.Format(logentry)
		if err != nil {
			log.Fatalf("Failed to format logentry: %v", err)
		}

		buf := []byte(fmt.Sprint(command, " "))
		buf = append(buf, fmtBuf...)

		if _, err := os.Stdout.Write(buf); err != nil {
			log.Fatalf("Failed to write formatted logentry: %v", err)
		}
	}

	if err := lines.Err(); err != nil {
		log.Fatalln("Error reading standard input:", err)
	}
}
