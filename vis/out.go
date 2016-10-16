package vis

import (
	"bytes"
	"encoding/json"
	"log"
	"net/url"
	"time"

	"net/http"

	"github.com/bahusvel/LogSpider/logs"
)

import "strconv"

// IngestBaseURL is the base URL for the LogDNA ingest API.
const IngestBaseURL = "https://logs.logdna.com/logs/ingest"

// DefaultFlushLimit is the number of log lines before we flush to LogDNA
const DefaultFlushLimit = 10

// Client is a client to the LogDNA logging service.
type Client struct {
	ApiKey          string
	connectionCache map[string]*payloadJSON
}

// logLineJSON represents a log line in the LogDNA ingest API JSON payload.
type logLineJSON struct {
	Timestamp int64  `json:"timestamp"`
	Line      string `json:"line"`
	File      string `json:"file"`
}

// payloadJSON is the complete JSON payload that will be sent to the LogDNA
// ingest API.
type payloadJSON struct {
	Lines []logLineJSON `json:"lines"`
}

// makeIngestURL creats a new URL to the a full LogDNA ingest API endpoint with
// API key and requierd parameters.
func (c Client) makeIngestURL(host string) url.URL {
	u, _ := url.Parse(IngestBaseURL)

	u.User = url.User(c.ApiKey)
	values := url.Values{}
	values.Set("hostname", host)
	values.Set("now", strconv.FormatInt(time.Now().UnixNano(), 10))
	u.RawQuery = values.Encode()

	return *u
}

// NewClient returns a Client configured to send logs to the LogDNA ingest API.
func NewClient(ApiKey string) *Client {
	var client Client
	client.ApiKey = ApiKey
	client.connectionCache = map[string]*payloadJSON{}

	return &client
}

// Log adds a new log line to Client's payload.
//
// To actually send the logs, Flush() needs to be called.
//
// Flush is called automatically if we reach the client's flush limit.
func (c *Client) Log(logEntry logs.LogEntry) {
	cache, ok := c.connectionCache[logEntry.Host]
	if !ok {
		cache = &payloadJSON{}
		c.connectionCache[logEntry.Host] = cache
	}
	if len(cache.Lines) == DefaultFlushLimit {
		c.Flush(logEntry.Host)
	}

	// Ingest API wants timestamp in milliseconds so we need to round timestamp
	// down from nanoseconds.
	logLine := logLineJSON{
		Timestamp: logEntry.Time.UnixNano() / 1000000,
		Line:      logEntry.Entry,
		File:      logEntry.Log,
	}
	cache.Lines = append(cache.Lines, logLine)
}

// Flush sends any buffered logs to LogDNA and clears the buffered logs.
func (c *Client) Flush(host string) error {
	// Return immediately if no logs to send
	if len(c.connectionCache[host].Lines) == 0 {
		return nil
	}

	jsonPayload, err := json.Marshal(c.connectionCache[host])
	if err != nil {
		return err
	}

	jsonReader := bytes.NewReader(jsonPayload)
	url := c.makeIngestURL(host)
	resp, err := http.Post(url.String(), "application/json", jsonReader)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	c.connectionCache[host] = &payloadJSON{}

	return err
}

// Close closes the client. It also sends any buffered logs.
func (c *Client) Close() error {
	for host, _ := range c.connectionCache {
		err := c.Flush(host)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}
