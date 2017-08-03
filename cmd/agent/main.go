package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/hpcloud/tail"
	"github.com/urfave/cli"
)

type Priority int

//var logs []chan *tail.Line
var logs []reflect.SelectCase
var logPaths []string
var conn net.Conn
var hostName, _ = os.Hostname()
var Quiet = false

//var re = regexp.MustCompile("[[:^ascii:]]")

var blacklist = []string{"flowd.log"}

type LogLine struct {
	Line string
	Time time.Time
	File string
}

func newFilePoller(Dirs []string) {
	for _, dir := range Dirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if !strings.HasSuffix(path, "log") {
				return nil
			}
			for _, b := range blacklist {
				if strings.HasSuffix(path, b) {
					return nil
				}
			}
			log.Println(path)
			t, err := tail.TailFile(path, tail.Config{Follow: true, ReOpen: true})
			if err != nil {
				log.Println("Error", path, err)
			}
			logPaths = append(logPaths, path)
			logs = append(logs, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(t.Lines)})

			return nil
		})
	}
}

func interrupt(intchan chan int) {
	for {
		intchan <- 0
		time.Sleep(1 * time.Second)
	}
}

func DefaultFormatter(p Priority, hostname, tag, content string) string {
	timestamp := time.Now().Format(time.RFC3339)
	msg := fmt.Sprintf("<%d> %s %s %s[%d]: %s",
		p, timestamp, hostname, tag, 0, content)
	return msg
}

// UnixFormatter omits the hostname, because it is only used locally.
func UnixFormatter(p Priority, t time.Time, hostname, tag, content string) string {
	timestamp := t.Format(time.Stamp)
	msg := fmt.Sprintf("<%d>%s %s[%d]: %s",
		p, timestamp, tag, os.Getpid(), content)
	return msg
}

// RFC3164Formatter provides an RFC 3164 compliant message.
func RFC3164Formatter(p Priority, hostname, tag, content string) string {
	timestamp := time.Now().Format(time.Stamp)
	msg := fmt.Sprintf("<%d>%s %s %s[%d]: %s",
		p, timestamp, hostname, tag, os.Getpid(), content)
	return msg
}

// RFC5424Formatter provides an RFC 5424 compliant message.
func RFC5424Formatter(p Priority, hostname, tag, content string) string {
	timestamp := time.Now().Format(time.RFC3339)
	pid := os.Getpid()
	appName := os.Args[0]
	msg := fmt.Sprintf("<%d>%d %s %s %s %d %s - %s",
		p, 1, timestamp, hostname, appName, pid, tag, content)
	return msg
}

func ProcessLogEntry(line *LogLine) {
	conn.Write([]byte(DefaultFormatter(0, hostName, line.File, line.Line)))
	if !Quiet {
		log.Println(line.File, line.Line)
	}
}

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name: "log-dir, d",
		},
		cli.StringFlag{
			Name: "server, s",
		},
		cli.StringSliceFlag{
			Name: "blacklist, b",
		},
		cli.BoolFlag{
			Name:        "queit, q",
			Destination: &Quiet,
		},
	}
	//knowledgeBase := make(analyse.KnowledgeDB)
	app.Action = func(c *cli.Context) error {
		logDirs := c.StringSlice("log-dir")
		if len(logDirs) == 0 {
			logDirs = append(logDirs, "/var/log")
		}

		if c.String("server") == "" {
			return cli.NewExitError("You must provide the server IP", -1)
		}

		blacklist = append(blacklist, c.StringSlice("blacklist")...)

		var err error
		conn, err = net.Dial("udp", c.String("server"))
		if err != nil {
			return err
		}

		intchan := make(chan int)
		logs = append(logs, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(intchan)})
		//go interrupt(intchan)
		newFilePoller(logDirs)

		for {
			i, value, ok := reflect.Select(logs)
			if !ok {
				continue
			}
			switch value.Interface().(type) {
			case *tail.Line:
				tline := value.Interface().(*tail.Line)
				line := LogLine{Line: tline.Text, Time: tline.Time, File: logPaths[i-1]}
				ProcessLogEntry(&line)

				//time.Sleep(10 * time.Millisecond)
			default:
			}

		}
		//return nil
	}

	app.Run(os.Args)
}
