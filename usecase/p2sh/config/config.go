// Package config /**
package config

import (
	"bytes"
	"encoding/json"
	"os"
)

var Version = "dev-dirty"

type Log struct {
	LogLevel   int    `json:"LogLevel"`
	LogFileDir string `json:"LogFileDir"`
}

type Signers struct {
	WIFs      []string
	Threshold int
}

type Spent struct {
	WIFs           []string `json:"WIFs"`
	TxId           string   `json:"TxId"`
	UtxoIndex      uint32   `json:"UtxoIndex"`
	Receiver       string   `json:"Receiver"`
	ReceiverAmount int64    `json:"ReceiverAmount"`
	Charger        string   `json:"Charger"`
	ChargerAmount  int64    `json:"ChargerAmount"`
}

type AppConfig struct {
	Logger   Log     `json:"Logger"`
	Signers  Signers `json:"Signers"`
	Spent    Spent   `json:"Spent"`
	Network  string  `json:"Network"`
	Electrum string  `json:"Electrum"`
}

var AppConf AppConfig

func init() {
	file, err := os.ReadFile(ConfigName)
	if err != nil {
		panic("config  file error:" + err.Error())
	}
	// Remove the UTF-8 Byte Order Mark
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

	config := AppConfig{}
	if err := json.Unmarshal(file, &config); err != nil {
		panic("unmarshal json config err:" + err.Error())
	}
	AppConf = config
}
