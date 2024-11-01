package mempool

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/satshub/go-bitcoind/common"
)

type mempoolUTXO struct {
	Txid   string `json:"txid"`
	Vout   int    `json:"vout"`
	Status struct {
		Confirmed   bool   `json:"confirmed"`
		BlockHeight int    `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		BlockTime   int64  `json:"block_time"`
	} `json:"status"`
	Value int64 `json:"value"`
}

type Utxo struct {
	TxHash   common.Hash
	Index    uint32
	Value    int64
	PkScript []byte
}

func (c *MempoolClient) GetUtxos(address string) ([]*Utxo, error) {
	res, err := c.request(http.MethodGet, fmt.Sprintf("/address/%s/utxo", address), nil, "json")
	if err != nil {
		return nil, err
	}
	//unmarshal the response
	var mutxos []mempoolUTXO
	err = json.Unmarshal(res, &mutxos)
	if err != nil {
		return nil, err
	}
	addr, err := btcutil.DecodeAddress(address, c.network)
	if err != nil {
		return nil, err
	}
	pkScript, _ := txscript.PayToAddrScript(addr)
	utxos := make([]*Utxo, len(mutxos))
	for i, mutxo := range mutxos {
		txHash, err := chainhash.NewHashFromStr(mutxo.Txid)
		if err != nil {
			return nil, err
		}
		utxos[i] = &Utxo{
			TxHash:   common.BytesToHash(txHash.CloneBytes()),
			Index:    uint32(mutxo.Vout),
			Value:    mutxo.Value,
			PkScript: pkScript,
		}
	}
	return utxos, nil
}
