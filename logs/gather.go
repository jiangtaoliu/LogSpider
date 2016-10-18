package logs

import (
	"bufio"
	"errors"
	"strings"
	"time"

	"github.com/bahusvel/NetworkScannerThingy/nstssh"
)

var CurrentTime = time.Now()
var excludeList = []string{"/var/lib/ceph"}

func init() {
	go updateTime()
}

func updateTime() {
	for {
		CurrentTime = time.Now()
		time.Sleep(1 * time.Second)
	}
}

type LogEntry struct {
	Host  string
	Log   string
	Time  time.Time
	Entry string
}

func isExclude(log string) bool {
	for _, exclude := range excludeList {
		if strings.HasPrefix(log, exclude) {
			return true
		}
	}
	return false
}

func FindLogs(host string) ([]string, error) {
	cmd := nstssh.Command(host, "find", "/var", "-name", "*.log")
	if cmd == nil {
		return []string{}, errors.New("Cannot establish connection")
	}
	data, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}
	logFiles := []string{}
	for _, line := range strings.Split(string(data), "\n") {
		if line != "" && !isExclude(line) {
			logFiles = append(logFiles, line)
		}
	}
	logFiles = append(logFiles, "/var/log/syslog")
	return logFiles, nil
}

func WatchLog(host string, log string, eventChannel chan LogEntry) error {
	cmd := nstssh.Command(host, "tail", "-f", log)
	if cmd == nil {
		return errors.New("Cannot establish connection")
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	go func() {
		reader := bufio.NewReader(out)
		line, err := reader.ReadString('\n')
		for ; err == nil; line, err = reader.ReadString('\n') {
			line = strings.Trim(line, "\n")
			if line != "" {
				eventChannel <- LogEntry{Host: host, Time: CurrentTime, Log: log, Entry: strings.Trim(line, "\n")}
			}
		}
	}()
	return cmd.Start()
}
