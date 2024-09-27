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

type Database struct {
	Type     string
	User     string
	Password string
	Host     string
	Port     string
	Db       string
}

type Bitcoind struct {
	User     string
	Password string
	Host     string
	Port     string
	ZMQHost  string
}

type JsonRpc struct {
	Host string
	Port uint32
}

type AppConfig struct {
	Logger   Log      `json:"Logger"`
	Database Database `json:"Database"`
	Bitcoind Bitcoind `json:"Mainnet"`
	Jsonrpc  JsonRpc  `json:"Jsonrpc"`
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
