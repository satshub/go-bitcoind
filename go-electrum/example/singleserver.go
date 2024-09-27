package main

import (
	"context"
	"log"
	"time"

	"github.com/satshub/go-bitcoind/go-electrum/electrum"
	//"github.com/checksum0/go-electrum/electrum"
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
	//tb1pst5dyk7tdymccy0xydcyyzvqgz2022t8576sjz2l65fzulrvy5rqcv6les
	scriptHash, err := electrum.AddressToElectrumScriptHash("2NCAZEVJwprU5vX7uTgBwj9Z3Jq4Xaa93LQ")
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

	res, err := client.BroadcastTransaction(context.Background(), "0100000001715694eda1e7388a860b4dc62bc2d888b7ad9cd9f2de4299f67348419c09847601000000fd44010047304402203d9a658b835631e27e964f45e47a2b7f1e0bfd6ec71353d607dfefb7d643bed8022034a912b38ec2d8364d0d5cd5c6b1aaf6d02909c0438550f61fbb47b5a3d9c60e0147304402200a563c98ad80df23baa489412336ff8b28c94263ab9db3eb470f6b5d09136b8902202c865a0d6257431640a34e2cd65ac696ff06ab8e2d2fd62a1bb8e74d63d4763501473044022058a960b1fdf19160d668eeb373162cb5b7e7e6a6f4ccfec56313cc895a0bf25d022038da5970b2f567cf28a6cae6c7e4cdcb1e6193237d350d71825b731d50037f45014c6952210291becfb80f5875b7cac527b0643b1b5962c8c7f735d7eaeefa558c31d4e6b04d21023770069d381fd9ac4ef66e37e974caee3ca34e03a0f325c91e2563ab79c3191d21021ba401b1e6deb228c0cfa27fb6d5515424a794d10d665c5d742cab10f5648bae53aeffffffff0290940d00000000001600146783c2f9954c1d0d137a11a0484cbcdb47af7860a08601000000000016001417935ae14aa26cd9f144724659cab750e707b08600000000")
	if err != nil {
		log.Fatal("broadcast error:", err)
	}
	log.Printf("broadcast result: %v", res)
	select {}
}
