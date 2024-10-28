package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func GenerateRandP2trAddr() (string, *btcutil.WIF) {
	// 博文的代码都默认为测试网
	cfg := &chaincfg.SigNetParams

	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		log.Fatalln(err)
		return "", nil
	}

	wif, err := btcutil.NewWIF(privateKey, cfg, true)
	fmt.Printf("Generated WIF Key: %s", wif.String())
	// Generated WIF Key: cViUtGHsa6XUxxk2Qht23NKJvEzQq5mJYQVFRsEbB1PmSHMmBs4T

	taprootAddr, err := btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(
			txscript.ComputeTaprootKeyNoScript(
				wif.PrivKey.PubKey())),
		&chaincfg.SigNetParams)

	log.Printf("Taproot testnet address: %s\n", taprootAddr.String())
	// Taproot testnet address: tb1p3d3l9m5d0gu9uykqurm4n8xcdmmw9tkhh8srxa32lvth79kz7vysx9jgcr

	return taprootAddr.EncodeAddress(), wif
}

func GetUnspent(address string) (*wire.OutPoint, *txscript.MultiPrevOutFetcher) {
	// 交易的哈希值，并且要指定输出位置
	txHash, err := chainhash.NewHashFromStr(
		"a60ed034940a64faa23bbc6345ca05fd602154fe4578be319a7df2b20708b49a")
	if err != nil {
		panic("hash error")
	}
	point := wire.NewOutPoint(txHash, uint32(0))

	// 交易的锁定脚本，对应的是 ScriptPubKey 字段
	script, err := hex.DecodeString("512084d69db1e2979e74f6f7d196e2364b6c2eb2a88d53d7550c65e727acf429aa5c")
	if err != nil {
		panic("decode script error")
	}
	output := wire.NewTxOut(int64(1000000), script)
	fetcher := txscript.NewMultiPrevOutFetcher(nil)
	fetcher.AddPrevOut(*point, output)

	return point, fetcher
}

func DecodeBitcoinAddress(strAddr string, cfg *chaincfg.Params) ([]byte,
	error) {
	taprootAddr, err := btcutil.DecodeAddress(strAddr, cfg)
	if err != nil {
		return nil, err
	}

	byteAddr, err := txscript.PayToAddrScript(taprootAddr)
	if err != nil {
		return nil, err
	}
	return byteAddr, nil
}

func BuildP2trTransaction() {
	// 默认的 version = 1
	tx := wire.NewMsgTx(wire.TxVersion)
	point, fetcher := GetUnspent("ignore address param")
	byteAddr, err := DecodeBitcoinAddress("tb1qvu0y94au0w6hny87y88stgyux84fst6lfvser2", &chaincfg.SigNetParams)
	if err != nil {
		panic("decode address error")
	}
	// 以前一笔交易的输出点作为输入
	in := wire.NewTxIn(point, nil, nil)
	tx.AddTxIn(in)

	// 新建输出，支付到指定地址并填充转移多少
	out := wire.NewTxOut(int64(980000), byteAddr)
	tx.AddTxOut(out)

	// 获取前一笔交易
	prevOutput := fetcher.FetchPrevOutput(in.PreviousOutPoint)

	wif, err := btcutil.DecodeWIF("cSeEkntEySVvLHAjJ9PaypAUZFJacrCQSj73d8hXhykhETQ38nuu")
	if err != nil {
		panic("recover  WIF error")
	}

	// 使用私钥生成见证脚本
	witness, err := txscript.TaprootWitnessSignature(tx,
		txscript.NewTxSigHashes(tx, fetcher), 0, prevOutput.Value,
		prevOutput.PkScript, txscript.SigHashDefault, wif.PrivKey)
	if err != nil {
		panic("witness error")
	}

	// 填充输入的见证脚本
	tx.TxIn[0].Witness = witness

	// 将完成签名的交易转为 hex 形式并输出
	var signedTx bytes.Buffer
	err = tx.Serialize(&signedTx)
	if err != nil {
		panic("serialize error")
	}
	finalRawTx := hex.EncodeToString(signedTx.Bytes())

	fmt.Printf("Signed Transaction:\n %s", finalRawTx)
	// Signed Transaction: 01000000000101b4332616ee7cd8298cbfe62fed450f6deb96b071a922ba21dd6155484fd582720000000000ffffffff01200300000000000022512063bb67ea89cdaa47ef81286bff2df1c9153e1fb0f09181fd1b2eda9f9d10a0c5014011a52fdf6ccdda65359ecc9761b199e132d92bb21be059c6c5fb23e86af7152d429dde23314df0db4bcd52428acffab876b8cca1e19d2788a8382c48141b19bd00000000
	//Signed Transaction: 010000000001019ab40807b2f27d9a31be7845fe542160fd05ca4563bc3ba2fa640a9434d00ea60100000000ffffffff0120f40e0000000000160014671e42d7bc7bb57990fe21cf05a09c31ea982f5f0140b8f30e0811bad36c2ccfafd7593fc12159e8c2babc23d64d84b91bbeb22447d0f581f8936ac5bd097b2770929f8574d2d7f99325fec2e44b7439177d4961e6ff00000000
}
