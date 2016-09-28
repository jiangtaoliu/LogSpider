package logs

import (
	"bufio"
	"strings"

	"github.com/bahusvel/NetworkScannerThingy/nstssh"
)

func FindLogs(host string) ([]string, error) {
	data, err := nstssh.Command(host, "find", "/var/log", "-name", "*.log").Output()
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

func WatchLog(host string, log string, eventChannel chan string) error {
	cmd := nstssh.Command(host, "tail", "-f", log)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	go func() {
		reader := bufio.NewReader(out)
		line, err := reader.ReadString('\n')
		for ; err == nil; line, err = reader.ReadString('\n') {
			eventChannel <- strings.Trim(line, "\n")
		}
	}()
	return cmd.Start()
}
