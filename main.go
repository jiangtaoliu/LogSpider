package main

import (
	"log"
	"os"
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

func identifyHost(address string) (*Host, error) {
	return &Host{IPAddress: address, Status: STATUS_UP}, nil
}

func hostAlive(host string) {
	if _, ok := hostMap[host]; !ok {
		tmpHost, err := identifyHost(host)
		if err != nil {
			log.Println(err)
			return
		}
		// setup other services to run and gather information for the host

		hostMap[host] = tmpHost
	}
	existingHost := hostMap[host]
	if existingHost.Status != STATUS_UP {
		hostUp(existingHost)
	}

}

func hostUp(host *Host) {
	host.Status = STATUS_UP
	host.StatusTime = time.Now()
}

func hostDown(host *Host) {
	host.Status = STATUS_DOWN
	host.StatusTime = time.Now()
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

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name: "ipranges, -i",
		},
	}

	app.Action = func(c *cli.Context) error {
		if len(c.StringSlice("ipranges")) == 0 {
			return cli.NewExitError("You did not specify any ranges", -1)
		}
		pingChan := make(chan string)
		scan.Init(pingChan)
		scan.PingHosts(c.StringSlice("ipranges"))
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
