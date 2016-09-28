package scan

import (
	"log"
	"net"
	"time"

	"github.com/tatsushid/go-fastping"
)

var pinger = fastping.NewPinger()

func Init(alive chan string) {
	pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		alive <- addr.String()
	}
}

func PingHosts(hosts []string) error {
	for _, host := range hosts {
		err := pinger.AddIP(host)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	pinger.Run()
	return nil
}
