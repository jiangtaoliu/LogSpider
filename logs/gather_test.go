package logs

import (
	"fmt"
	"os"
	"testing"

	"github.com/bahusvel/NetworkScannerThingy/nstssh"
)

func TestWatchLog(t *testing.T) {
	nstssh.IDENTITY = os.Getenv("HOME") + "/.ssh/id_rsa"
	out := make(chan LogEntry)
	err := WatchLog("192.168.1.248", "/var/log/syslog", out)
	if err != nil {
		t.Fatal(err)
	}
	for line := range out {
		fmt.Println(line)
	}
}
