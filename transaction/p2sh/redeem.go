// Package p2sh_2_of_3 /**
package main

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

func BuildMultiSigRedeemScript() (string, string, string, error) {

	// you can use your wif

	//wifStr1 := "cQpHXfs91s5eR9PWXui6qo2xjoJb2X3VdUspwKXe4A8Dybvut2rL"
	wifStr1 := "cSeEkntEySVvLHAjJ9PaypAUZFJacrCQSj73d8hXhykhETQ38nuu" //8725
	wif1, err := btcutil.DecodeWIF(wifStr1)
	if err != nil {
		return "", "", "", err
	}
	// public key extracted from wif.PrivKey
	pk1 := wif1.PrivKey.PubKey().SerializeCompressed()

	//wifStr2 := "cVgxEkRBtnfvd41ssd4PCsiemahAHidFrLWYoDBMNojUeME8dojZ"
	wifStr2 := "cQJ1CwMcLeL8RZKTk3RNYyKPu9Q9SidQSdEy1KLrxRw1kZrjeNkT" //3712.5
	wif2, err := btcutil.DecodeWIF(wifStr2)
	if err != nil {
		return "", "", "", err
	}
	pk2 := wif2.PrivKey.PubKey().SerializeCompressed()

	//wifStr3 := "cPXZBMz5pKytwCyUNAdq94R9VafU8L2QmAW8uw3gKrzjuCWCd3TM"
	wifStr3 := "cUVDyeJwR17x1soRJWg9wPxkgyjpmnZjpV2WTSpK6CfLRnzqz49g" //1875.625
	wif3, err := btcutil.DecodeWIF(wifStr3)
	if err != nil {
		return "", "", "", nil
	}
	pk3 := wif3.PrivKey.PubKey().SerializeCompressed()

	// create redeem script for 2 of 3 multi-sig
	builder := txscript.NewScriptBuilder()
	// add the minimum number of needed signatures
	builder.AddOp(txscript.OP_2)
	// add the 3 public key
	builder.AddData(pk1).AddData(pk2).AddData(pk3)
	// add the total number of public keys in the multi-sig screipt
	builder.AddOp(txscript.OP_3)
	// add the check-multi-sig op-code
	builder.AddOp(txscript.OP_CHECKMULTISIG)
	// redeem script is the script program in the format of []byte
	redeemScript, err := builder.Script()
	if err != nil {
		return "", "", "", err
	}

	// dis asemble the script program, so can see its structure
	redeemStr, err := txscript.DisasmString(redeemScript)
	if err != nil {
		return "", "", "", nil
	}

	// calculate the hash160 of the redeem script
	redeemHash := btcutil.Hash160(redeemScript)

	// if using Bitcoin main net then pass &chaincfg.MainNetParams as second argument
	addr, err := btcutil.NewAddressScriptHashFromHash(redeemHash, &chaincfg.SigNetParams)
	if err != nil {
		return "", "", "", err
	}

	return redeemStr, hex.EncodeToString(redeemHash), addr.EncodeAddress(), nil

}
