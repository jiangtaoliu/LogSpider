package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
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

var hostMap = map[string]*Host{}
var scannerTimer *time.Timer
var logChannel = make(chan logs.LogEntry)

func newHost(address string) (*Host, error) {
	log.Println("New Host", address)
	return &Host{IPAddress: address, SSHEnabled: true}, nil
}

func hostAlive(host string) {
	if _, ok := hostMap[host]; !ok {
		tmpHost, err := newHost(host)
		if err != nil {
			log.Println(err)
			return
		}
		hostMap[host] = tmpHost
	}
	existingHost := hostMap[host]

	if existingHost.Status != STATUS_UP {
		go hostUp(existingHost)
	} else {
		existingHost.Status = STATUS_UP
		existingHost.StatusTime = time.Now()
	}
}

func hostUp(host *Host) {
	host.Status = STATUS_UP
	host.StatusTime = time.Now()
	log.Println("Host went back up", host)

	if !host.SSHEnabled {
		return
	}

	err := nstssh.CopyID("localhost", host.IPAddress, "cp-x2520")
	if err != nil {
		log.Println("Cannot establish ssh connectivity to", host)
		host.SSHEnabled = false
		return
	} else {
		log.Println("CopyID to", host, "Successful")
	}

	hostLogs, err := logs.FindLogs(host.IPAddress)
	if err != nil {
		log.Println("Cannot fetch logs from", host)
		return
	}

	for _, hostLog := range hostLogs {
		go logs.WatchLog(host.IPAddress, hostLog, logChannel)
	}
}

func hostDown(host *Host) {
	host.Status = STATUS_DOWN
	host.StatusTime = time.Now()
	log.Println("Host down", host)
}

func timeoutScanner() {
	timeout := time.Now().Add(-HOST_TIMEOUT)
	for _, host := range hostMap {
		if host.Status == STATUS_UP && host.StatusTime.Before(timeout) {
			hostDown(host)
		}
	}
	scannerTimer.Reset(HOST_TIMEOUT)
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
			ipsInRange, err := IPRange(iprange)
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
				hostAlive(host)
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
