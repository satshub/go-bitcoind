/**
 * Description:
 * Author: Yihen.Liu
 * Create: 2021-07-30
 */
package p2sh

import "fmt"

func _main() {
	redeemScript, redeemHash, addr, err := BuildMultiSigRedeemScript()
	if err == nil {
		fmt.Println("redeem script:", redeemScript)
		fmt.Println("redeem hash:", redeemHash)
		fmt.Println("p2sh addr:", addr)
	}

	spendHex, err := SpendMultiSig()
	if err == nil {
		fmt.Println("spend hex:", spendHex)
	}

}

func main() {
	if res, err := DisAsembleScript(); err != nil {
		fmt.Println("err:", err.Error())
	} else {
		fmt.Println("script:", res)
	}
}
