package main

import (
	"flag"

	"github.com/oremj/webpush-simulator/simulator"
)

var concurrentConnectionLimit int
var connectionLimit int

var testURL string

func init() {
	flag.IntVar(&concurrentConnectionLimit, "concurrency", 10, "how many concurrent connection attempts")
	flag.IntVar(&connectionLimit, "connections", 100, "how many connections to establish")
	flag.StringVar(&testURL, "url", "wss://autopush.stage.mozaws.net/", "url to test against")
}

func main() {
	flag.Parse()
	simulator := simulator.New(simulator.Options{
		Connections:           connectionLimit,
		ConcurrentConnections: concurrentConnectionLimit,
		PushUrl:               testURL,
	})
	simulator.Run()
}
