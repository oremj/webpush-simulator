package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/oremj/webpush-simulator/simulator"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var concurrentConnectionLimit int
var connectionLimit int

var testURL string

var metricAddr string

func init() {
	flag.IntVar(&concurrentConnectionLimit, "concurrency", 10, "how many concurrent connection attempts")
	flag.IntVar(&connectionLimit, "connections", 100, "how many connections to establish")
	flag.StringVar(&testURL, "url", "wss://autopush.stage.mozaws.net/", "url to test against")

	flag.StringVar(&metricAddr, "metricaddr", ":80", "metric http listen address")
}

func main() {
	flag.Parse()

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Println(http.ListenAndServe(metricAddr, nil))
	}()

	simulator := simulator.New(simulator.Options{
		Connections:           connectionLimit,
		ConcurrentConnections: concurrentConnectionLimit,
		PushUrl:               testURL,
	})
	simulator.Run()
}
