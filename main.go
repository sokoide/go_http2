package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type options struct {
	url    string
	tls    string
	key    string
	cert   string
	cacert string
	count  int
}

var o options

func parseFlags() {
	flag.StringVar(&o.url, "url", "lab2:23001", "URL of Fluent-bit http-in")
	flag.StringVar(&o.tls, "tls", "none", "none | mTLS")
	flag.StringVar(&o.key, "key", "", "Key file for mTLS")
	flag.StringVar(&o.cert, "cert", "", "Cert file for mTLS")
	flag.StringVar(&o.cacert, "cacert", "", "CA Cert file for mTLS")
	flag.IntVar(&o.count, "count", 5, "Number of requests to send")
	flag.Parse()
}

func getTlsConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair(o.cert, o.key)
	if err != nil {
		panic(err)
	}
	caCert, err := os.ReadFile(o.cacert)
	if err != nil {
		panic(err)
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		panic("failed to parse CA cert")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:          caCertPool,
		InsecureSkipVerify: true,
	}
	return tlsConfig
}

func main() {
	parseFlags()

	logrus.Info("started...")
	url := o.url
	ct := "application/json"
	d := map[string]string{
		"key": "value",
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 10 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2: true,
	}

	if o.tls == "mtls" {
		transport.TLSClientConfig = getTlsConfig()
	}

	client := &http.Client{Transport: transport}

	logrus.Infof("Sending log to %s", url)

	data, err := json.Marshal(d)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		logrus.Error("failed to create a request")
		return
	}

	req.Header.Set("Content-Type", ct)
	for i := range o.count {
		resp, err := client.Do(req)
		if err != nil {
			logrus.Errorf("failed to do client.Do. err:%+v", err)
		} else {
			logrus.Infof("%d: Protocol: %s, StatusCode: %d", i, resp.Proto, resp.StatusCode)
		}
	}
}
