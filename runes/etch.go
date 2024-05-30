package runes

import (
	"bytes"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/bxelab/runestone"
)

var lock sync.Mutex

func BuildEtchingDummyRevealTx(etching *runestone.Etching, feeRate int64, privateKey string, toAddr string, network *chaincfg.Params) (int64, string, error) {
	rs := runestone.Runestone{Etching: etching}
	data, err := rs.Encipher()
	if err != nil {
		//log.Error("Etching rune encipher error:", err.Error())
		return 0, "", err
	}

	//etchJson, _ := json.Marshal(etching)
	//log.Debugf("etching json content:%s", string(etchJson))

	commitment := etching.Rune.Commitment()
	prvKey, _, _ := GetPrivateKeyAddr(privateKey, network)
	var txFee int64
	var inscribeAddr string
	txFee, inscribeAddr, err = BuildRuneEtchingDummyRevealTx(prvKey, data, commitment, feeRate, 546, network, toAddr)
	if err != nil {
		return 0, "", err
	}

	//log.Debugf("inscribe address: %x\n", inscribeAddr)
	//log.Debugf("tx fee from build etching: %d\n", txFee)
	return txFee, inscribeAddr, nil
}

func BuildEtchingCompleteRevealTx(etching *runestone.Etching, feeRate int64, privateKey string, commitTxHash string, toAddr string, network *chaincfg.Params) ([]byte, int64, string, error) {
	rs := runestone.Runestone{Etching: etching}
	data, err := rs.Encipher()
	if err != nil {
		//log.Error("Etching rune encipher error:", err.Error())
		return []byte{}, 0, "", err
	}

	//etchJson, _ := json.Marshal(etching)
	//log.Debugf("etching json content:%s", string(etchJson))

	commitment := etching.Rune.Commitment()
	prvKey, _, _ := GetPrivateKeyAddr(privateKey, network)
	//utxos, err := btcConnector.GetUtxos(address)
	var rTx []byte
	var txFee int64
	var inscribeAddr string
	rTx, txFee, inscribeAddr, err = BuildRuneEtchingCompleteRevealTx(prvKey, data, commitment, feeRate, 546, network, toAddr, commitTxHash)
	if err != nil {
		return []byte{}, 0, "", err
	}

	//log.Debugf("reveal Tx: %x\n", rTx)
	//log.Debugf("inscribe address: %s\n", inscribeAddr)
	//log.Debugf("reveal tx fee about building etching: %d\n", txFee)
	return rTx, txFee, inscribeAddr, nil
}

func BuildEtchingTxs(etching *runestone.Etching, feeRate int64, privateKey string, network *chaincfg.Params) ([]byte, []byte, int64, string, error) {
	rs := runestone.Runestone{Etching: etching}
	data, err := rs.Encipher()
	if err != nil {
		//log.Error("Etching rune encipher error:", err.Error())
		return []byte{}, []byte{}, 0, "", err
	}

	//etchJson, _ := json.Marshal(etching)
	//log.Debugf("etching json content:%s", string(etchJson))

	commitment := etching.Rune.Commitment()
	btcConnector := NewMempoolConnector()
	prvKey, address, _ := GetPrivateKeyAddr(privateKey, network)
	utxos, err := btcConnector.GetUtxos(address)
	var cTx, rTx []byte
	var txFee int64
	var inscribeAddr string
	mime, logoData := "", []byte{}
	if len(mime) == 0 {
		cTx, rTx, txFee, inscribeAddr, err = BuildRuneEtchingTxs(prvKey, utxos, data, commitment, feeRate, 546, network, address)
	} else {
		cTx, rTx, txFee, err = BuildInscriptionTxs(prvKey, utxos, mime, logoData, feeRate, 546, network, commitment, data)
	}
	if err != nil {
		return []byte{}, []byte{}, 0, "", err
	}

	//log.Debugf("commit Tx: %x\n", cTx)
	//log.Debugf("reveal Tx: %x\n", rTx)
	//log.Debugf("inscribe address: %x\n", inscribeAddr)
	//log.Debugf("tx fee from build etching: %d\n", txFee)
	return cTx, rTx, txFee, inscribeAddr, nil
}

func SendTx(connector *MempoolConnector, ctx []byte, rtx []byte) error {
	tx := wire.NewMsgTx(wire.TxVersion)
	tx.Deserialize(bytes.NewReader(ctx))
	ctxHash, err := connector.SendRawTransaction(tx, false)
	if err != nil {
		//log.Error("SendRawTransaction error:", err.Error())
		return err
	}
	//log.Debug("committed tx hash:", ctxHash)
	if rtx == nil {
		return err
	}
	//log.Info("waiting for confirmations..., please don't close the program.")
	//wail ctx tx confirm
	lock.Lock()
	go func(ctxHash *chainhash.Hash) {
		for {
			time.Sleep(30 * time.Second)
			txInfo, err := connector.GetTxByHash(ctxHash.String())
			if err != nil {
				//log.Error("GetTransaction error:", err.Error())
				continue
			}
			//log.Info("commit tx confirmations:", txInfo.Confirmations)
			if txInfo.Confirmations > runestone.COMMIT_CONFIRMATIONS {
				//if txInfo.Confirmations >= CommitConfirmBlockNum() {
				break
			}
		}
		lock.Unlock()
	}(ctxHash)
	lock.Lock() //wait
	tx = wire.NewMsgTx(wire.TxVersion)
	tx.Deserialize(bytes.NewReader(rtx))
	rtxHash, err := connector.SendRawTransaction(tx, false)
	if err != nil {
		//log.Error("SendRawTransaction error:", err.Error())
		return err
	}
	println("Etch complete, reveal tx hash:", rtxHash)
	return nil
}

func SendRevealTx(connector *MempoolConnector, rtx []byte) error {
	tx := wire.NewMsgTx(wire.TxVersion)
	tx.Deserialize(bytes.NewReader(rtx))
	rtxHash, err := connector.SendRawTransaction(tx, false)
	if err != nil {
		//log.Error("SendRawTransaction error:", err.Error())
		return err
	}
	println("Etch complete, reveal tx hash:", rtxHash)
	return nil
}
