package mempool

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func TestGetRawTransaction(t *testing.T) {
	//https://mempool.space/signet/api/tx/b752d80e97196582fd02303f76b4b886c222070323fb7ccd425f6c89f5445f6c/hex
	client := NewClient(&chaincfg.SigNetParams)
	txId, err := chainhash.NewHashFromStr("1e82e3cc1580dcbbd5798102a96aecfa7d836f0baaa860a826c1154bf51de50a")
	if err != nil {
		t.Error(err)
	}
	transaction, err := client.GetRawTransaction(txId)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(transaction.TxHash().String())
	}
}
