package output

import (
	"testing"

	"github.com/bahusvel/NetworkScannerThingy/logs"
)

func TestInsert(t *testing.T) {
	client := &ElasticOutput{ServerURL: "http://192.168.1.83:9200", IndexName: "logs"}
	err := client.Init()
	if err != nil {
		t.Error(err)
		return
	}
	entry := logs.LogEntry{Host: "test", Entry: "hello world"}
	err = client.SendLogEntry(entry)
	if err != nil {
		t.Error(err)
		return
	}
}
