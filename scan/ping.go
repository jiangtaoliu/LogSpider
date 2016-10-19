package scan

import (
	"bytes"
	"errors"
	"log"
	"net"
	"strings"
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
}

func PingHosts(hosts []string) error {
	for _, host := range hosts {
		err := pinger.AddIP(host)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	pinger.RunLoop()
	return nil
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
