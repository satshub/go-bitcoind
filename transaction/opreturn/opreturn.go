package opreturn

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
)

const sendURL = "https://api.chain.com/v2/testnet3/transactions/send"

var (
	pkWIF              = "cMcv2Y3vDY2STEkFqsDrVryZ7dZHkL9gNExMg1jmk2BSVMizinHu"
	prevOutPkScriptStr = "76a9147af1bab2645028cd20a491b7929dec96f94d5efc88ac"
	prevOutHashStr     = "4ca3ab297341bec8603f16a747068975531339bf72469b40bc89cfd54eeb56fa"
	prevOutIndex       = uint32(0)
	changeAddressStr   = "mrj2K6txjo2QBcSmuAzHj4nD1oXSEJE1Qo"
	change             = 100000
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	wif, err := btcutil.DecodeWIF(pkWIF)
	if err != nil {
		log.Fatal(err)
	}

	prevOutHash, err := chainhash.NewHashFromStr(prevOutHashStr)
	if err != nil {
		log.Fatal(err)
	}
	prevOutPkScript, err := hex.DecodeString(prevOutPkScriptStr)
	if err != nil {
		log.Fatal(err)
	}
	changeAddress, err := btcutil.DecodeAddress(changeAddressStr, &chaincfg.TestNet3Params)
	if err != nil {
		log.Fatal(err)
	}
	sendTx(buildTxOPRETURN(wif.PrivKey, changeAddress, int64(change), prevOutHash, prevOutPkScript, "Hello from Chain."))
}

func buildTxOPRETURN(key *btcec.PrivateKey, changeAddress btcutil.Address, change int64, hash *chainhash.Hash, script []byte, data string) []byte {
	tx := wire.NewMsgTx(wire.TxVersion)

	txin := wire.NewTxIn(wire.NewOutPoint(hash, prevOutIndex), nil, nil)
	tx.AddTxIn(txin)

	b := txscript.NewScriptBuilder()
	b.AddOp(txscript.OP_RETURN)
	b.AddData([]byte(data))

	pkScript, err := txscript.PayToAddrScript(changeAddress)
	if err != nil {
		log.Fatal(err)
	}

	tx.AddTxOut(wire.NewTxOut(change, pkScript))
	opReturnScript, err := b.Script()
	tx.AddTxOut(wire.NewTxOut(0, opReturnScript))

	sig, err := txscript.SignatureScript(tx, 0, script, txscript.SigHashAll, key, true)
	if err != nil {
		log.Fatal(err)
	}
	txin.SignatureScript = sig

	var signedTxHex bytes.Buffer
	if err := tx.Serialize(&signedTxHex); err != nil {
		log.Fatal(err)
	}
	return signedTxHex.Bytes()
}

func sendTx(signedHex []byte) {
	var sendTxReq = struct {
		Hex string `json:"signed_hex"`
	}{hex.EncodeToString(signedHex)}

	log.Printf("reqBody=%s\n", sendTxReq.Hex)

	var reqBuf bytes.Buffer
	if err := json.NewEncoder(&reqBuf).Encode(sendTxReq); err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("POST", sendURL, &reqBuf)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth("GUEST-TOKEN", "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("respBody=%s\n", string(respBody))
}
