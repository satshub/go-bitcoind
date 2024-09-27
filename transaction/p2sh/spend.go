/**
 * Description:
 * Author: Yihen.Liu
 * Create: 2021-07-30
 */
package main

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

// SpendMultiSig we can broadcast raw tx in: https://blockstream.info/testnet/tx/push or by bitcoin-cli sendrawtranscation.
func SpendMultiSig() (string, error) {
	// you can use your wif
	wifStr1 := "cSeEkntEySVvLHAjJ9PaypAUZFJacrCQSj73d8hXhykhETQ38nuu" //通过address包生成的私钥, Path(BIP84)
	wif1, err := btcutil.DecodeWIF(wifStr1)
	if err != nil {
		return "", err
	}
	// public key extracted from wif.PrivKey
	pk1 := wif1.PrivKey.PubKey().SerializeCompressed()

	wifStr2 := "cQJ1CwMcLeL8RZKTk3RNYyKPu9Q9SidQSdEy1KLrxRw1kZrjeNkT" //通过address包生成的私钥, Path(BIP84)
	wif2, err := btcutil.DecodeWIF(wifStr2)
	if err != nil {
		return "", err
	}
	pk2 := wif2.PrivKey.PubKey().SerializeCompressed()

	wifStr3 := "cUVDyeJwR17x1soRJWg9wPxkgyjpmnZjpV2WTSpK6CfLRnzqz49g" //通过address包生成的私钥, Path(BIP84)
	wif3, err := btcutil.DecodeWIF(wifStr3)
	if err != nil {
		return "", nil
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
		return "", err
	}

	redeemTx := wire.NewMsgTx(wire.TxVersion)

	// you should provide your UTXO hash
	utxoHash, err := chainhash.NewHashFromStr("be7f3a84f7c95fe18d677878be20eb722075697edfad2d96a01276b2f6b924bd")
	if err != nil {
		return "", nil
	}

	// and add the index of the UTXO
	outPoint := wire.NewOutPoint(utxoHash, 1)

	txIn := wire.NewTxIn(outPoint, nil, nil)

	redeemTx.AddTxIn(txIn)

	// adding the output to tx
	decodedAddr, err := btcutil.DecodeAddress("tb1qv7pu97v4fsws6ym6zxsysn9umdr677rq0sh777", &chaincfg.SigNetParams)
	if err != nil {
		return "", err
	}
	destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
	if err != nil {
		return "", err
	}

	//adding the output to charge address
	chargeAddress, err := btcutil.DecodeAddress("tb1qz7f44c225fkdnu2ywfr9nj4h2rns0vyxn9jfwn", &chaincfg.SigNetParams)
	if err != nil {
		return "", err
	}
	chargeAddressByte, err := txscript.PayToAddrScript(chargeAddress)
	if err != nil {
		return "", err
	}
	// adding the destination address and the amount to the transaction
	redeemTxOut := wire.NewTxOut(800000, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	//charge Tx Out
	chargeTxOut := wire.NewTxOut(100000, chargeAddressByte)
	redeemTx.AddTxOut(chargeTxOut)
	// signing the tx

	sig1, err := txscript.RawTxInSignature(redeemTx, 0, redeemScript, txscript.SigHashAll, wif1.PrivKey)
	if err != nil {
		return "", err
	}

	//sig2, err := txscript.RawTxInSignature(redeemTx, 0, redeemScript, txscript.SigHashAll, wif2.PrivKey)
	//if err != nil {
	//	return "", err
	//}

	sig3, err := txscript.RawTxInSignature(redeemTx, 0, redeemScript, txscript.SigHashAll, wif3.PrivKey)
	if err != nil {
		fmt.Println("got error in constructing sig3")
		return "", err
	}

	signature := txscript.NewScriptBuilder()
	signature.AddOp(txscript.OP_FALSE).AddData(sig1)
	signature.AddData(sig3).AddData(redeemScript)
	signatureScript, err := signature.Script()
	if err != nil {
		// Handle the error.
		return "", err
	}

	redeemTx.TxIn[0].SignatureScript = signatureScript

	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)

	hexSignedTx := hex.EncodeToString(signedTx.Bytes())

	return hexSignedTx, nil
}
