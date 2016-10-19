package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	//"github.com/bahusvel/NetworkScannerThingy/analyse"
	"github.com/bahusvel/NetworkScannerThingy/logs"
	"github.com/bahusvel/NetworkScannerThingy/nstssh"
	"github.com/bahusvel/NetworkScannerThingy/scan"
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

var mapMutex = sync.RWMutex{}
var hostMap = map[string]*Host{}
var scannerTimer *time.Timer
var logChannel = make(chan logs.LogEntry)

func newHost(address string) (*Host, error) {
	log.Println("New Host", address)
	host := &Host{IPAddress: address, SSHEnabled: true}
	err := nstssh.CopyID("localhost", host.IPAddress, "cp-x2520")
	if err != nil {
		log.Println("Cannot establish ssh connectivity to", host)
		host.SSHEnabled = false
		return nil, err
	}
	err = nstssh.SetupMultiPlexing(host.IPAddress)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return host, nil
}

func hostAlive(host string) {
	mapMutex.RLock()
	existingHost, ok := hostMap[host]
	mapMutex.RUnlock()
	if !ok {
		var err error
		existingHost, err = newHost(host)
		if err != nil {
			log.Println(err)
			return
		}
		mapMutex.Lock()
		hostMap[host] = existingHost
		mapMutex.Unlock()
	}

	if existingHost.Status != STATUS_UP {
		start := time.Now()
		for {
			err := hostUp(existingHost)
			if err == nil {
				break
			}
			if time.Now().Sub(start) > HOST_TIMEOUT {
				log.Printf("Connection to %s timeout\n", existingHost.IPAddress)
				break
			}
		}
	}
	existingHost.Status = STATUS_UP
	existingHost.StatusTime = time.Now()
}

func hostUp(host *Host) error {
	log.Println("Host went up", host)

	if !host.SSHEnabled {
		return nil
	}

	hostLogs, err := logs.FindLogs(host.IPAddress)
	if err != nil {
		log.Println("Cannot fetch logs from", host)
		return err
	}

	for _, hostLog := range hostLogs {
		err := logs.WatchLog(host.IPAddress, hostLog, logChannel)
		if err != nil {
			log.Printf("Cannot open log %s at %s %s\n", host.IPAddress, hostLog, err)
		}
	}
	return nil
}

func hostDown(host *Host) {
	host.Status = STATUS_DOWN
	host.StatusTime = time.Now()
	log.Println("Host down", host)
}

func timeoutScanner() {
	timeout := time.Now().Add(-HOST_TIMEOUT)
	mapMutex.RLock()
	for _, host := range hostMap {
		if host.Status == STATUS_UP && host.StatusTime.Before(timeout) {
			hostDown(host)
		}
	}
	mapMutex.RUnlock()
	scannerTimer.Reset(HOST_TIMEOUT)
}

func main() {
	nstssh.Init(os.Getenv("HOME") + "/.ssh/id_rsa")
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
			ipsInRange, err := scan.IPRange(iprange)
			if err != nil {
				return err
			}
			ips = append(ips, ipsInRange...)
		}
		log.Println("Pinging hosts", ips)
		pingChan := make(chan string)
		scan.Init(pingChan)
		go scan.PingHosts(ips)
		scannerTimer = time.AfterFunc(HOST_TIMEOUT, timeoutScanner)
		for {
			select {
			case host := <-pingChan:
				go hostAlive(host)
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
