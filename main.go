package main

import (
	"bytes"
	"errors"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/bahusvel/NetworkScannerThingy/scan"
	"github.com/urfave/cli"
)

const HOST_TIMEOUT = 10 * time.Second

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
}

var hostMap = map[string]*Host{}
var scannerTimer *time.Timer

func newHost(address string) (*Host, error) {
	log.Println("New Host", address)
	return &Host{IPAddress: address}, nil
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

	existingHost.Status = STATUS_UP
	existingHost.StatusTime = time.Now()

	if existingHost.Status != STATUS_UP {
		hostUp(existingHost)
	}
}

func hostUp(host *Host) {
	log.Println("Host went back up", host)
}

func hostDown(host *Host) {
	host.Status = STATUS_DOWN
	host.StatusTime = time.Now()
	log.Println("Host down", host)
}

func timeoutScanner() {
	timeout := time.Now().Add(-HOST_TIMEOUT)
	for _, host := range hostMap {
		if host.StatusTime.Before(timeout) {
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
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name: "ipranges, i",
		},
	}

	app.Action = func(c *cli.Context) error {
		if len(c.StringSlice("ipranges")) == 0 {
			return cli.NewExitError("You did not specify any ranges", -1)
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
			}
		}
		//return nil
	}

	app.Run(os.Args)
}
