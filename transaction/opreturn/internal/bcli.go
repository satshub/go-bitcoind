package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
)

type BitcoinRpc struct {
	RpcUser    string
	RpcPW      string
	RpcConnect string
	RpcPort    string
	RpcPath    string
}

func defaultJsonRpcInfo() (info map[string]interface{}) {
	info = make(map[string]interface{})
	info["jsonrpc"] = "1.0"
	info["id"] = "GoBitcoinCliLight"
	return
}

func (bitcoinRpc BitcoinRpc) request(jsonRpcBytes []byte) (body []byte, err error) {

	request, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%s/%s", bitcoinRpc.RpcConnect, bitcoinRpc.RpcPort, bitcoinRpc.RpcPath), bytes.NewBuffer(jsonRpcBytes))
	if err != nil {
		err = fmt.Errorf("@http.NewRequest('POST', ...): %v", err)
		return
	}
	request.Header.Set("content-type", "text/plain;")
	request.SetBasicAuth(bitcoinRpc.RpcUser, bitcoinRpc.RpcPW)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		err = fmt.Errorf("@client.Do(request): %v", err)
		return
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("@io.ReadAll(resp.Body): %v", err)
		return
	}
	return
}

func (bitcoinRpc BitcoinRpc) ListUnspentOfAddress(minconf int, maxconf int, addresses []string) (result []map[string]interface{}, err error) {

	if minconf <= 1 || minconf >= 9999999 {
		minconf = 1 // Default
	}
	if maxconf <= 1 || maxconf >= 9999999 {
		maxconf = 9999999 // Default
	}

	result = make([]map[string]interface{}, 0)
	jsonRpcInfo := defaultJsonRpcInfo()
	jsonRpcInfo["method"] = "listunspent"
	jsonRpcInfo["params"] = []interface{}{minconf, maxconf, addresses}
	jsonRpcBytes, err := json.Marshal(jsonRpcInfo)
	if err != nil {
		err = fmt.Errorf("@json.Marshal(jsonRpcInfo): %v. %+v", err, jsonRpcInfo)
		return
	}

	body, err := bitcoinRpc.request(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@bitcoinRpc.request(jsonRpcBytes): %v", err)
		return
	}

	type listUnspentInfo struct {
		TxID          string   `json:"txid"`          // (string) the transaction id
		Vout          int      `json:"vout"`          // (numeric) the vout value
		Address       string   `json:"address"`       // (string) the bitcoin address
		Label         string   `json:"label"`         // (string) The associated label, or "" for the default label
		ScriptPutKey  string   `json:"scriptPubKey"`  // (string) the script key
		Amount        float64  `json:"amount"`        // (numeric) the transaction output amount in BTC
		Confirmations int      `json:"confirmations"` // (numeric) The number of confirmations
		RedeemScript  string   `json:"redeemScript"`  // (string) The redeemScript if scriptPubKey is P2SH
		WitnessScript string   `json:"witnessScript"` // (string) witnessScript if the scriptPubKey is P2WSH or P2SH-P2WSH
		Spendable     bool     `json:"spendable"`     // (boolean) Whether we have the private keys to spend this output
		Solvable      bool     `json:"solvable"`      // (boolean) Whether we know how to spend this output, ignoring the lack of keys
		Reused        bool     `json:"reused"`        // (boolean) (only present if avoid_reuse is set) Whether this output is reused/dirty (sent to an address that was previously spent from)
		Desc          string   `json:"desc"`          // (string) (only when solvable) A descriptor for spending this output
		ParentDescs   []string `json:"parent_descs"`  //
		Safe          bool     `json:"safe"`          // (boolean) Whether this output is considered safe to spend. Unconfirmed transactions
	}
	type resultListUnspent struct {
		ListUnspents []listUnspentInfo `json:"result"`
	}
	bodyResult := resultListUnspent{}
	err = json.Unmarshal(body, &bodyResult)
	if err != nil {
		err = fmt.Errorf("@json.Unmarshal(body, &bodyResult): %s", err)
		return
	}

	for _, listUnspent := range bodyResult.ListUnspents {
		tmpMap := make(map[string]interface{})
		inrec, errInner := json.Marshal(listUnspent)
		if errInner != nil {
			err = fmt.Errorf("@json.Marshal(listUnspent): %v", err)
			return
		}
		err = json.Unmarshal(inrec, &tmpMap)
		if err != nil {
			err = fmt.Errorf("@json.Unmarshal(inrec, &tmpMap): %v", err)
			return
		}
		result = append(result, tmpMap)
	}

	return
}

func (bitcoinRpc BitcoinRpc) CreateRawTransaction(inTxUnspents []map[string]interface{}, outAddresses map[string]float64, outDataHex string) (rawTx string, err error) {

	tCreateTxOuts := make([]map[string]interface{}, 0)

	// outParamsAddressAmount
	for outAddress, outAmount := range outAddresses {
		tParamsAddress := make(map[string]interface{})
		outAmount = math.Round((outAmount)*100000000) / 100000000
		if outAmount < 0.00000000 {
			// Only Filtering when minus-amount
			// Zero-amount is needed sometimes
			// Zero-amount will be controlled on service
			continue
		}
		tParamsAddress[outAddress] = outAmount
		tCreateTxOuts = append(tCreateTxOuts, tParamsAddress)
	}

	// outParamsData
	if outDataHex != "" {
		tCreateTxOuts = append(tCreateTxOuts, map[string]interface{}{"data": outDataHex})
	}

	if len(tCreateTxOuts) == 0 {
		err = fmt.Errorf("len(tCreateTxOuts) == 0 : incorrect outAddresses and outDataHex")
		return
	}

	jsonRpcInfo := defaultJsonRpcInfo()
	jsonRpcInfo["method"] = "createrawtransaction"
	jsonRpcInfo["params"] = []interface{}{inTxUnspents, tCreateTxOuts}
	jsonRpcBytes, err := json.Marshal(jsonRpcInfo)
	if err != nil {
		err = fmt.Errorf("@json.Marshal(jsonRpcInfo): %v", err)
		return
	}

	body, err := bitcoinRpc.request(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@bitcoinRpc.request(jsonRpcBytes): %v", err)
		return
	}

	type resultCreateRaxTx struct {
		RawTx string `json:"result"`
	}
	result := resultCreateRaxTx{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = fmt.Errorf("@json.Unmarshal(body, &result): %v", err)
		return
	}
	rawTx = result.RawTx

	if rawTx == "" {
		err = fmt.Errorf("rawTx == '': rawTx of result is empty")
		return
	}

	return
}

func (bitcoinRpc BitcoinRpc) DumpPrivateKey(address string) (privKey string, err error) {

	jsonRpcInfo := defaultJsonRpcInfo()
	jsonRpcInfo["method"] = "dumpprivkey"
	jsonRpcInfo["params"] = []interface{}{address}
	jsonRpcBytes, err := json.Marshal(jsonRpcInfo)
	if err != nil {
		err = fmt.Errorf("@json.Marshal(jsonRpcInfo): %v", err)
		return
	}

	body, err := bitcoinRpc.request(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@bitcoinRpc.request(jsonRpcBytes): %v", err)
		return
	}

	type resultDumpPrivateKey struct {
		PrivKey string `json:"result"`
	}
	result := resultDumpPrivateKey{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = fmt.Errorf("@json.Unmarshal(body, &result): %v", err)
		return
	}

	privKey = result.PrivKey

	return
}

func (bitcoinRpc BitcoinRpc) SignRawTransactionWithKey(rawTx string, privKey string) (signedRawTx string, err error) {

	jsonRpcInfo := defaultJsonRpcInfo()
	jsonRpcInfo["method"] = "signrawtransactionwithkey"
	jsonRpcInfo["params"] = []interface{}{rawTx, []string{privKey}}
	jsonRpcBytes, err := json.Marshal(jsonRpcInfo)
	if err != nil {
		err = fmt.Errorf("@json.Marshal(jsonRpcInfo): %v", err)
		return
	}

	body, err := bitcoinRpc.request(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@bitcoinRpc.request(jsonRpcBytes): %v", err)
		return
	}

	type signedRawTxInfo struct {
		Hex      string `json:"hex"`      // (string) the transaction id
		Complete bool   `json:"complete"` // (numeric) the vout value
	}
	type resultSignedRawTxInfo struct {
		SignedRawTxInfo signedRawTxInfo `json:"result"`
	}

	// tmpBody := string(body)
	// fmt.Printf("%s", tmpBody)

	result := resultSignedRawTxInfo{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = fmt.Errorf("@json.Unmarshal(body, &result): %v", err)
		return
	}

	signedRawTx = result.SignedRawTxInfo.Hex

	return
}

func (bitcoinRpc BitcoinRpc) SendRawTransaction(signedRawTx string) (txID string, err error) {

	jsonRpcInfo := defaultJsonRpcInfo()
	jsonRpcInfo["method"] = "sendrawtransaction"
	jsonRpcInfo["params"] = []interface{}{signedRawTx}
	jsonRpcBytes, err := json.Marshal(jsonRpcInfo)
	if err != nil {
		err = fmt.Errorf("@json.Marshal(jsonRpcInfo): %v", err)
		return
	}

	body, err := bitcoinRpc.request(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@bitcoinRpc.request(jsonRpcBytes): %v", err)
		return
	}

	type resultSendRawTx struct {
		TxID string `json:"result"`
	}
	result := resultSendRawTx{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = fmt.Errorf("@json.Unmarshal(body, &result): %v", err)
		return
	}

	txID = result.TxID
	return
}

func (bitcoinRpc BitcoinRpc) GetBlockCount() (blockCount int64, err error) {

	jsonRpcInfo := defaultJsonRpcInfo()
	jsonRpcInfo["method"] = "getblockcount"
	jsonRpcInfo["params"] = []interface{}{}
	jsonRpcBytes, err := json.Marshal(jsonRpcInfo)
	if err != nil {
		err = fmt.Errorf("@json.Marshal(jsonRpcInfo): %v", err)
		return
	}

	body, err := bitcoinRpc.request(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@bitcoinRpc.request(jsonRpcBytes): %v", err)
		return
	}

	type resultGetBlockCount struct {
		BlockCount int64 `json:"result"`
	}
	result := resultGetBlockCount{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = fmt.Errorf("@json.Unmarshal(body, &result): %v", err)
		return
	}

	blockCount = result.BlockCount
	return
}

func (bitcoinRpc BitcoinRpc) GetBlockHash(blockNumber int64) (blockHash string, err error) {

	jsonRpcInfo := defaultJsonRpcInfo()
	jsonRpcInfo["method"] = "getblockhash"
	jsonRpcInfo["params"] = []interface{}{blockNumber}
	jsonRpcBytes, err := json.Marshal(jsonRpcInfo)
	if err != nil {
		err = fmt.Errorf("@json.Marshal(jsonRpcInfo): %v", err)
		return
	}

	body, err := bitcoinRpc.request(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@bitcoinRpc.request(jsonRpcBytes): %v", err)
		return
	}

	type resultGetBlockHash struct {
		BlockHash string `json:"result"`
	}
	result := resultGetBlockHash{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = fmt.Errorf("@json.Unmarshal(body, &result): %v", err)
		return
	}

	blockHash = result.BlockHash
	return
}

func (bitcoinRpc BitcoinRpc) GetBlock(blockHash string) (block map[string]interface{}, err error) {

	jsonRpcInfo := defaultJsonRpcInfo()
	jsonRpcInfo["method"] = "getblock"
	jsonRpcInfo["params"] = []interface{}{blockHash}
	jsonRpcBytes, err := json.Marshal(jsonRpcInfo)
	if err != nil {
		err = fmt.Errorf("@json.Marshal(jsonRpcInfo): %v", err)
		return
	}

	body, err := bitcoinRpc.request(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@bitcoinRpc.request(jsonRpcBytes): %v", err)
		return
	}

	type BlockInfo struct {
		Hash   string   `json:"hash"`   // (string) the transaction id
		Height int64    `json:"height"` // (numeric) the vout value
		Time   int64    `json:"time"`   // (string) the bitcoin address
		Tx     []string `json:"tx"`     // (string) The associated label, or "" for the default label
		NTx    int64    `json:"nTx"`    // (string) the script key
	}

	type resultGetBlock struct {
		Block BlockInfo `json:"result"`
	}
	result := resultGetBlock{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = fmt.Errorf("@json.Unmarshal(body, &result): %v", err)
		return
	}

	block = make(map[string]interface{})
	block["hash"] = result.Block.Hash
	block["height"] = result.Block.Height
	block["time"] = result.Block.Time
	block["tx"] = result.Block.Tx
	block["nTx"] = result.Block.NTx

	return
}

func (bitcoinRpc BitcoinRpc) GetRawTransaction(txID string) (rawTxInfo map[string]interface{}, err error) {

	rawTxInfo = make(map[string]interface{})

	jsonRpcInfo := defaultJsonRpcInfo()
	jsonRpcInfo["method"] = "getrawtransaction"
	jsonRpcInfo["params"] = []interface{}{txID, true}
	jsonRpcBytes, err := json.Marshal(jsonRpcInfo)
	if err != nil {
		err = fmt.Errorf("@json.Marshal(jsonRpcInfo): %v", err)
		return
	}

	body, err := bitcoinRpc.request(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@bitcoinRpc.request(jsonRpcBytes): %v", err)
		return
	}

	type vin struct {
		TxID string `json:"txid"`
		Vout int    `json:"vout"`
	}

	type scriptPubKey struct {
		Asm     string `json:"asm"`
		Address string `json:"address"`
	}
	type vout struct {
		Value        float64      `json:"value"`
		N            int          `json:"n"`
		ScriptPubKey scriptPubKey `json:"scriptPubKey"`
	}

	type raxTxInfo struct {
		InActiveChain bool   `json:"in_active_chain"` // (boolean) Whether specified block is in the active chain or not (only present with explicit "blockhash" argument)
		Hex           string `json:"hex"`             // (string) The transaction hash (differs from txid for witness transactions)
		TxID          string `json:"txid"`            // (string) The transaction id (same as provided)
		Hash          string `json:"hash"`            // (string) The transaction hash (differs from txid for witness transactions)
		Size          int64  `json:"size"`            // (numeric) The serialized transaction size
		LockTime      int64  `json:"locktime"`        // (numeric) The lock time
		Vin           []vin  `json:"vin"`
		Vout          []vout `json:"vout"`
		BlockHash     string `json:"blockhash"`     // (string) the block hash
		Confirmations int    `json:"confirmations"` // (numeric) The confirmations
		BlockTime     int64  `json:"blocktime"`     // (numeric) The block time expressed in UNIX epoch time
		Time          int64  `json:"time"`          // (numeric) Same as "blocktime"
	}

	type resultGetRawTx struct {
		RaxTx raxTxInfo `json:"result"`
	}

	result := resultGetRawTx{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = fmt.Errorf("@json.Unmarshal(body, &result): %v", err)
		return
	}

	tVins := make([]map[string]interface{}, 0)
	for _, tRawVin := range result.RaxTx.Vin {
		tVin := make(map[string]interface{})
		tVin["txid"] = tRawVin.TxID
		tVin["vout"] = tRawVin.Vout
		// tVin["address"] = ""
		tVins = append(tVins, tVin)
	}

	tVouts := make([]map[string]interface{}, 0)
	for _, tRawVout := range result.RaxTx.Vout {
		tVout := make(map[string]interface{})
		tVout["address"] = tRawVout.ScriptPubKey.Address
		tVout["n"] = tRawVout.N
		tVout["value"] = tRawVout.Value
		tVout["scriptPubKey"] = tRawVout.ScriptPubKey
		tVouts = append(tVouts, tVout)
	}

	rawTxInfo["in_active_chain"] = result.RaxTx.InActiveChain
	rawTxInfo["txid"] = result.RaxTx.TxID
	rawTxInfo["hex"] = result.RaxTx.Hex
	rawTxInfo["hash"] = result.RaxTx.Hash
	rawTxInfo["size"] = result.RaxTx.Size
	rawTxInfo["locktime"] = result.RaxTx.LockTime
	rawTxInfo["vin"] = tVins
	rawTxInfo["vout"] = tVouts
	rawTxInfo["blockhash"] = result.RaxTx.BlockHash
	rawTxInfo["confirmations"] = result.RaxTx.Confirmations
	rawTxInfo["blocktime"] = result.RaxTx.BlockTime
	rawTxInfo["time"] = result.RaxTx.Time

	return
}

func (bitcoinRpc BitcoinRpc) GetNewAddress(walletName string, label string, addressType string) (newAddress string, err error) {

	bitcoinRpc.RpcPath = fmt.Sprintf("wallet/%s", walletName)
	jsonRpcInfo := defaultJsonRpcInfo()
	params := make([]string, 0)
	params = append(params, label)
	switch addressType {
	case "":
		break
	case "legacy", "p2sh-segwit", "bech32":
		params = append(params, addressType)
	default:
		err = fmt.Errorf("incorrect addressType[%s]", addressType)
		return
	}
	jsonRpcInfo["method"] = "getnewaddress"
	jsonRpcInfo["params"] = params
	jsonRpcBytes, err := json.Marshal(jsonRpcInfo)
	if err != nil {
		err = fmt.Errorf("@json.Marshal(jsonRpcInfo): %v", err)
		return
	}

	body, err := bitcoinRpc.request(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@bitcoinRpc.request(jsonRpcBytes): %v", err)
		return
	}

	type resultGetNewAddress struct {
		NewAddress string `json:"result"`
	}
	result := resultGetNewAddress{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = fmt.Errorf("@json.Unmarshal(body, &result): %v", err)
		return
	}

	newAddress = result.NewAddress

	return
}

func (bitcoinRpc BitcoinRpc) ListReceivedByAddress(walletName string, minconf int, includeEmpty bool, includeWatchonly bool, addressFilter string) (results []map[string]interface{}, err error) {

	bitcoinRpc.RpcPath = fmt.Sprintf("wallet/%s", walletName)

	jsonRpcInfo := defaultJsonRpcInfo()
	jsonRpcInfo["method"] = "listreceivedbyaddress"
	jsonRpcInfo["params"] = []interface{}{minconf, includeEmpty, includeWatchonly, addressFilter}
	jsonRpcBytes, err := json.Marshal(jsonRpcInfo)
	if err != nil {
		err = fmt.Errorf("@json.Marshal(jsonRpcInfo): %v", err)
		return
	}

	body, err := bitcoinRpc.request(jsonRpcBytes)
	if err != nil {
		err = fmt.Errorf("@bitcoinRpc.request(jsonRpcBytes): %v", err)
		return
	}

	type ListReceivedByAddressInfo struct {
		InvolvesWatchOnly bool     `json:"involvesWatchonly"`
		Address           string   `json:"address"`
		Amount            float64  `json:"amount"`
		Confirmations     int64    `json:"confirmations"`
		Label             string   `json:"label"`
		TxIDs             []string `json:"txids"`
	}
	type ResultListReceivedByAddress struct {
		Infos []ListReceivedByAddressInfo `json:"result"`
	}

	bodyResult := ResultListReceivedByAddress{}
	err = json.Unmarshal(body, &bodyResult)
	if err != nil {
		err = fmt.Errorf("@json.Unmarshal(body, &bodyResult): %v", err)
		return
	}

	results = make([]map[string]interface{}, 0)
	for _, info := range bodyResult.Infos {
		tResult := make(map[string]interface{})
		tResult["address"] = info.Address
		tResult["amount"] = info.Amount
		tResult["confirmations"] = info.Confirmations
		tResult["involvesWatchonly"] = info.InvolvesWatchOnly
		tResult["label"] = info.Label
		tResult["txids"] = info.TxIDs
		results = append(results, tResult)
	}

	return
}
