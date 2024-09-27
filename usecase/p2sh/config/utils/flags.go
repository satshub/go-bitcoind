/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package utils

import (
	"strings"

	"github.com/satshub/go-bitcoind/usecase/address/config"
	"github.com/urfave/cli"
)

var (
	ConfigFlag = cli.StringFlag{
		Name:  "config",
		Usage: "Genesis block config `<file>`. If doesn't specifies, use main net config as default.",
		Value: "./config.json",
	}

	LogLevelFlag = cli.UintFlag{
		Name:  "loglevel",
		Usage: "Set the log level to `<level>` (0~6). 0:Trace 1:Debug 2:Info 3:Warn 4:Error 5:Fatal 6:MaxLevel",
		Value: config.DEFAULT_LOG_LEVEL,
	}

	//pass := flag.String("pass", "", "protect bip39 mnemonic with a passphrase")
	//compress := true // generate a compressed public key
	//bip39Enable := flag.Bool("bip39", false, "mnemonic code for generating deterministic keys")

	//number := flag.Int("n", 10, "set number of keys to generate")
	//mnemonic := flag.String("mnemonic", "scout shoot capable river old waste air gauge execute share loop nothing", "optional list of words to re-generate a root key")
	//brain := flag.String("brain", "brain wallet context", "some words that will generate mnemonic")
	//net := flag.String("net", "mainnet", "address which network user want to create on")

	PasswordFlag = cli.StringFlag{
		Name:  "pass",
		Usage: "Protect bip39 mnemonic with a passphrase",
		Value: "",
	}

	CompressFlag = cli.BoolFlag{
		Name:  "compress",
		Usage: "Generate a compressed public key",
	}

	NumberFlag = cli.UintFlag{
		Name:  "num",
		Usage: "Set number of keys to generate",
		Value: 10,
	}

	MnemonicFlag = cli.StringFlag{
		Name:  "mnemonic",
		Usage: "Optional list of words to re-generate a root key",
	}

	BrainFlag = cli.StringFlag{
		Name:  "brain",
		Usage: "Some words that will generate mnemonic",
		Value: "hello world",
	}

	NetFlag = cli.StringFlag{
		Name:  "net",
		Usage: "Address which network user want to create on",
		Value: "signet",
	}

	HexFlag = cli.StringFlag{
		Name:     "hex",
		Usage:    "broadcast context",
		Value:    "",
		Required: true,
	}
)

// GetFlagName deal with short flag, and return the flag name whether flag name have short name
func GetFlagName(flag cli.Flag) string {
	name := flag.GetName()
	if name == "" {
		return ""
	}
	return strings.TrimSpace(strings.Split(name, ",")[0])
}
