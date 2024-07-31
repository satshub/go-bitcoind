package main

import (
	"context"
	"log"
	"time"

	"github.com/checksum0/go-electrum/electrum"
)

func main() {
	client, err := electrum.NewClientTCP(context.Background(), "106.75.5.22:50001")

	if err != nil {
		log.Fatal(err)
	}

	serverVer, protocolVer, err := client.ServerVersion(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Server version: %s [Protocol %s]", serverVer, protocolVer)

	go func() {
		for {
			if err := client.Ping(context.Background()); err != nil {
				log.Fatal(err)
			}
			if banner, err := client.ServerBanner(context.Background()); err != nil {
				log.Fatal(err)
			} else {
				log.Printf("Server banner: %s", banner)
			}
			time.Sleep(3 * time.Second)
		}
	}()
	select {}
}
