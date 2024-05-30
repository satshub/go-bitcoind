package taproot

import (
	tx "github.com/satshub/go-bitcoind/transaction/taproot/internal"
)

func _main() {
	// rt := tx.RedeemP2wpkh
	// rt := tx.RedeemP2trKeyPath
	// rt := tx.RedeemP2wsh
	rt := tx.RedeemP2trScriptPath

	switch rt {
	case tx.RedeemP2wpkh:
		P2wpkh()
	case tx.RedeemP2trKeyPath:
		KeyPath()
	case tx.RedeemP2wsh:
		P2wsh()
	case tx.RedeemP2trScriptPath:
		ScriptPath()
	}
}
