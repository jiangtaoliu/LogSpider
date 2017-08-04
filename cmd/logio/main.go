package main

import (
	"fmt"
	"log"
	"net"

	"gopkg.in/mcuadros/go-syslog.v2"
)

var (
	conn     net.Conn = nil
	Protocol          = "tcp"
	Server            = "172.16.1.109:28777"
)

type LogLine struct {
	Line string
	File string
	Node string
}

func sendTCPMessage(message string) error {
	if conn == nil {
		var err error
		conn, err = net.Dial(Protocol, Server)
		if err != nil {
			return err
		}
	}
	_, err := conn.Write([]byte(message))
	if err != nil {
		conn = nil
	}
	return err
}

func SendLog(line *LogLine) error {
	err := sendTCPMessage(fmt.Sprintf("+log|%s|%s|info|%s\r\n", line.File, line.Node, line.Line))
	return err
}

func main() {
	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.RFC5424)
	server.SetHandler(handler)
	server.ListenTCP("0.0.0.0:514")
	server.Boot()

	go func(channel syslog.LogPartsChannel) {
		for logParts := range channel {
			err := SendLog(&LogLine{logParts["message"].(string), logParts["msg_id"].(string), logParts["hostname"].(string)})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(logParts)
		}
	}(channel)

	server.Wait()

	log.Println("Success")

}
