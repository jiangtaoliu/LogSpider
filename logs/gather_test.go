package logs

import (
	"fmt"
	"testing"
)

func TestWatchLog(t *testing.T) {
	out := make(chan string)
	err := WatchLog("localhost", "/var/log/syslog", out)
	if err != nil {
		t.Fatal(err)
	}
	for line := range out {
		fmt.Println(line)
	}
}
