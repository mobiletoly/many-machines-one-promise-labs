package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	g42 "github.com/mobiletoly/many-machines-one-promise-labs/episodes/03-g42-partition/start"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("http://%s\n", listener.Addr())
	if err := http.Serve(listener, g42.NewHandler(g42.NewAccount())); err != nil {
		log.Fatal(err)
	}
}
