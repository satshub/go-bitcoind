package internal

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestListUnspentOfAddress(t *testing.T) {

	bitcoinRpc := BitcoinRpc{
		RpcUser:    "ideajoo",
		RpcPW:      "ideajoo123",
		RpcConnect: "127.0.0.1",
		RpcPort:    "18332",
		RpcPath:    "wallet/test_07",
	}

	result, err := bitcoinRpc.ListUnspentOfAddress(0, 0, []string{"tb1q8yu29c59hlmem3hed28f49k4f3kwwkrv4smgkh", "tb1qmhqe8pr06v0mefelardj4h6hkq095e5dh72mv3"})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	jsonString, err := json.Marshal(result)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("\n\n== result ==\n%s\n", jsonString)
}

func TestCreateRawTransaction(t *testing.T) {

	bitcoinRpc := BitcoinRpc{
		RpcUser:    "ideajoo",
		RpcPW:      "ideajoo123",
		RpcConnect: "127.0.0.1",
		RpcPort:    "18332",
	}

	inTxUnspents := make([]map[string]interface{}, 0)

	tmpUnspent := make(map[string]interface{})
	tmpUnspent["txid"] = "789cb8c274df8820b1c817bff91476661c3300bfeb3fa382c42520c9b888a1b3"
	tmpUnspent["vout"] = 1
	inTxUnspents = append(inTxUnspents, tmpUnspent)

	tmpUnspent = make(map[string]interface{})
	tmpUnspent["txid"] = "4b4c1a2f4adc0f8b1058cdb1509847b174223feeedbd83740e7edad03c1149bc"
	tmpUnspent["vout"] = 1
	inTxUnspents = append(inTxUnspents, tmpUnspent)

	outAddress := make(map[string]float64)
	outAddress["tb1q8yu29c59hlmem3hed28f49k4f3kwwkrv4smgkh"] = 0.00002000
	outAddress["tb1q3flg4mlnuk2xexu773g8d4lh6nl48rc6w6vhsm"] = 0.00003000

	outDataHex := "48454c4c4f20696465616a6f6f2f676f2d626974636f696e2d636c692d6c69676874" // "HELLO ideajoo/go-bitcoin-cli-light"

	result, err := bitcoinRpc.CreateRawTransaction(inTxUnspents, outAddress, outDataHex)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("\n\n== result rawTx ==\n%+v\n", result)
	// 020000000244199d95b6dc4eb1d6b7dc9dddf9f092751fa41ea739d3c46b32b69b9f0beab00100000000fdffffff55a4a5010bca54b6fdd507cf9850c95142a2fab14db7ec7530b2bba76f6579980100000000fdffffff020000000000000000246a2248454c4c4f20696465616a6f6f2f676f2d626974636f696e2d636c692d6c69676874c05d0000000000001600143938a2e285bff79dc6f96a8e9a96d54c6ce7586c00000000
}

func TestDumpPrivateKey(t *testing.T) {
	bitcoinRpc := BitcoinRpc{
		RpcUser:    "ideajoo",
		RpcPW:      "ideajoo123",
		RpcConnect: "127.0.0.1",
		RpcPort:    "18332",
	}
	tAddress := "tb1q8yu29c59hlmem3hed28f49k4f3kwwkrv4smgkh"
	resultPrivKey, err := bitcoinRpc.DumpPrivateKey(tAddress)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("\n\n== result ==\n%s\n", resultPrivKey)
	// cQLN8Z38G7MJk82JMFbuQcXSfQGHeZKshWJ4haSmnb9AxX9Et4Vy
}

func TestSignRawTransactionWithKey(t *testing.T) {
	bitcoinRpc := BitcoinRpc{
		RpcUser:    "ideajoo",
		RpcPW:      "ideajoo123",
		RpcConnect: "127.0.0.1",
		RpcPort:    "18332",
	}
	tRawTx := "020000000244199d95b6dc4eb1d6b7dc9dddf9f092751fa41ea739d3c46b32b69b9f0beab00100000000fdffffff55a4a5010bca54b6fdd507cf9850c95142a2fab14db7ec7530b2bba76f6579980100000000fdffffff020000000000000000246a2248454c4c4f20696465616a6f6f2f676f2d626974636f696e2d636c692d6c69676874c05d0000000000001600143938a2e285bff79dc6f96a8e9a96d54c6ce7586c00000000"
	tPrivKey := "cQLN8Z38G7MJk82JMFbuQcXSfQGHeZKshWJ4haSmnb9AxX9Et4Vy"
	resultSignedRawTx, err := bitcoinRpc.SignRawTransactionWithKey(tRawTx, tPrivKey)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("\n\n== result ==\n%s\n", resultSignedRawTx)
	// 0200000000010244199d95b6dc4eb1d6b7dc9dddf9f092751fa41ea739d3c46b32b69b9f0beab00100000000fdffffff55a4a5010bca54b6fdd507cf9850c95142a2fab14db7ec7530b2bba76f6579980100000000fdffffff020000000000000000246a2248454c4c4f20696465616a6f6f2f676f2d626974636f696e2d636c692d6c69676874c05d0000000000001600143938a2e285bff79dc6f96a8e9a96d54c6ce7586c02473044022032b8e51b0e6be0846f2bd458919e3dad85d3923afce20ff6c3494a63eb88014002204c136999d2a60f23e12bbaa5f5a1e9e0704c00441defd77bb7c55ce86a538f4c01210307fb2416e1477f965dfee36f9525b0642759b22c23430ebe9a63124d62634b520247304402203800d79251b9eaf995549ee9c64c43a46fa33071a43dcac76f6d9328e67e2177022008ec36403a89ec8bbbea163932bd4048c93431336a6c0f94db6c749b631304ee01210307fb2416e1477f965dfee36f9525b0642759b22c23430ebe9a63124d62634b5200000000
}

func TestSendRawTransaction(t *testing.T) {
	bitcoinRpc := BitcoinRpc{
		RpcUser:    "ideajoo",
		RpcPW:      "ideajoo123",
		RpcConnect: "127.0.0.1",
		RpcPort:    "18332",
	}
	tSignedRawTx := "0200000000010244199d95b6dc4eb1d6b7dc9dddf9f092751fa41ea739d3c46b32b69b9f0beab00100000000fdffffff55a4a5010bca54b6fdd507cf9850c95142a2fab14db7ec7530b2bba76f6579980100000000fdffffff020000000000000000246a2248454c4c4f20696465616a6f6f2f676f2d626974636f696e2d636c692d6c69676874c05d0000000000001600143938a2e285bff79dc6f96a8e9a96d54c6ce7586c02473044022032b8e51b0e6be0846f2bd458919e3dad85d3923afce20ff6c3494a63eb88014002204c136999d2a60f23e12bbaa5f5a1e9e0704c00441defd77bb7c55ce86a538f4c01210307fb2416e1477f965dfee36f9525b0642759b22c23430ebe9a63124d62634b520247304402203800d79251b9eaf995549ee9c64c43a46fa33071a43dcac76f6d9328e67e2177022008ec36403a89ec8bbbea163932bd4048c93431336a6c0f94db6c749b631304ee01210307fb2416e1477f965dfee36f9525b0642759b22c23430ebe9a63124d62634b5200000000"
	resultTxID, err := bitcoinRpc.SendRawTransaction(tSignedRawTx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("\n\n== result ==\n%s\n", resultTxID)
	// fb92e4a2aab9e55f11dfe3bf047a8d37fde0b274e99cee08db943f12f6975788
}

func TestGetBlockCount(t *testing.T) {
	bitcoinRpc := BitcoinRpc{
		RpcUser:    "ideajoo",
		RpcPW:      "ideajoo123",
		RpcConnect: "127.0.0.1",
		RpcPort:    "18332",
	}
	resultBlockCount, err := bitcoinRpc.GetBlockCount()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("\n\n== result ==\n%d\n", resultBlockCount)
	// 2344981
}

func TestGetBlockHash(t *testing.T) {
	bitcoinRpc := BitcoinRpc{
		RpcUser:    "ideajoo",
		RpcPW:      "ideajoo123",
		RpcConnect: "127.0.0.1",
		RpcPort:    "18332",
	}
	resultBlockHash, err := bitcoinRpc.GetBlockHash(2344981)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("\n\n== result ==\n%s\n", resultBlockHash)
	// 000000000000e7f3e8f60f9431725df65cdeb5c13386f03edaba73269e2d313d
}

func TestGetBlock(t *testing.T) {
	bitcoinRpc := BitcoinRpc{
		RpcUser:    "ideajoo",
		RpcPW:      "ideajoo123",
		RpcConnect: "127.0.0.1",
		RpcPort:    "18332",
	}
	resultBlock, err := bitcoinRpc.GetBlock("000000000000e7f3e8f60f9431725df65cdeb5c13386f03edaba73269e2d313d")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("\n\n== result ==\n%+v\n", resultBlock["tx"])
	for _, txid := range resultBlock["tx"].([]string) {
		fmt.Printf("\n== txid :%s", txid)
	}
}

func TestGetRawTransaction(t *testing.T) {
	bitcoinRpc := BitcoinRpc{
		RpcUser:    "ideajoo",
		RpcPW:      "ideajoo123",
		RpcConnect: "127.0.0.1",
		RpcPort:    "18332",
	}
	result, err := bitcoinRpc.GetRawTransaction("45c0f0a58d4f356b605a26b8aceb3faab24cf067c7d084b252d4ff863c692771")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	jsonString, err := json.Marshal(result)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("\n\n== result ==\n%s\n", jsonString)
}

func TestGetNewAddress(t *testing.T) {

	bitcoinRpc := BitcoinRpc{
		RpcUser:    "ideajoo",
		RpcPW:      "ideajoo123",
		RpcConnect: "127.0.0.1",
		RpcPort:    "18332",
	}

	tWallet := "test"
	resultNewAddress, err := bitcoinRpc.GetNewAddress(tWallet, "", "")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("\n\n== result ==\n%s\n", resultNewAddress)
	// tb1qa6v5vvpagj7lqnummqff0jm086y3vq3jjc9r90
}

func TestListReceivedByAddress(t *testing.T) {
	bitcoinRpc := BitcoinRpc{
		RpcUser:    "ideajoo",
		RpcPW:      "ideajoo123",
		RpcConnect: "127.0.0.1",
		RpcPort:    "18332",
	}

	tWallet := "test"
	resultListReceivedByAddress, err := bitcoinRpc.ListReceivedByAddress(tWallet, 1, true, true, "")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	jsonString, err := json.Marshal(resultListReceivedByAddress)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("\n\n== result ==\n%s\n", jsonString)

}
