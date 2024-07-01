package mempool

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func TestGetRawTransaction(t *testing.T) {
	//https://mempool.space/signet/api/tx/b752d80e97196582fd02303f76b4b886c222070323fb7ccd425f6c89f5445f6c/hex
	client := NewClient(&chaincfg.MainNetParams)
	txId, _ := chainhash.NewHashFromStr("ad6c6a33a7f6f8784d9982b69ec1c048a82f2e8dce0b190356bd99ead68ac93d")
	transaction, err := client.GetRawTransaction(txId)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(transaction.TxHash().String())
	}
}
