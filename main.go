package main

import (
	"log"
	"os"

	"github.com/bahusvel/NetworkScannerThingy/scan"
	"github.com/urfave/cli"
)

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
	IPAddress string
	OS        string
	Status    string
}

var hostMap = map[string]*Host{}

func identifyHost(address string) (*Host, error) {
	return &Host{IPAddress: address, Status: STATUS_UP}, nil
}

func hostAlive(host string) {
	if existingHost, ok := hostMap[host]; !ok {
		tmpHost, err := identifyHost(host)
		if err != nil {
			log.Println(err)
			return
		}

		// setup other services to run and gather information for the host

		hostMap[host] = tmpHost
	} else {
		existingHost.Status = STATUS_UP
	}
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
