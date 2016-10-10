package vis

import (
	"testing"
	"time"

	"github.com/bahusvel/LogSpider/logs"
)

func TestLog(t *testing.T) {
	client := NewClient("")
	client.Log(logs.LogEntry{Host: "192.168.1.1", Time: time.Now(), Log: "/var/log/test.log", Entry: "Hello"})
	err := client.Flush("192.168.1.1")
	if err != nil {
		t.Error(err)
	}
	client.Close()
}
