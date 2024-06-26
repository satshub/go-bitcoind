// 1. Get address from Bitcoin Core and update `sendAddrStr`.
// 	    bitcoin-cli -regtest getnewaddress "" bech32m
// 2. Update `privKeyStr` as you like.
// 3. execute "go run .", and get "prev address".
// 4. Send bitcoin to "prev address".
//		bitcoin-cli -regtest -named sendtoaddress address=<prev address> amount=0.1 fee_rate=1
// 5. Get transaction information from Bitcoin Core.
//		bitcoin-cli -regtest gettransaction <previous txid>
// 6. Update `prevHashStr` and `prevIndex` from "gettransaction" result.
// 7. execute "go run .", and get "raw tx".
// 8. Send raw transaction.
//		bitcoin-cli -regtest sendrawtransaction <raw tx>

package taproot

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	tx "github.com/satshub/go-bitcoind/transaction/taproot/internal"
)

const (
	_privKeyStr = "112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00"

	// previous outpoint
	_prevHashStr   = "994e2da234734d14ec61eb95d3076d82ef2b660c026fc0f6378e585cbd3a51bc"
	_prevIndex     = uint32(1)
	_prevAmountSat = int64(10_000_000)
	_feeSat        = int64(200)

	// send address: bitcoin-cli -regtest getnewaddress "" bech32m
	_sendAddrStr = "bcrt1pypjucsfaqlfga7kxal0gfttpd95c8pe3vdexrgxjp5fh606mf09s7gvluq"
)

func KeyPath() {
	privKey, _ := hex.DecodeString(_privKeyStr)
	key := tx.NewKey(privKey, Network)

	p2tr, err := key.CreateP2TR()
	if err != nil {
		fmt.Printf("fail CreateP2TR(): %v\n", err)
		return
	}
	fmt.Printf("send to this address: %s\n\n", p2tr)

	// redeem
	prevHash, _ := chainhash.NewHashFromStr(_prevHashStr)
	rawTx, txid, err := key.CreateRawTxP2TR(prevHash, _prevIndex, _prevAmountSat, _sendAddrStr, _feeSat)
	if err != nil {
		fmt.Printf("fail CreateRawTxP2TR: %v\n", err)
		return
	}
	fmt.Printf("raw tx: %x\n", rawTx)
	fmt.Printf("txid: %s\n", txid)
}
