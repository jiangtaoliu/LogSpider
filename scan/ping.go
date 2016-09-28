package scan

import (
	"log"
	"net"
	"time"

	"github.com/tatsushid/go-fastping"
)

const PING_INTERVAL = 5 * time.Second

var pinger = fastping.NewPinger()

func Init(alive chan string) {
	pinger.MaxRTT = PING_INTERVAL
	pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		alive <- addr.String()
	}
	pinger.OnIdle = Reping
}

func PingHosts(hosts []string) error {
	for _, host := range hosts {
		err := pinger.AddIP(host)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return pinger.Run()
}

func Reping() {
	err := pinger.Run()
	if err != nil {
		log.Println(err)
	}
}
