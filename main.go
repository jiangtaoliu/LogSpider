package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	//"github.com/bahusvel/LogSpider/analyse"
	"github.com/bahusvel/LogSpider/logs"
	"github.com/bahusvel/LogSpider/nstssh"
	"github.com/urfave/cli"
)

const HOST_TIMEOUT = 30 * time.Second

const (
	OS_LINUX   = "linux"
	OS_WINDOWS = "windows"
	OS_DARWIN  = "darwin"
)

const (
	STATUS_UP   = "up"
	STATUS_DOWN = "down"
)

type Host struct {
	IPAddress  string
	OS         string
	Status     string
	StatusTime time.Time
	SSHEnabled bool
}

type ConnectionError struct {
	Host  string
	Error error
}

var mapMutex = sync.RWMutex{}
var hostMap = map[string]*Host{}
var scannerTimer *time.Timer
var logChannel = make(chan logs.LogEntry)
var connectionChannel = make(chan ConnectionError)

func SpiderHost(host string) {
	time.Sleep(1 * time.Second)
	cmd := nstssh.Command(host, "echo", "1")
	if cmd == nil {
		log.Println("Cannot establish ssh connectivity to", host, errors.New("cmd is nil"))
		connectionChannel <- ConnectionError{host, errors.New("cmd is nil")}
		return
	}

	err := cmd.Run()

	if err != nil {
		log.Println("Cannot establish ssh connectivity to", host, err)
		connectionChannel <- ConnectionError{host, err}
		return
	}

	hostLogs, err := logs.FindLogs(host)
	if err != nil {
		log.Println("Cannot fetch logs from", host, err)
		connectionChannel <- ConnectionError{host, err}
		return
	}

	log.Println("Found logs", hostLogs)

	for _, hostLog := range hostLogs {
		err := logs.WatchLog(host, hostLog, logChannel)
		if err != nil {
			log.Printf("Cannot open log %s at %s %s\n", host, hostLog, err)
		}
	}

	for {
		err := nstssh.Command(host, "echo", "1").Run()

		if err != nil {
			log.Println("Cannot establish ssh connectivity to", host, err)
			connectionChannel <- ConnectionError{host, err}
			return
		}
	}
}

func incrementIP(ip net.IP) net.IP {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
	return ip
}

func IPRange(iprange string) ([]string, error) {
	parts := strings.Split(iprange, "-")
	if len(parts) != 2 {
		return []string{}, errors.New("Invalid IP Range Format")
	}
	end := net.ParseIP(parts[1])
	ips := []string{}
	for ip := net.ParseIP(parts[0]); bytes.Compare(ip, end) <= 0; ip = incrementIP(ip) {
		ips = append(ips, ip.String())
	}
	return ips, nil
}

func main() {
	nstssh.Init("./id_rsa")
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name: "ipranges, i",
		},
		cli.StringFlag{
			Name: "output, o",
		},
	}
	//knowledgeBase := make(analyse.KnowledgeDB)
	app.Action = func(c *cli.Context) error {

		if len(c.StringSlice("ipranges")) == 0 {
			return cli.NewExitError("You did not specify any ranges", -1)
		}

		if c.String("output") == "" {
			return cli.NewExitError("You did not specify output file", -1)
		}

		file, err := os.Create(c.String("output"))
		if err != nil {
			return err
		}

		ips := []string{}
		for _, iprange := range c.StringSlice("ipranges") {
			ipsInRange, err := IPRange(iprange)
			if err != nil {
				return err
			}
			ips = append(ips, ipsInRange...)
		}

		for _, ip := range ips {
			log.Println("Spidering", ip)
			go SpiderHost(ip)
		}

		for {
			select {
			case connectionError := <-connectionChannel:
				go SpiderHost(connectionError.Host)
			case logEntry := <-logChannel:
				//_, isNew := knowledgeBase.Classify(logEntry)
				//if isNew {
				file.WriteString(fmt.Sprintf("%s,%s,%s\n", logEntry.Host, logEntry.Log, logEntry.Entry))
				//}
			}

		}
		//return nil
	}

	app.Run(os.Args)
}
