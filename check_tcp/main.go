package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

const name = "info.tdoc.newrelic.check_tcp"
const procotolVersion = "1"
const integrationVersion = "0.0.1"
const eventType = "TdocCheckTCP"
const DefaultTimeout = 100 // msec

const (
	CodeOK          = 200
	ConnectionError = 500
	ConnectTimeout  = 501
	ReadTimeout     = 502
	Closed          = 503
	DNSError        = 504
	DNSTimeout      = 505
)

var strMap = map[int]string{
	CodeOK:          "OK",
	ConnectionError: "ConnectionError",
	ConnectTimeout:  "ConnectTimeout",
	ReadTimeout:     "ReadTimeout",
	Closed:          "Closed",
	DNSError:        "DNSError",
	DNSTimeout:      "DNSTimeout",
}

var errNoSuchHost = errors.New("no such host")

type MetricData struct {
	EventType  string `json:"event_type"`
	StatusCode int    `json:"status_code"`
	Status     string `json:"status"`
}

type IntegrationData struct {
	Name               string       `json:"name"`
	ProtocolVersion    string       `json:"protocol_version"`
	IntegrationVersion string       `json:"integration_version"`
	Metrics            []MetricData `json:"metrics"`
}

func checkTCP(addr string, timeout int) (int, error) {
	to := time.Duration(timeout) * time.Millisecond
	c, err := net.DialTimeout("tcp", addr, to)
	if operror, ok := err.(*net.OpError); ok {
		if operror.Timeout() {
			return ConnectTimeout, nil
		}
		if strings.HasSuffix(operror.Error(), errNoSuchHost.Error()) {
			log.WithError(operror).Debug("LookupError: " + addr)
			return DNSError, nil
		}
		log.WithError(operror).Debug("OpError: " + addr)
		return ConnectionError, nil
	}
	if dnserr, ok := err.(*net.DNSError); ok {
		if dnserr.IsTimeout {
			return DNSTimeout, nil
		}
		log.WithError(dnserr).Debug("DNSError: " + addr)
		return DNSError, nil
	}
	if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
		return ConnectTimeout, nil
	}
	if err != nil {
		log.WithError(err).Debug("someerror: " + addr)
		return ConnectionError, nil
	}

	c.(*net.TCPConn).SetNoDelay(true)
	c.SetReadDeadline(time.Now().Add(to))

	_, err = c.Read([]byte{})
	if err == io.EOF {
		return Closed, nil
	}
	if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
		return ReadTimeout, nil
	}
	return CodeOK, nil
}

func getArgs() (addr string, timeout int) {
	addr = os.Getenv("ADDR")

	t := os.Getenv("TIMEOUT")
	if t != "" {
		tmp, err := strconv.ParseInt(t, 10, 16)
		if err != nil {
			log.WithError(err).Fatal("invalid timeout:" + t)
		}
		timeout = int(tmp)
	} else {
		timeout = DefaultTimeout
	}

	return addr, timeout
}

func main() {
	verbose := flag.Bool("v", false, "Print more information to logs")
	flag.Parse()

	log.SetOutput(os.Stderr)
	if *verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	data := IntegrationData{
		Name:               name,
		ProtocolVersion:    procotolVersion,
		IntegrationVersion: integrationVersion,
		Metrics:            make([]MetricData, 0),
	}

	addr, timeout := getArgs()
	if addr == "" {
		log.Fatal("invalid dest: " + addr)
	}

	code, _ := checkTCP(addr, timeout)
	str := strMap[code]
	metric := MetricData{
		EventType:  eventType,
		StatusCode: code,
		Status:     str,
	}

	data.Metrics = append(data.Metrics, metric)
	output, err := json.Marshal(data)
	if err != nil {
		log.WithError(err).Fatal("json marshal")
	}

	if string(output) == "null" {
		fmt.Println("[]")
	} else {
		fmt.Println(string(output))
	}
}
