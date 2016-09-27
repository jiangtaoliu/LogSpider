package logs

import (
	"strings"

	"github.com/bahusvel/NetworkScannerThingy/nstssh"
)

func FindLogs(host string) ([]string, error) {
	data, err := nstssh.Command(host, "find", "/var", "-name", "*.log").Output()
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

func WatchLog(host string, log string, eventChannel string) error {
	return nil
}
