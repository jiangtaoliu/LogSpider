package logs

import (
	"bufio"
	"errors"
	"strings"
	"time"

	"github.com/bahusvel/NetworkScannerThingy/nstssh"
)

var CurrentTime = time.Now()

func init() {
	go updateTime()
}

func updateTime() {
	for {
		CurrentTime = time.Now()
		time.Sleep(1)
	}
}

type LogEntry struct {
	Host  string
	Log   string
	Time  time.Time
	Entry string
}

func FindLogs(host string) ([]string, error) {
	cmd := nstssh.Command(host, "find", "/var/log", "-name", "*log")
	if cmd == nil {
		return []string{}, errors.New("Cannot establish connection")
	}
	data, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}
	logFiles := []string{}
	for _, line := range strings.Split(string(data), "\n") {
		if line != "" {
			logFiles = append(logFiles, line)
		}
	}
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
			eventChannel <- LogEntry{Host: host, Time: CurrentTime, Log: log, Entry: strings.Trim(line, "\n")}
		}
	}()
	return cmd.Start()
}
