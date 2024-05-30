/**
 * Description:
 * Author: Yihen.Liu
 * Create: 2021-07-30
 */
package p2sh

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
	wifStr1 := "cUUPbCpaRTdYXXFUXfDDSAHwEHbqayKMvzs1sQhgMcKtrAuPoUj7" //8725
	wif1, err := btcutil.DecodeWIF(wifStr1)
	if err != nil {
		return "", err
	}
	// public key extracted from wif.PrivKey
	pk1 := wif1.PrivKey.PubKey().SerializeCompressed()

	wifStr2 := "cTRFPhJZfRjYxf15pzE5XRocs4YMdzJb1nHW5g89RLjqm3w4zbtU" //3712.5
	wif2, err := btcutil.DecodeWIF(wifStr2)
	if err != nil {
		return "", err
	}
	pk2 := wif2.PrivKey.PubKey().SerializeCompressed()

	wifStr3 := "cSGrjKnJo6uxXCpf1YmKPwQmQDB7FyRk277GPvER1KDeSYJ13nUX" //1875.625
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
	utxoHash, err := chainhash.NewHashFromStr("f133a702b8a4692bb72a14d50f2830d3b2c0b53d2e07249b3021d65f2836ff77")
	if err != nil {
		return "", nil
	}

	// and add the index of the UTXO
	outPoint := wire.NewOutPoint(utxoHash, 1)

	txIn := wire.NewTxIn(outPoint, nil, nil)

	redeemTx.AddTxIn(txIn)

	// adding the output to tx
	decodedAddr, err := btcutil.DecodeAddress("bcrt1qa2u8nlqasxjkctuukjr4ve7wknvdd7mkvgn4qd", &chaincfg.RegressionNetParams)
	if err != nil {
		return "", err
	}
	destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
	if err != nil {
		return "", err
	}

	//adding the output to charge address
	chargeAddress, err := btcutil.DecodeAddress("bcrt1qht7jcacajzanpqa9y2tx7rz2ce9uckqwc2xype", &chaincfg.RegressionNetParams)
	if err != nil {
		return "", err
	}
	chargeAddressByte, err := txscript.PayToAddrScript(chargeAddress)
	if err != nil {
		return "", err
	}
	// adding the destination address and the amount to the transaction
	redeemTxOut := wire.NewTxOut(799999000, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	//charge Tx Out
	chargeTxOut := wire.NewTxOut(200000000, chargeAddressByte)
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
