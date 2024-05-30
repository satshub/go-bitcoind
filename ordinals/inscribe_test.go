package ordinals

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/pkg/errors"
	"github.com/satshub/go-bitcoind/jsonrpc"
	memPool "github.com/satshub/go-bitcoind/mempool.space"
)

func TestBatchInscribe(t *testing.T) {
	//unisat 测试网p2tr地址的私钥，p2tr: tb1pst5dyk7tdymccy0xydcyyzvqgz2022t8576sjz2l65fzulrvy5rqcv6les
	privateKeyHex := "a868774f27a34e28aef14a95e2ddfa9baf2bc9a83b632b111e8e9d1eb5fbb6e9"
	destinations := []string{
		"tb1qg9hl3ulg20hel6aen5dtmhzhprjee039heu5hj",
		"2ND3jnE3N6iCdzHf77SY44LinxRx5Vg7zDs",
		"mzHchNdivKvhLRLww3VLC37VqwEHPMN3ak",
	}
	fileList := []string{
		"1.txt",
		"2.txt",
		"3.txt",
	}

	PrepareBatchIssued(1, destinations, fileList, privateKeyHex, 1, 1, &chaincfg.TestNet3Params)
}

// ////////////////////////////////////////////////////////////////////////////////////////
//
//					------>----- Middleman-0 ------->-------
//				   /                                        \
//				  /                 ......					 \
//				 /											  \
//				/											   \
//	   Sender  ------->------- Middleman-m ------->-------- Receiver
//	         \											   /
//	          \				    ......					  /
//	           \											 /
//	            \----->------- Middleman-n ------>------ /
//
// ////////////////////////////////////////////////////////////////////////////////////////
// RBF(replace-by-fee) 在 bip125 中规范,
// 只要存在 TxIn 成员中的 nSequence 字段值小于 0xffffffff - 1 那么就可以进行替换交易
//var defaultSequenceNum = wire.MaxTxInSequenceNum - 10
/*
func sequenceNum(enableRBF bool) uint32 {
	if enableRBF {
		return defaultSequenceNum
	} else {
		return wire.MaxTxInSequenceNum - 1
	}
}
*/
type CoinReceiverData struct {
	Destination string
	Amount      int64
}

type SenderRequest struct {
	CommitTxOutPointList   []*wire.OutPoint
	CommitTxPrivateKeyList []*btcec.PrivateKey // If used without RPC,
	// a local signature is required for committing the commit tx.
	// Currently, CommitTxPrivateKeyList[i] sign CommitTxOutPointList[i]
	CommitFeeRate      int64
	FeeRate            int64
	DataList           []CoinReceiverData
	SingleRevealTxOnly bool // Currently, the official Ordinal parser can only parse a single NFT per transaction.
	// When the official Ordinal parser supports parsing multiple NFTs in the future, we can consider using a single reveal transaction.
	RevealOutValue    int64
	ChangeAddress     string //找零地址; 如果不设置，将使用第一个utxo的所对应的锁定脚本；
	EnableRBF         bool
	OnlyOneRevealAddr bool
}

type taprootAccount struct {
	privateKey            *btcec.PrivateKey
	address               string
	recoveryPrivateKeyWIF string
}

type senderTxCtxData struct {
	privateKey              *btcec.PrivateKey
	senderScript            []byte
	commitTxAddressPkScript []byte
	controlBlockWitness     []byte
	commitTxAddress         string
	recoveryPrivateKeyWIF   string
	revealTxPrevOutput      *wire.TxOut
}

/*
	type blockchainClient struct {
		rpcClient    *rpcclient.Client
		btcApiClient btcapi.BTCAPIClient
	}
*/

type SenderToolBox struct {
	net                       *chaincfg.Params
	client                    *blockchainClient
	commitTxPrevOutputFetcher *txscript.MultiPrevOutFetcher
	commitTxPrivateKeyList    []*btcec.PrivateKey
	txCtxDataList             []*senderTxCtxData
	revealTxPrevOutputFetcher *txscript.MultiPrevOutFetcher
	revealTx                  []*wire.MsgTx
	commitTx                  *wire.MsgTx
	revealPrivateKey          *btcec.PrivateKey
}

const (
//defaultSequenceNum    = wire.MaxTxInSequenceNum - 10
//defaultRevealOutValue = int64(500) // 500 sat, ord default 10000

// MaxStandardTxWeight = blockchain.MaxBlockWeight / 10
)

func NewSenderToolBox(net *chaincfg.Params, rpcclient *rpcclient.Client, request *SenderRequest) (*SenderToolBox, error) {
	tool := &SenderToolBox{
		net: net,
		client: &blockchainClient{
			rpcClient: rpcclient,
		},
		commitTxPrevOutputFetcher: txscript.NewMultiPrevOutFetcher(nil),
		txCtxDataList:             make([]*senderTxCtxData, len(request.DataList)),
		revealTxPrevOutputFetcher: txscript.NewMultiPrevOutFetcher(nil),
	}
	return tool, tool.initToolBox(net, request)
}

func NewSenderToolBoxWithBtcApiClient(net *chaincfg.Params, btcApiClient memPool.BTCAPIClient, request *SenderRequest) (*SenderToolBox, error) {
	if len(request.CommitTxPrivateKeyList) != len(request.CommitTxOutPointList) {
		return nil, errors.New("the length of CommitTxPrivateKeyList and CommitTxOutPointList should be the same")
	}
	tool := &SenderToolBox{
		net: net,
		client: &blockchainClient{
			btcApiClient: btcApiClient,
		},
		commitTxPrevOutputFetcher: txscript.NewMultiPrevOutFetcher(nil),
		commitTxPrivateKeyList:    request.CommitTxPrivateKeyList,
		revealTxPrevOutputFetcher: txscript.NewMultiPrevOutFetcher(nil),
	}

	if request.OnlyOneRevealAddr {
		privateKey, err := btcec.NewPrivateKey()
		if err != nil {
			return nil, errors.New("create taproot private err")
		}
		tool.revealPrivateKey = privateKey
	}

	return tool, tool.initToolBox(net, request)
}

func (tool *SenderToolBox) transmitAddress(i int) string {
	return tool.txCtxDataList[i].commitTxAddress
}

func (tool *SenderToolBox) initToolBox(net *chaincfg.Params, request *SenderRequest) error {
	revealOutValue := defaultRevealOutValue
	if request.RevealOutValue > 0 {
		revealOutValue = request.RevealOutValue
	}
	tool.txCtxDataList = make([]*senderTxCtxData, len(request.DataList))
	destinations := make([]string, len(request.DataList))
	amounts := make([]int64, len(request.DataList))
	for i := 0; i < len(request.DataList); i++ {
		txCtxData, err := tool.CreateSenderTxCtxData(net, request.DataList[i], request.OnlyOneRevealAddr)
		if err != nil {
			return err
		}
		tool.txCtxDataList[i] = txCtxData
		destinations[i] = request.DataList[i].Destination
		amounts[i] = request.DataList[i].Amount
	}
	totalRevealPrevOutput, err := tool.buildEmptyRevealTx(request.SingleRevealTxOnly, destinations, amounts, revealOutValue, request.FeeRate, request.EnableRBF)
	if err != nil {
		return err
	}
	err = tool.buildCommitTx(request.CommitTxOutPointList, totalRevealPrevOutput, request.CommitFeeRate, request.ChangeAddress, request.EnableRBF)
	if err != nil {
		return err
	}
	err = tool.completeRevealTx()
	if err != nil {
		return err
	}
	err = tool.signCommitTx()
	if err != nil {
		return errors.Wrap(err, "sign commit tx error")
	}
	return err
}

func (tool *SenderToolBox) CreateSenderTxCtxData(net *chaincfg.Params, data CoinReceiverData, onlyOneRevealAddr bool) (*senderTxCtxData, error) {
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, err
	}
	if onlyOneRevealAddr == true {
		privateKey = tool.revealPrivateKey
	}
	senderBuilder := txscript.NewScriptBuilder().
		AddData(schnorr.SerializePubKey(privateKey.PubKey())).
		AddOp(txscript.OP_CHECKSIG)
		/*
			AddOp(txscript.OP_FALSE).
			AddOp(txscript.OP_IF).
			AddData([]byte("ord")).
			// Two OP_DATA_1 should be OP_1. However, in the following link, it's not set as OP_1:
			// https://github.com/casey/ord/blob/0.5.1/src/inscription.rs#L17
			// Therefore, we use two OP_DATA_1 to maintain consistency with ord.
			AddOp(txscript.OP_DATA_1).
			AddOp(txscript.OP_DATA_1).
			AddData([]byte(data.ContentType)).
			AddOp(txscript.OP_0)
		*/
		/*
			maxChunkSize := 520
			bodySize := len(data.Body)
			for i := 0; i < bodySize; i += maxChunkSize {
				end := i + maxChunkSize
				if end > bodySize {
					end = bodySize
				}
				// to skip txscript.MaxScriptSize 10000
				inscriptionBuilder.AddFullData(data.Body[i:end])
			}
		*/
	senderScript, err := senderBuilder.Script()
	if err != nil {
		return nil, err
	}
	// to skip txscript.MaxScriptSize 10000
	//inscriptionScript = append(inscriptionScript, txscript.OP_ENDIF)

	leafNode := txscript.NewBaseTapLeaf(senderScript)
	proof := &txscript.TapscriptProof{
		TapLeaf:  leafNode,
		RootNode: leafNode,
	}

	controlBlock := proof.ToControlBlock(privateKey.PubKey())
	controlBlockWitness, err := controlBlock.ToBytes()
	if err != nil {
		return nil, err
	}

	tapHash := proof.RootNode.TapHash()
	commitTxAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootOutputKey(privateKey.PubKey(), tapHash[:])), net)
	if err != nil {
		return nil, err
	}
	commitTxAddressPkScript, err := txscript.PayToAddrScript(commitTxAddress)
	if err != nil {
		return nil, err
	}

	recoveryPrivateKeyWIF, err := btcutil.NewWIF(txscript.TweakTaprootPrivKey(*privateKey, tapHash[:]), net, true)
	if err != nil {
		return nil, err
	}

	return &senderTxCtxData{
		privateKey:              privateKey,
		senderScript:            senderScript,
		commitTxAddressPkScript: commitTxAddressPkScript,
		commitTxAddress:         commitTxAddress.EncodeAddress(),
		controlBlockWitness:     controlBlockWitness,
		recoveryPrivateKeyWIF:   recoveryPrivateKeyWIF.String(),
	}, nil
}

func (tool *SenderToolBox) buildEmptyRevealTx(singleRevealTxOnly bool, destination []string, amounts []int64, revealOutValue, feeRate int64, enableRBF bool) (int64, error) {
	var revealTx []*wire.MsgTx
	totalPrevOutput := int64(0)
	total := len(tool.txCtxDataList)
	addTxInTxOutIntoRevealTx := func(tx *wire.MsgTx, index int) error {
		in := wire.NewTxIn(&wire.OutPoint{Index: uint32(index)}, nil, nil)
		in.Sequence = sequenceNum(enableRBF)
		tx.AddTxIn(in)
		receiver, err := btcutil.DecodeAddress(destination[index], tool.net)
		if err != nil {
			return err
		}
		scriptPubKey, err := txscript.PayToAddrScript(receiver)
		if err != nil {
			return err
		}
		out := wire.NewTxOut(amounts[index], scriptPubKey)
		tx.AddTxOut(out)
		return nil
	}
	if singleRevealTxOnly {
		revealTx = make([]*wire.MsgTx, 1)
		tx := wire.NewMsgTx(wire.TxVersion)
		for i := 0; i < total; i++ {
			err := addTxInTxOutIntoRevealTx(tx, i)
			if err != nil {
				return 0, err
			}
		}
		eachRevealBaseTxFee := int64(tx.SerializeSize()) * feeRate / int64(total)
		prevOutput := (revealOutValue + eachRevealBaseTxFee) * int64(total)
		{
			emptySignature := make([]byte, 64)
			emptyControlBlockWitness := make([]byte, 33)
			for i := 0; i < total; i++ {
				fee := (int64(wire.TxWitness{emptySignature, tool.txCtxDataList[i].senderScript, emptyControlBlockWitness}.SerializeSize()+2+3) / 4) * feeRate
				tool.txCtxDataList[i].revealTxPrevOutput = &wire.TxOut{
					PkScript: tool.txCtxDataList[i].commitTxAddressPkScript,
					Value:    revealOutValue + eachRevealBaseTxFee + fee,
				}
				prevOutput += fee
			}
		}
		totalPrevOutput = prevOutput
		revealTx[0] = tx
	} else {
		revealTx = make([]*wire.MsgTx, total)
		for i := 0; i < total; i++ {
			tx := wire.NewMsgTx(wire.TxVersion)
			err := addTxInTxOutIntoRevealTx(tx, i)
			if err != nil {
				return 0, err
			}
			prevOutput := amounts[i] + int64(tx.SerializeSize())*feeRate
			{
				emptySignature := make([]byte, 64)
				emptyControlBlockWitness := make([]byte, 33)
				fee := (int64(wire.TxWitness{emptySignature, tool.txCtxDataList[i].senderScript, emptyControlBlockWitness}.SerializeSize()+2+3) / 4) * feeRate
				prevOutput += fee
				tool.txCtxDataList[i].revealTxPrevOutput = &wire.TxOut{
					PkScript: tool.txCtxDataList[i].commitTxAddressPkScript,
					Value:    prevOutput,
				}
			}
			totalPrevOutput += prevOutput
			revealTx[i] = tx
		}
	}
	tool.revealTx = revealTx
	return totalPrevOutput, nil
}

func (tool *SenderToolBox) getTxOutByOutPoint(outPoint *wire.OutPoint) (*wire.TxOut, error) {
	var txOut *wire.TxOut
	if tool.client.rpcClient != nil {
		tx, err := tool.client.rpcClient.GetRawTransactionVerbose(&outPoint.Hash)
		if err != nil {
			return nil, err
		}
		if int(outPoint.Index) >= len(tx.Vout) {
			return nil, errors.New("err out point")
		}
		vout := tx.Vout[outPoint.Index]
		pkScript, err := hex.DecodeString(vout.ScriptPubKey.Hex)
		if err != nil {
			return nil, err
		}
		amount, err := btcutil.NewAmount(vout.Value)
		if err != nil {
			return nil, err
		}
		txOut = wire.NewTxOut(int64(amount), pkScript)
	} else {
		tx, err := tool.client.btcApiClient.GetRawTransaction(&outPoint.Hash)
		if err != nil {
			return nil, err
		}
		if int(outPoint.Index) >= len(tx.TxOut) {
			return nil, errors.New("err out point")
		}
		txOut = tx.TxOut[outPoint.Index]
	}
	tool.commitTxPrevOutputFetcher.AddPrevOut(*outPoint, txOut)
	return txOut, nil
}

func (tool *SenderToolBox) buildLockedScript(address string) *[]byte {

	addr, err := btcutil.DecodeAddress(address, tool.net)
	if err != nil {
		return nil
	}
	scriptPubKey, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil
	}
	return &scriptPubKey
}

func (tool *SenderToolBox) buildCommitTx(commitTxOutPointList []*wire.OutPoint, totalRevealPrevOutput, commitFeeRate int64, changeAddress string, enableRBF bool) error {
	totalSenderAmount := btcutil.Amount(0)
	tx := wire.NewMsgTx(wire.TxVersion)
	var changePkScript *[]byte
	changePkScript = tool.buildLockedScript(changeAddress)
	for i := range commitTxOutPointList {
		txOut, err := tool.getTxOutByOutPoint(commitTxOutPointList[i])
		if err != nil {
			return err
		}
		if changePkScript == nil { // first sender as change address, 找零地址设置, 默认回到发送者地址；
			changePkScript = &txOut.PkScript
		}
		in := wire.NewTxIn(commitTxOutPointList[i], nil, nil)
		in.Sequence = sequenceNum(enableRBF)
		tx.AddTxIn(in)

		totalSenderAmount += btcutil.Amount(txOut.Value)
	}
	for i := range tool.txCtxDataList {
		tx.AddTxOut(tool.txCtxDataList[i].revealTxPrevOutput)
	}

	tx.AddTxOut(wire.NewTxOut(0, *changePkScript))
	fee := btcutil.Amount(mempool.GetTxVirtualSize(btcutil.NewTx(tx))) * btcutil.Amount(commitFeeRate)
	changeAmount := totalSenderAmount - btcutil.Amount(totalRevealPrevOutput) - fee
	if changeAmount > 0 {
		tx.TxOut[len(tx.TxOut)-1].Value = int64(changeAmount)
	} else {
		tx.TxOut = tx.TxOut[:len(tx.TxOut)-1]
		if changeAmount < 0 {
			feeWithoutChange := btcutil.Amount(mempool.GetTxVirtualSize(btcutil.NewTx(tx))) * btcutil.Amount(commitFeeRate)
			if totalSenderAmount-btcutil.Amount(totalRevealPrevOutput)-feeWithoutChange < 0 {
				return errors.New("insufficient balance")
			}
		}
	}
	tool.commitTx = tx
	return nil
}

func (tool *SenderToolBox) completeRevealTx() error {
	for i := range tool.txCtxDataList {
		tool.revealTxPrevOutputFetcher.AddPrevOut(wire.OutPoint{
			Hash:  tool.commitTx.TxHash(),
			Index: uint32(i),
		}, tool.txCtxDataList[i].revealTxPrevOutput)
		if len(tool.revealTx) == 1 {
			tool.revealTx[0].TxIn[i].PreviousOutPoint.Hash = tool.commitTx.TxHash()
		} else {
			tool.revealTx[i].TxIn[0].PreviousOutPoint.Hash = tool.commitTx.TxHash()
		}
	}
	witnessList := make([]wire.TxWitness, len(tool.txCtxDataList))
	for i := range tool.txCtxDataList {
		revealTx := tool.revealTx[0]
		idx := i
		if len(tool.revealTx) != 1 {
			revealTx = tool.revealTx[i]
			idx = 0
		}
		witnessArray, err := txscript.CalcTapscriptSignaturehash(txscript.NewTxSigHashes(revealTx, tool.revealTxPrevOutputFetcher),
			txscript.SigHashDefault, revealTx, idx, tool.revealTxPrevOutputFetcher, txscript.NewBaseTapLeaf(tool.txCtxDataList[i].senderScript))
		if err != nil {
			return err
		}
		signature, err := schnorr.Sign(tool.txCtxDataList[i].privateKey, witnessArray)
		if err != nil {
			return err
		}
		witnessList[i] = wire.TxWitness{signature.Serialize(), tool.txCtxDataList[i].senderScript, tool.txCtxDataList[i].controlBlockWitness}
	}
	for i := range witnessList {
		if len(tool.revealTx) == 1 {
			tool.revealTx[0].TxIn[i].Witness = witnessList[i]
		} else {
			tool.revealTx[i].TxIn[0].Witness = witnessList[i]
		}
	}
	// check tx max tx wight
	for i, tx := range tool.revealTx {
		revealWeight := blockchain.GetTransactionWeight(btcutil.NewTx(tx))
		if revealWeight > MaxStandardTxWeight {
			return errors.New(fmt.Sprintf("reveal(index %d) transaction weight greater than %d (MAX_STANDARD_TX_WEIGHT): %d", i, MaxStandardTxWeight, revealWeight))
		}
	}
	return nil
}

func (tool *SenderToolBox) signCommitTx() error {
	if len(tool.commitTxPrivateKeyList) == 0 {
		commitSignTransaction, isSignComplete, err := tool.client.rpcClient.SignRawTransactionWithWallet(tool.commitTx)
		if err != nil {
			//log.Errorf("sign commit tx error, %v", err)
			return err
		}
		if !isSignComplete {
			return errors.New("sign commit tx error")
		}
		tool.commitTx = commitSignTransaction
	} else {
		witnessList := make([]wire.TxWitness, len(tool.commitTx.TxIn))
		for i := range tool.commitTx.TxIn {
			txOut := tool.commitTxPrevOutputFetcher.FetchPrevOutput(tool.commitTx.TxIn[i].PreviousOutPoint)
			witness, err := txscript.TaprootWitnessSignature(tool.commitTx, txscript.NewTxSigHashes(tool.commitTx, tool.commitTxPrevOutputFetcher),
				i, txOut.Value, txOut.PkScript, txscript.SigHashDefault, tool.commitTxPrivateKeyList[i])
			if err != nil {
				return err
			}
			witnessList[i] = witness
		}
		for i := range witnessList {
			tool.commitTx.TxIn[i].Witness = witnessList[i]
		}
	}
	return nil
}

func (tool *SenderToolBox) BackupRecoveryKeyToRpcNode() error {
	if tool.client.rpcClient == nil {
		return errors.New("rpc client is nil")
	}
	descriptors := make([]jsonrpc.Descriptor, len(tool.txCtxDataList))
	for i := range tool.txCtxDataList {
		descriptorInfo, err := tool.client.rpcClient.GetDescriptorInfo(fmt.Sprintf("rawtr(%s)", tool.txCtxDataList[i].recoveryPrivateKeyWIF))
		if err != nil {
			return err
		}
		descriptors[i] = jsonrpc.Descriptor{
			Desc: *btcjson.String(fmt.Sprintf("rawtr(%s)#%s", tool.txCtxDataList[i].recoveryPrivateKeyWIF, descriptorInfo.Checksum)),
			Timestamp: btcjson.TimestampOrNow{
				Value: "now",
			},
			Active:    btcjson.Bool(false),
			Range:     nil,
			NextIndex: nil,
			Internal:  btcjson.Bool(false),
			Label:     btcjson.String("commit tx recovery key"),
		}
	}
	results, err := jsonrpc.ImportDescriptors(tool.client.rpcClient, descriptors)
	if err != nil {
		return err
	}
	if results == nil {
		return errors.New("commit tx recovery key import failed, nil result")
	}
	for _, result := range *results {
		if !result.Success {
			return errors.New("commit tx recovery key import failed")
		}
	}
	return nil
}

func (tool *SenderToolBox) GetRecoveryKeyWIFList() []string {
	wifList := make([]string, len(tool.txCtxDataList))
	for i := range tool.txCtxDataList {
		wifList[i] = tool.txCtxDataList[i].recoveryPrivateKeyWIF
	}
	return wifList
}

/*
	func getTxHex(tx *wire.MsgTx) (string, error) {
		var buf bytes.Buffer
		if err := tx.Serialize(&buf); err != nil {
			return "", err
		}
		return hex.EncodeToString(buf.Bytes()), nil
	}
*/
func (tool *SenderToolBox) GetCommitTxHex() (string, error) {
	return getTxHex(tool.commitTx)
}

func (tool *SenderToolBox) GetRevealTxHexList() ([]string, error) {
	txHexList := make([]string, len(tool.revealTx))
	for i := range tool.revealTx {
		txHex, err := getTxHex(tool.revealTx[i])
		if err != nil {
			return nil, err
		}
		txHexList[i] = txHex
	}
	return txHexList, nil
}

func (tool *SenderToolBox) sendRawTransaction(tx *wire.MsgTx) (*chainhash.Hash, error) {
	if tool.client.rpcClient != nil {
		return tool.client.rpcClient.SendRawTransaction(tx, false)
	} else {
		return tool.client.btcApiClient.BroadcastTx(tx)
	}
}

func (tool *SenderToolBox) calculateFee() int64 {
	fees := int64(0)
	for _, in := range tool.commitTx.TxIn {
		fees += tool.commitTxPrevOutputFetcher.FetchPrevOutput(in.PreviousOutPoint).Value
	}
	for _, out := range tool.commitTx.TxOut {
		fees -= out.Value
	}
	for _, tx := range tool.revealTx {
		for _, in := range tx.TxIn {
			fees += tool.revealTxPrevOutputFetcher.FetchPrevOutput(in.PreviousOutPoint).Value
		}
		for _, out := range tx.TxOut {
			fees -= out.Value
		}
	}
	return fees
}

func (tool *SenderToolBox) SendCoinsConcurrency() (commitTxHash *chainhash.Hash, revealTxHashList []*chainhash.Hash, fees int64, err error) {
	fees = tool.calculateFee()
	commitTxHash, err = tool.sendRawTransaction(tool.commitTx)
	if err != nil {
		return nil, nil, fees, errors.Wrap(err, "send commit tx error")
	}
	revealTxHashList = make([]*chainhash.Hash, len(tool.revealTx))
	for i := range tool.revealTx {
		_revealTxHash, err := tool.sendRawTransaction(tool.revealTx[i])
		if err != nil {
			return commitTxHash, revealTxHashList, fees, errors.Wrap(err, fmt.Sprintf("send reveal tx error, %d。", i))
		}
		revealTxHashList[i] = _revealTxHash
	}
	return commitTxHash, revealTxHashList, fees, nil
}

func (tool *SenderToolBox) RevealTxs() []*wire.MsgTx {
	return tool.revealTx
}

func (tool *SenderToolBox) CommitTx() *wire.MsgTx {
	return tool.commitTx
}

func (tool *SenderToolBox) GetCommitAddress() (string, error) {
	pkSript := tool.txCtxDataList[0].commitTxAddressPkScript
	_, address, _, err := txscript.ExtractPkScriptAddrs(pkSript, tool.net)
	if err != nil {
		return "", err
	}
	return address[0].String(), nil
}

func BatchTaprootSendCoinsImmediately(destinations []string, amounts []int64, utxoPrivateKeyHex string, enableRBF bool, onlyOneRevealAddr bool, netParams *chaincfg.Params) (
	*chainhash.Hash, []*chainhash.Hash, *SenderToolBox, int64, error) {
	if len(destinations) == 0 || len(amounts) == 0 || len(destinations) != len(amounts) {
		return nil, nil, nil, 0, errors.New("Error destination address amount != files amount")
	}

	//netParams := &chaincfg.TestNet3Params
	btcApiClient := memPool.NewClient(netParams)

	var dataList []CoinReceiverData
	for i := range destinations {
		dataList = append(dataList, CoinReceiverData{
			Destination: destinations[i],
			Amount:      amounts[i],
		})
	}

	commitTxOutPointList := make([]*wire.OutPoint, 0)
	commitTxPrivateKeyList := make([]*btcec.PrivateKey, 0)

	{
		utxoPrivateKeyBytes, err := hex.DecodeString(utxoPrivateKeyHex)
		if err != nil {
			return nil, nil, nil, 0, errors.New("decode privKey err," + err.Error())
		}
		utxoPrivateKey, _ := btcec.PrivKeyFromBytes(utxoPrivateKeyBytes)

		utxoTaprootAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootKeyNoScript(utxoPrivateKey.PubKey())), netParams)
		if err != nil {
			return nil, nil, nil, 0, errors.New("get p2tr addr err," + err.Error())
		}

		unspentList, err := btcApiClient.ListUnspent(utxoTaprootAddress)

		if err != nil {
			return nil, nil, nil, 0, errors.New("list unspent utxo err," + err.Error())
		}

		for i := range unspentList {
			commitTxOutPointList = append(commitTxOutPointList, unspentList[i].Outpoint)
			commitTxPrivateKeyList = append(commitTxPrivateKeyList, utxoPrivateKey)
		}
	}

	request := SenderRequest{
		CommitTxOutPointList:   commitTxOutPointList,
		CommitTxPrivateKeyList: commitTxPrivateKeyList,
		CommitFeeRate:          2,
		FeeRate:                1,
		DataList:               dataList,
		SingleRevealTxOnly:     false,
		EnableRBF:              enableRBF,
		OnlyOneRevealAddr:      onlyOneRevealAddr,
	}

	tool, err := NewSenderToolBoxWithBtcApiClient(netParams, btcApiClient, &request)
	if err != nil {
		return nil, nil, nil, 0, errors.New("init batch send coin client err," + err.Error())
	}
	commitTxHash, revealTxHashList, fees, err := tool.SendCoinsConcurrency()
	if err != nil {
		return nil, nil, nil, 0, errors.New(fmt.Sprintf("send coin concurrency err:%s", err.Error()))
	}
	/*
		log.Infof("commitTxHash:%s\n", commitTxHash.String())
		for i := range revealTxHashList {
			log.Infof("revealTxHash:%s", revealTxHashList[i].String())
			log.Infof("transmit address %d: %s\n", i, tool.transmitAddress(i))
		}

		log.Infof("reveal tx and commit tx total fees: %d", fees)
	*/
	return commitTxHash, revealTxHashList, tool, fees, nil
}

func TestOnceMuxReceivers(t *testing.T) {
	//unisat 测试网p2tr地址的私钥，p2tr: tb1pst5dyk7tdymccy0xydcyyzvqgz2022t8576sjz2l65fzulrvy5rqcv6les
	privateKeyHex := "a868774f27a34e28aef14a95e2ddfa9baf2bc9a83b632b111e8e9d1eb5fbb6e9"
	destinations := []string{
		//"tb1qg9hl3ulg20hel6aen5dtmhzhprjee039heu5hj",
		//"2ND3jnE3N6iCdzHf77SY44LinxRx5Vg7zDs",
		//"mzHchNdivKvhLRLww3VLC37VqwEHPMN3ak",
		"tb1pn0p79t9vx39kscnya2xwa45d79fc3x0ella2a634wcxwm0e0f6uqqm0ya9",
		"tb1pn0p79t9vx39kscnya2xwa45d79fc3x0ella2a634wcxwm0e0f6uqqm0ya9",
		"tb1pn0p79t9vx39kscnya2xwa45d79fc3x0ella2a634wcxwm0e0f6uqqm0ya9",
	}
	amounts := []int64{
		100,
		200,
		300,
	}
	enableRBF := true
	onlyOneRevealAddr := true
	commitTxHash, revealTxHashList, _, _, err := BatchTaprootSendCoinsImmediately(destinations, amounts, privateKeyHex, enableRBF, onlyOneRevealAddr, &chaincfg.TestNet3Params)
	println("commit tx hash:%s, revealTxHashList:%v", commitTxHash, revealTxHashList)
	if err != nil {
		//log.Errorf("batch inscribe NFT err:", err.Error())
	}
}
