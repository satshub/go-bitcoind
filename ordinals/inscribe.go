package ordinals

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/satshub/go-bitcoind/mempool.space"
)

func BatchIssuedImmediately(destinations, files []string, utxoPrivateKeyHex string, netParams *chaincfg.Params) (
	*chainhash.Hash, []*chainhash.Hash, []string, int64, error) {
	if len(destinations) == 0 || len(files) == 0 || (len(destinations) != len(files)) {
		return nil, nil, []string{}, 0, errors.New("Error destination address amount != files amount")
	}

	//netParams := &chaincfg.TestNet3Params
	btcApiClient := mempool.NewClient(netParams)

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, nil, []string{}, 0, errors.New("get work dir err," + err.Error())
	}

	var dataList []InscriptionData
	for i := range destinations {
		filePath := fmt.Sprintf("%s/inscribefiles/%s", workingDir, files[i])
		// if file size too max will return sendrawtransaction RPC error: {"code":-26,"message":"tx-size"}
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return nil, nil, []string{}, 0, errors.New("read inscribe file err," + err.Error())
		}

		contentType := http.DetectContentType(fileContent)
		//log.Infof("file contentType %s", contentType)

		dataList = append(dataList, InscriptionData{
			ContentType: contentType,
			Body:        fileContent,
			Destination: destinations[i],
		})
	}

	commitTxOutPointList := make([]*wire.OutPoint, 0)
	commitTxPrivateKeyList := make([]*btcec.PrivateKey, 0)

	{
		utxoPrivateKeyBytes, err := hex.DecodeString(utxoPrivateKeyHex)
		if err != nil {
			return nil, nil, []string{}, 0, errors.New("decode privKey err," + err.Error())
		}
		utxoPrivateKey, _ := btcec.PrivKeyFromBytes(utxoPrivateKeyBytes)

		utxoTaprootAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootKeyNoScript(utxoPrivateKey.PubKey())), netParams)
		if err != nil {
			return nil, nil, []string{}, 0, errors.New("get p2tr addr err," + err.Error())
		}

		unspentList, err := btcApiClient.ListUnspent(utxoTaprootAddress)

		if err != nil {
			return nil, nil, []string{}, 0, errors.New("list unspent utxo err," + err.Error())
		}

		for i := range unspentList {
			commitTxOutPointList = append(commitTxOutPointList, unspentList[i].Outpoint)
			commitTxPrivateKeyList = append(commitTxPrivateKeyList, utxoPrivateKey)
		}
	}

	request := InscriptionRequest{
		CommitTxOutPointList:   commitTxOutPointList,
		CommitTxPrivateKeyList: commitTxPrivateKeyList,
		CommitFeeRate:          2,
		FeeRate:                1,
		DataList:               dataList,
		SingleRevealTxOnly:     false,
	}

	tool, err := NewInscriptionToolWithBtcApiClient(netParams, btcApiClient, &request)
	if err != nil {
		return nil, nil, []string{}, 0, errors.New("init inscription client err," + err.Error())
	}
	commitTxHash, revealTxHashList, inscriptions, failTxIndex, fees, err := tool.Inscribe()
	if err != nil {
		return nil, nil, []string{}, 0, errors.New("inscribe err," + err.Error())
	}
	if len(failTxIndex) > 0 {
		return nil, nil, []string{}, 0, errors.New("inscribe err, fail indexs:" + fmt.Sprintf("%+v", failTxIndex))
	}
	/*
		log.Infof("commitTxHash:%s", commitTxHash.String())
		for i := range revealTxHashList {
			log.Infof("revealTxHash:%s", revealTxHashList[i].String())
		}
		for i := range inscriptions {
			log.Infof("inscription, ", inscriptions[i])
		}
		log.Debugf("reveal tx and commit tx total fees: ", fees)
	*/
	return commitTxHash, revealTxHashList, inscriptions, fees, nil
}

func decodeContentTypeFromFileName(name string) string {
	fileType := strings.Split(strings.ToLower(name), ".")
	switch strings.ToLower(fileType[1]) {
	case "png":
		return "image/png"
	case "jpg":
		return "image/jpg"
	case "jpeg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "svg":
		return "image/svg+xml"
	case "webp":
		return "image/webp"
	case "mp4":
		return "video/mp4"
	case "mp3":
		return "audio/mpeg"
	case "txt":
		return "text/plain;charset=utf-8"
	case "html":
		return "text/html;charset=utf-8"
	default:
		println("Unknow inscribed nft file suffix:%s, fullname:%s", fileType, name)
		return "Unknown"
	}
}

func PrepareBatchIssued(projectId int32, destinations, files []string, utxoPrivateKeyHex string, feeRate, commitFeeRate int64, netParams *chaincfg.Params) (*InscriptionRequest, error) {
	if len(destinations) == 0 || len(files) == 0 || (len(destinations) != len(files)) {
		return nil, errors.New("Error destination address amount != files amount")
	}

	btcApiClient := mempool.NewClient(netParams)

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, errors.New("Error getting current working directory, " + err.Error())
	}

	var dataList []InscriptionData
	for i := range destinations {
		filePath := fmt.Sprintf("%s/inscribefiles/%d/%s", workingDir, projectId, files[i])
		// if file size too max will return sendrawtransaction RPC error: {"code":-26,"message":"tx-size"}
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return nil, errors.New("Error reading file " + err.Error())
		}

		contentType := decodeContentTypeFromFileName(files[i])
		//contentType := "image/svg+xml"
		//log.Infof("file name:%s,file contentType %s\n", files[i], contentType)
		if contentType == "Unknown" {
			return nil, errors.New("Error nft file type")
		}
		dataList = append(dataList, InscriptionData{
			ContentType: contentType,
			Body:        fileContent,
			Destination: destinations[i],
		})
	}

	commitTxOutPointList := make([]*wire.OutPoint, 0)
	commitTxPrivateKeyList := make([]*btcec.PrivateKey, 0)
	{
		utxoPrivateKeyBytes, err := hex.DecodeString(utxoPrivateKeyHex)
		if err != nil {
			return nil, errors.New("decode private key hex err:" + err.Error())
		}
		utxoPrivateKey, _ := btcec.PrivKeyFromBytes(utxoPrivateKeyBytes)
		utxoTaprootAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootKeyNoScript(utxoPrivateKey.PubKey())), netParams)
		if err != nil {
			return nil, errors.New("generate taproot addr err:" + err.Error())
		}

		unspentList, err := btcApiClient.ListUnspent(utxoTaprootAddress)

		if err != nil {
			return nil, errors.New("list unspent err " + err.Error())
		}

		for i := range unspentList {
			commitTxOutPointList = append(commitTxOutPointList, unspentList[i].Outpoint)
			commitTxPrivateKeyList = append(commitTxPrivateKeyList, utxoPrivateKey)
		}
	}

	request := InscriptionRequest{
		CommitTxOutPointList:   commitTxOutPointList,
		CommitTxPrivateKeyList: commitTxPrivateKeyList,
		CommitFeeRate:          commitFeeRate,
		FeeRate:                feeRate,
		DataList:               dataList,
		SingleRevealTxOnly:     false,
		RevealOutValue:         550,
		EnableRBF:              false,
	}

	return &request, nil
}
