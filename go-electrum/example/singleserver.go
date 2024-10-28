package main

import (
	"context"
	"log"
	"time"

	"github.com/satshub/go-bitcoind/go-electrum/electrum"
	//"github.com/checksum0/go-electrum/electrum"
)

func main() {
	client, err := electrum.NewClientTCP(context.Background(), "45.43.60.97:60601")

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

	scriptHash, err := electrum.AddressToElectrumScriptHash("tb1pph5avhhj5qsv6fpmfvcwc0klhhyndss4xh360a27uumsuc0mquxsen7lel")
	if err != nil {
		log.Fatal("address to script hash error:", err)
	} else {
		utxos, err := client.ListUnspent(context.Background(), scriptHash)
		if err != nil {
			log.Fatal("list unspent error:", err)
		} else {
			for _, utxo := range utxos {
				log.Printf("utxo: %+v", utxo)

			}

		}
	}
	//tb1pst5dyk7tdymccy0xydcyyzvqgz2022t8576sjz2l65fzulrvy5rqcv6les
	/*
		scriptHash, err := electrum.AddressToElectrumScriptHash("tb1psntfmv0zj708fahh6xtwydjtdsht92yd20t42rr9uun6eapf4fwqdc8w7w")
		if err != nil {
			log.Fatal("address to script hash error:", err)
		} else {
			utxos, err := client.ListUnspent(context.Background(), scriptHash)
			if err != nil {
				log.Fatal("list unspent error:", err)
			} else {
				for _, utxo := range utxos {
					log.Printf("utxo: %v", utxo)

				}

			}
		}
		res, err := client.BroadcastTransaction(context.Background(), "010000000001019ab40807b2f27d9a31be7845fe542160fd05ca4563bc3ba2fa640a9434d00ea60000000000ffffffff0120f40e0000000000160014671e42d7bc7bb57990fe21cf05a09c31ea982f5f0140f937a331e63897761dbbe0a298eabd2a68584e1ce785c6f5c552e4cffebb2eb4b132899b11539efb40c8efc0c235d4963b389b35b517b6ff67577fa166a3837800000000")
		if err != nil {
			log.Fatal("broadcast error:", err)
		}
		log.Printf("broadcast result: %v", res)
	*/
	select {}
}
