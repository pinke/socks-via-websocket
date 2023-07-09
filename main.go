package main

import (
	"flag"
	"github.com/pinke/socks-via-websocket/server"
)

var (
	Listener string
)

func main() {
	flag.StringVar(&Listener, "listen", "", "Listen as service ,eg :8080 or 127.0.0.1:8080")
	flag.Parse()
	if Listener == "" {
		flag.Usage()
		return
	}
	if Listener != "" {
		panic(server.Start(Listener))
	} else {

	}
}
