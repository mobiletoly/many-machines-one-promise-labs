package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	orders "github.com/mobiletoly/many-machines-one-promise-labs/capstones/01-two-servers-one-operation/solution"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	schema := os.Getenv("LAB_SCHEMA")
	if databaseURL == "" || schema == "" {
		log.Fatal("DATABASE_URL and LAB_SCHEMA are required")
	}

	store, err := orders.OpenStore(
		context.Background(),
		databaseURL,
		schema,
		os.Getenv("LAB_GATE_URL"),
		os.Getenv("LAB_INSTANCE_ID"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("http://%s\n", listener.Addr())
	if err := http.Serve(listener, orders.NewHandler(store)); err != nil {
		log.Fatal(err)
	}
}
