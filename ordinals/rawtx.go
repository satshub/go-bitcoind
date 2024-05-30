package ordinals

// btcApiClient := mempool.NewClient(setting.NetworkParams)
/*
func SendHexTransaction(hexTx string, apiClient *mempool.MempoolClient) (string, error) {
	// Decode the serialized transaction hex to raw bytes.
	serializedTx, err := hex.DecodeString(hexTx)
	if err != nil {
		return "", errors.New("serialize tx err:" + err.Error())
	}

	// Deserialize the transaction and return it.
	var msgTx wire.MsgTx
	if err := msgTx.Deserialize(bytes.NewReader(serializedTx)); err != nil {
		return "", errors.New("deserialize tx err:" + err.Error())
	}

	txHash, err := apiClient.BroadcastTx(&msgTx)
	return txHash.String(), err
}

*/
