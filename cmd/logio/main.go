package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/urfave/cli"

	"gopkg.in/mcuadros/go-syslog.v2"
)

var (
	conn       net.Conn = nil
	Server              = "172.16.1.109:28777"
	Quiet               = false
	ListenAddr          = "0.0.0.0:514"
)

type LogLine struct {
	Line string
	File string
	Node string
}

func sendTCPMessage(message string) error {
	if conn == nil {
		var err error
		conn, err = net.Dial("tcp", Server)
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
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "server, s",
			Value:       "0.0.0.0:514",
			Destination: &ListenAddr,
		},
		cli.StringFlag{
			Name:        "logio, l",
			Value:       "localhost:28777",
			Destination: &Server,
		},
		cli.BoolFlag{
			Name:        "queit, q",
			Destination: &Quiet,
		},
	}

	app.Action = func(c *cli.Context) error {
		channel := make(syslog.LogPartsChannel)
		handler := syslog.NewChannelHandler(channel)

		server := syslog.NewServer()
		server.SetFormat(syslog.RFC3164)
		server.SetHandler(handler)
		err := server.ListenTCP(ListenAddr)
		if err != nil {
			log.Fatal(err)
		}
		err = server.Boot()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("ListenAddr", ListenAddr)

		go func(channel syslog.LogPartsChannel) {
			defer func() {
				if n := recover(); n != nil {
					log.Println("Exception", n)
				}
			}()
			log.Println("channel", channel)
			for logParts := range channel {
				//streams := strings.Fields(logParts["content"].(string))[2]
				//streamstr := strings.Replace(streams, "[0]:", "", -1)
				if logParts["hostname"].(string) != "monitoringvm" {
					err := SendLog(&LogLine{logParts["content"].(string), logParts["tag"].(string), logParts["hostname"].(string)})
					if err != nil {
						log.Fatal(err)
					}
					if !Quiet {
						fmt.Println(logParts)
					}
				}
			}
		}(channel)

		server.Wait()

		return nil
	}

	app.Run(os.Args)

}
