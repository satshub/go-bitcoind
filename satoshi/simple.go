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

func main() {
	rawTx, err := CreateTx("91izeJtyQ1DNGkiRtMGRKBEKYQTX46Ug8mGtKWpX9mDKqArsLpH",
		"mkYvnmm3KUBkvVqUAYsG6A6amt5Dva4jzX", 60000)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("raw signed transaction is: ", rawTx)
}

func NewTx() (*wire.MsgTx, error) {
	return wire.NewMsgTx(wire.TxVersion), nil
}

func GetUTXO(address string) (string, int64, string, error) {

	// Provide your url to get UTXOs, read the response
	// unmarshal it, and extract necessary data
	// newURL := fmt.Sprintf("https://your.favorite.block-explorer/%s", address)

	//response, err := http.Get(newURL)
	//if err != nil {
	// fmt.Println("error in GetUTXO, http.Get")
	// return nil, 0, "", err
	//}
	//defer response.Body.Close()
	//body, err := ioutil.ReadAll(response.Body)

	// based on the response you get, should define a struct
	// so before unmarshaling check your JSON response model

	//var blockChairResp = model.BlockChairResp{}
	//err = json.Unmarshal(body, &blockChairResp)
	//if err != nil {
	// fmt.Println("error in GetUTXO, json.Unmarshal")
	// return  nil, 0, "", err
	//}

	var previousTxid string = "16688d2946c3e029ca91ce730109994c2bcafb859d580a6f7c820fb2bb5b6afc"
	var balance int64 = 62000
	var pubKeyScript string = "76a91455d5e92958a8b06b4ff15cd2dd3d254f375e98db88ac"
	return previousTxid, balance, pubKeyScript, nil
}

func CreateTx(privKey string, destination string, amount int64) (string, error) {

	wif, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		return "", err
	}

	// use TestNet3Params for interacting with bitcoin testnet
	// if we want to interact with main net should use MainNetParams
	addrPubKey, err := btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeUncompressed(), &chaincfg.TestNet3Params)
	if err != nil {
		return "", err
	}

	txid, balance, pkScript, err := GetUTXO(addrPubKey.EncodeAddress())
	if err != nil {
		return "", err
	}

	/*
	 * 1 or unit-amount in Bitcoin is equal to 1 satoshi and 1 Bitcoin = 100000000 satoshi
	 */

	// checking for sufficiency of account
	if balance < amount {
		return "", fmt.Errorf("the balance of the account is not sufficient")
	}

	// extracting destination address as []byte from function argument (destination string)
	destinationAddr, err := btcutil.DecodeAddress(destination, &chaincfg.TestNet3Params)
	if err != nil {
		return "", err
	}

	destinationAddrByte, err := txscript.PayToAddrScript(destinationAddr)
	if err != nil {
		return "", err
	}

	// creating a new bitcoin transaction, different sections of the tx, including
	// input list (contain UTXOs) and outputlist (contain destination address and usually our address)
	// in next steps, sections will be field and pass to sign
	redeemTx, err := NewTx()
	if err != nil {
		return "", err
	}

	utxoHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		return "", err
	}

	// the second argument is vout or Tx-index, which is the index
	// of spending UTXO in the transaction that Txid referred to
	// in this case is 1, but can vary different numbers
	outPoint := wire.NewOutPoint(utxoHash, 1)

	// making the input, and adding it to transaction
	txIn := wire.NewTxIn(outPoint, nil, nil)
	redeemTx.AddTxIn(txIn)

	// adding the destination address and the amount to
	// the transaction as output
	redeemTxOut := wire.NewTxOut(amount, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	// now sign the transaction
	finalRawTx, err := SignTx(privKey, pkScript, redeemTx)

	return finalRawTx, nil
}

func SignTx(privKey string, pkScript string, redeemTx *wire.MsgTx) (string, error) {

	wif, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		return "", err
	}

	sourcePKScript, err := hex.DecodeString(pkScript)
	if err != nil {
		return "", nil
	}

	// since there is only one input in our transaction
	// we use 0 as second argument, if the transaction
	// has more args, should pass related index
	signature, err := txscript.SignatureScript(redeemTx, 0, sourcePKScript, txscript.SigHashAll, wif.PrivKey, false)
	if err != nil {
		return "", nil
	}

	// since there is only one input, and want to add
	// signature to it use 0 as index
	redeemTx.TxIn[0].SignatureScript = signature

	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)

	hexSignedTx := hex.EncodeToString(signedTx.Bytes())

	return hexSignedTx, nil
}
