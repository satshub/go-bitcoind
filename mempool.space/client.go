package mempool

import (
	"io"
	"log"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
)

type MempoolClient struct {
	baseURL string
	network *chaincfg.Params
}

func NewClient(netParams *chaincfg.Params) *MempoolClient {
	baseURL := ""
	if netParams.Net == wire.MainNet {
		baseURL = "https://mempool.space/api"
	} else if netParams.Net == wire.TestNet3 {
		baseURL = "https://mempool.space/testnet/api"
	} else if netParams.Net == chaincfg.SigNetParams.Net {
		//baseURL = "https://mempool.space/signet/api"
		baseURL = "https://signet.sathub.io/api"
	} else {
		log.Fatal("mempool don't support other netParams")
	}
	return &MempoolClient{
		baseURL: baseURL,
		network: netParams,
	}
}

func (c *MempoolClient) request(method, subPath string, requestBody io.Reader) ([]byte, error) {
	return Request(method, c.baseURL, subPath, requestBody)
}

var _ BTCAPIClient = (*MempoolClient)(nil)
