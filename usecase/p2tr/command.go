package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/satshub/go-bitcoind/go-electrum/electrum"
	"github.com/satshub/go-bitcoind/usecase/p2tr/config"
	"github.com/satshub/go-bitcoind/usecase/p2tr/config/utils"
	"github.com/satshub/go-bitcoind/usecase/p2tr/log"
	"github.com/urfave/cli"
)

var SinglePrivateKeyExecutive = cli.Command{
	//Usage:     "Import blocks to DB from a file",
	Name:        "single",
	ArgsUsage:   "",
	Action:      spentBTCWithinSinglePrivateKey,
	Flags:       []cli.Flag{},
	Description: "Generate p2sh Redeem",
}

var SpentExecutive = cli.Command{
	//Usage:     "Import blocks to DB from a file",
	Name:        "spent",
	ArgsUsage:   "",
	Action:      spentGenerator,
	Flags:       []cli.Flag{},
	Description: "Generate p2sh Redeem",
}

var BroadcastExecutive = cli.Command{
	//Usage:     "Import blocks to DB from a file",
	Name:        "broadcast",
	ArgsUsage:   "--hex=<raw transaction hex encode>",
	Action:      doBroadcast,
	Flags:       []cli.Flag{utils.HexFlag},
	Description: "broadcast transaction",
}

func doBroadcast(ctx *cli.Context) error {
	log.InitLog(config.AppConf.Logger.LogLevel, config.AppConf.Logger.LogFileDir, log.Stdout)

	rawTx := ctx.String(utils.GetFlagName(utils.HexFlag))
	client, err := electrum.NewClientTCP(context.Background(), config.AppConf.Electrum)
	log.Info("rawTx:", rawTx, "electrum:", config.AppConf.Electrum)
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.BroadcastTransaction(context.Background(), rawTx)
	if err != nil {
		log.Fatalf("broadcast error:%+v", err)
	}

	log.Infof("broadcast result: %v", res)
	return nil
}

func spentGenerator(ctx *cli.Context) error {
	log.InitLog(config.AppConf.Logger.LogLevel, config.AppConf.Logger.LogFileDir, log.Stdout)

	sort.Strings(config.AppConf.Signers.WIFs)
	spendHex, err := SpendMultiSig(config.AppConf.Signers.WIFs, config.AppConf.Signers.Threshold)
	if err == nil {
		fmt.Println("spend hex:", spendHex)
	} else {
		log.Errorf("spend error: %v", err)
	}

	return nil
}

func spentBTCWithinSinglePrivateKey(ctx *cli.Context) error {
	log.InitLog(config.AppConf.Logger.LogLevel, config.AppConf.Logger.LogFileDir, log.Stdout)

	//sort.Strings(config.AppConf.Signers.WIFs)
	//redeemScript, redeemHash, addr, err := BuildMultiSigRedeemScript(config.AppConf.Signers.WIFs, config.AppConf.Signers.Threshold)
	BuildP2trTransaction()
	/*
		if err == nil {
			fmt.Println("redeem script:", redeemScript)
			fmt.Println("redeem hash:", redeemHash)
			fmt.Println("p2sh addr:", addr)
			fmt.Println("network:", config.AppConf.Network)
		}
	*/
	return nil
}

func redeemScript(wifs []string, threshold int) ([]byte, error) {
	if len(wifs) < threshold {
		panic("not enough wifs")
	}

	extractPubKey := func(wif string) ([]byte, error) {
		wif1, err := btcutil.DecodeWIF(wif)
		if err != nil {
			return nil, err
		}
		return wif1.PrivKey.PubKey().SerializeCompressed(), nil
	}

	pks := make([][]byte, len(wifs))
	for i, wif := range wifs {
		pk, err := extractPubKey(wif)
		if err != nil {
			panic("extractPubKey" + err.Error())
		}
		pks[i] = pk
	}

	// create redeem script for 2 of 3 multi-sig
	builder := txscript.NewScriptBuilder()
	// add the minimum number of needed signatures
	builder.AddOp(byte(threshold + 80) /*txscript.OP_2*/)
	//builder.AddOp(txscript.OP_2)
	// add the 3 public key
	for _, pk := range pks {
		builder.AddData(pk)
	}
	//builder.AddData(pk1).AddData(pk2).AddData(pk3)
	// add the total number of public keys in the multi-sig screipt
	builder.AddOp(byte(len(pks) + 80) /*txscript.OP_3*/)
	// add the check-multi-sig op-code
	builder.AddOp(txscript.OP_CHECKMULTISIG)
	// redeem script is the script program in the format of []byte
	return builder.Script()
}

func BuildMultiSigRedeemScript(wifs []string, threshold int) (string, string, string, error) {
	redeemScript, err := redeemScript(wifs, threshold)
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
	addr, err := btcutil.NewAddressScriptHashFromHash(redeemHash, NetworkParams(config.AppConf.Network))
	if err != nil {
		return "", "", "", err
	}

	return redeemStr, hex.EncodeToString(redeemHash), addr.EncodeAddress(), nil
}

func SpendMultiSig(wifs []string, threshold int) (string, error) {
	redeemScript, err := redeemScript(wifs, threshold)
	if err != nil {
		return "", err
	}

	redeemTx := wire.NewMsgTx(wire.TxVersion)

	// you should provide your UTXO hash
	utxoHash, err := chainhash.NewHashFromStr(config.AppConf.Spent.TxId)
	if err != nil {
		return "", nil
	}

	// and add the index of the UTXO
	outPoint := wire.NewOutPoint(utxoHash, config.AppConf.Spent.UtxoIndex)

	txIn := wire.NewTxIn(outPoint, nil, nil)

	redeemTx.AddTxIn(txIn)

	// adding the output to tx
	decodedAddr, err := btcutil.DecodeAddress(config.AppConf.Spent.Receiver, NetworkParams(config.AppConf.Network))
	if err != nil {
		return "", errors.New("error decoding dest address " + err.Error())
	}
	destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
	if err != nil {
		return "", err
	}

	//adding the output to charge address
	chargeAddress, err := btcutil.DecodeAddress(config.AppConf.Spent.Charger, NetworkParams(config.AppConf.Network))
	if err != nil {
		return "", errors.New("error decoding charge address " + err.Error())
	}
	chargeAddressByte, err := txscript.PayToAddrScript(chargeAddress)
	if err != nil {
		return "", err
	}
	// adding the destination address and the amount to the transaction
	redeemTxOut := wire.NewTxOut(config.AppConf.Spent.ReceiverAmount, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	//charge Tx Out
	chargeTxOut := wire.NewTxOut(config.AppConf.Spent.ChargerAmount, chargeAddressByte)
	redeemTx.AddTxOut(chargeTxOut)
	// signing the tx

	signature := txscript.NewScriptBuilder()
	signature.AddOp(txscript.OP_FALSE)
	for _, wif := range config.AppConf.Spent.WIFs {
		wif1, err := btcutil.DecodeWIF(wif)
		if err != nil {
			return "", err
		}
		sig, err := txscript.RawTxInSignature(redeemTx, 0, redeemScript, txscript.SigHashAll, wif1.PrivKey)
		if err != nil {
			return "", err
		}
		signature.AddData(sig)
	}
	signature.AddData(redeemScript)
	signatureScript, err := signature.Script()
	if err != nil {
		return "", err
	}

	redeemTx.TxIn[0].SignatureScript = signatureScript

	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)

	hexSignedTx := hex.EncodeToString(signedTx.Bytes())

	return hexSignedTx, nil
}

func NetworkParams(network string) *chaincfg.Params {
	switch network {
	case "mainnet":
		return &chaincfg.MainNetParams
	case "testnet":
		return &chaincfg.TestNet3Params
	case "signet":
		return &chaincfg.SigNetParams
	case "simnet":
		return &chaincfg.SimNetParams
	default:
		return &chaincfg.MainNetParams
	}
}
