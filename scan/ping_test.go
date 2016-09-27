package scan

import (
	"log"
	"testing"
)

func TestPing(t *testing.T) {
	pingChan := make(chan string)
	Init(pingChan)
	go PingHosts([]string{"192.168.2.1"})
	log.Println("pinging")
	for {
		select {
		case host := <-pingChan:
			log.Println("Host is up", host)
		}
	}
}
