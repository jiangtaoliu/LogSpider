package scan

import (
	"github.com/tatsushid/go-fastping"
	"log"
	"net"
	"os"
	"time"
)

var pinger = fastping.NewPinger()
var responseChan chan string

func Init(response chan string) {
	responseChan = response
	pinger.OnRecv = listener
}

func listener(addr *net.IPAddr, rtt time.Duration) {
	responseChan <- addr.String()

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
