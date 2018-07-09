package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"

	"github.com/coreos/go-systemd/journal"

	"github.com/sirupsen/logrus"
)

var (
	fieldsRemoved = [...]string{
		"MESSAGE",
		"PRIORITY",
		"SYSLOG_FACILITY",
		"SYSLOG_IDENTIFIER",
		// "_SYSTEMD_UNIT",
	}
)

func isIn(str string, strArr []string) bool {
	for _, v := range strArr {
		if v == str {
			return true
		}
	}
	return false
}

func main() {
	log := logrus.New()

	lines := bufio.NewScanner(os.Stdin)
	for lines.Scan() {
		var jFields map[string]string
		if err := json.Unmarshal(lines.Bytes(), &jFields); err != nil {
			log.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		priority, err := strconv.Atoi(jFields["PRIORITY"])
		if err != nil {
			log.Fatalf("Failed to parse priority: %v", err)
		}
		message := jFields["MESSAGE"]
		// unit := jFields["_SYSTEMD_UNIT"]

		lFields := make(logrus.Fields)
		for k, v := range jFields {

			if len(k) > 0 && k[0] == '_' {
				continue
			}

			if isIn(k, fieldsRemoved[:]) {
				continue
			}

			lFields[k] = v
		}

		/*
			logrus.DebugLevel: journal.PriDebug,
			logrus.InfoLevel:  journal.PriInfo,
			logrus.WarnLevel:  journal.PriWarning,
			logrus.ErrorLevel: journal.PriErr,
			logrus.FatalLevel: journal.PriCrit,
			logrus.PanicLevel: journal.PriEmerg,
		*/

		logentry := log.WithFields(lFields)
		switch journal.Priority(priority) {
		case journal.PriDebug:
			logentry.Debug(message)
		case journal.PriInfo:
			logentry.Info(message)
		case journal.PriWarning:
			logentry.Warn(message)
		case journal.PriErr:
			logentry.Error(message)
		case journal.PriCrit:
			// We need a way to print Fatal and Panic
			// messages without killing the program
			// logentry.Fatal(message)
			logentry.Error(message)
		case journal.PriEmerg:
			// logentry.Panic(message)
			logentry.Error(message)
		default:
			logentry.Print(message)
		}

	}

	if err := lines.Err(); err != nil {
		log.Fatalln("Error reading standard input:", err)
	}
}
