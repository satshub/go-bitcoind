/**
 * Description:
 * Author: Yihen.Liu
 * Create: 2021-07-30
 */
package p2sh

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/txscript"
)

func DisAsembleScript() (string, error) {
	// you can provide your locking script to dis asemble
	// locking script: OP_HASH160 <redeem-hash> OP_EQUAL
	lockingScript := "a914f63e2cbcc678236f683d267e7bb298ffdcd57b0487"
	script, err := hex.DecodeString(lockingScript)
	if err != nil {
		return "", err
	}
	scriptStr, err := txscript.DisasmString(script)
	if err != nil {
		return "", err
	}
	fmt.Println("lock script:", scriptStr)
	return scriptStr, nil
}
