package main

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/satshub/go-bitcoind/address"
	"github.com/satshub/go-bitcoind/usecase/address/config"
	"github.com/satshub/go-bitcoind/usecase/address/config/utils"
	"github.com/satshub/go-bitcoind/usecase/address/log"
	"github.com/tyler-smith/go-bip39"
	"github.com/urfave/cli"
)

var Bip39Executive = cli.Command{
	//Usage:     "Import blocks to DB from a file",
	Name:      "bip39",
	ArgsUsage: "",
	Action:    bip39Generator,
	Flags: []cli.Flag{
		utils.NetFlag,
		utils.NumberFlag,
		utils.BrainFlag,
		utils.MnemonicFlag,
		utils.PasswordFlag,
		utils.CompressFlag,
	},
	Description: "Generate a deterministic key using BIP39 mnemonic code",
}

var Bip44Executive = cli.Command{
	//Usage:     "Import blocks to DB from a file",
	Name:      "bip44",
	ArgsUsage: "",
	Action:    bip44Generator,
	Flags: []cli.Flag{
		utils.NetFlag,
		utils.NumberFlag,
		utils.BrainFlag,
		utils.MnemonicFlag,
		utils.PasswordFlag,
		utils.CompressFlag,
	},
	Description: "Generate a deterministic key using BIP44 mnemonic code",
}

var Bip49Executive = cli.Command{
	//Usage:     "Import blocks to DB from a file",
	Name:      "bip49",
	ArgsUsage: "",
	Action:    bip49Generator,
	Flags: []cli.Flag{
		utils.NetFlag,
		utils.NumberFlag,
		utils.BrainFlag,
		utils.MnemonicFlag,
		utils.PasswordFlag,
		utils.CompressFlag,
	},
	Description: "Generate a deterministic key using BIP49 mnemonic code",
}

var Bip84Executive = cli.Command{
	//Usage:     "Import blocks to DB from a file",
	Name:      "bip84",
	ArgsUsage: "",
	Action:    bip84Generator,
	Flags: []cli.Flag{
		utils.NetFlag,
		utils.NumberFlag,
		utils.BrainFlag,
		utils.MnemonicFlag,
		utils.PasswordFlag,
		utils.CompressFlag,
	},
	Description: "Generate a deterministic key using BIP84 mnemonic code",
}

func bip39Generator(ctx *cli.Context) error {
	compress := true
	log.InitLog(config.AppConf.Logger.LogLevel, config.AppConf.Logger.LogFileDir, log.Stdout)

	address.Network = ctx.String(utils.GetFlagName(utils.NetFlag))

	fmt.Printf("\n%-34s %-52s %-42s %-42s %s\n", "Bitcoin Address", "WIF(Wallet Import Format)", "SegWit(bech32)", "SegWit(nested)", "P2TR")
	fmt.Println(strings.Repeat("-", 185))

	for i := uint(0); i < ctx.Uint(utils.GetFlagName(utils.NumberFlag)); i++ {
		wif, address, segwitBech32, segwitNested, p2tr, err := address.Generate(compress)
		if err != nil {
			panic("bip39 generate error:" + err.Error())
		}
		fmt.Printf("%-34s %s %s %s %s\n", address, wif, segwitBech32, segwitNested, p2tr)
	}
	fmt.Println()
	return nil
}

func bip44Generator(ctx *cli.Context) error {
	compress := true
	log.InitLog(config.AppConf.Logger.LogLevel, config.AppConf.Logger.LogFileDir, log.Stdout)

	address.Network = ctx.String(utils.GetFlagName(utils.NetFlag))

	mnemonic := func() string {
		if brain := ctx.String(utils.GetFlagName(utils.BrainFlag)); brain != "" {
			fmt.Printf("%-18s %s", "Brain Context:", ctx.String(utils.GetFlagName(utils.BrainFlag)))
			entropy, err := hex.DecodeString(address.Sha256(brain))
			if err != nil {
				panic("BIP44 decode sha-brain context error" + err.Error())
			}
			mnemonic, err := bip39.NewMnemonic(entropy)
			if err != nil {
				panic("BIP44 new mnemonic error" + err.Error())
			}
			return mnemonic
		}
		return ctx.String(utils.GetFlagName(utils.MnemonicFlag))
	}()

	km, err := address.NewKeyManager(128, ctx.String(utils.GetFlagName(utils.PasswordFlag)), mnemonic)
	if err != nil {
		panic("BIP44 new key manager error:" + err.Error())
	}
	masterKey, err := km.GetMasterKey()
	if err != nil {
		panic("BIP44 get master key error:" + err.Error())
	}
	passphrase := km.GetPassphrase()
	if passphrase == "" {
		passphrase = "<none>"
	}
	fmt.Printf("\n%-18s %s\n", "BIP39 Mnemonic:", km.GetMnemonic())
	fmt.Printf("%-18s %s\n", "BIP39 Passphrase:", passphrase)
	fmt.Printf("%-18s %x\n", "BIP39 Seed:", km.GetSeed())
	fmt.Printf("%-18s %s\n", "BIP32 Root Key:", masterKey.B58Serialize())
	fmt.Printf("%-18s %s\n", "Address Network:", address.Network)

	fmt.Printf("\n%-18s %-34s %-60s  %s\n", "Path(BIP44)", "Bitcoin Address", "WIF(Wallet Import Format)", "P2TR")
	fmt.Println(strings.Repeat("-", 170))
	for i := 0; i < ctx.Int(utils.GetFlagName(utils.NumberFlag)); i++ {
		key, err := km.GetKey(address.PurposeBIP44, address.CoinTypeBTC, 0, 0, uint32(i))
		if err != nil {
			panic("get key error:" + err.Error())
		}
		wif, address, _, _, p2tr, err := key.Encode(compress)
		if err != nil {
			panic("encode key error:" + err.Error())
		}

		fmt.Printf("%-18s %-34s %s  %s\n", key.GetPath(), address, wif, p2tr)
	}

	return nil
}

func bip49Generator(ctx *cli.Context) error {
	compress := true
	log.InitLog(config.AppConf.Logger.LogLevel, config.AppConf.Logger.LogFileDir, log.Stdout)

	address.Network = ctx.String(utils.GetFlagName(utils.NetFlag))

	mnemonic := func() string {
		if brain := ctx.String(utils.GetFlagName(utils.BrainFlag)); brain != "" {
			fmt.Printf("%-18s %s", "Brain Context:", ctx.String(utils.GetFlagName(utils.BrainFlag)))
			entropy, err := hex.DecodeString(address.Sha256(brain))
			if err != nil {
				panic("BIP49 decode sha-brain context error" + err.Error())
			}

			mnemonic, err := bip39.NewMnemonic(entropy)
			if err != nil {
				panic("BIP49 new mnemonic error" + err.Error())
			}
			return mnemonic
		}
		return ctx.String(utils.GetFlagName(utils.MnemonicFlag))
	}()

	km, err := address.NewKeyManager(128, ctx.String(utils.GetFlagName(utils.PasswordFlag)), mnemonic)
	if err != nil {
		panic("BIP49 new key manager error:" + err.Error())
	}
	masterKey, err := km.GetMasterKey()
	if err != nil {
		panic("BIP49 get master key error:" + err.Error())
	}
	passphrase := km.GetPassphrase()
	if passphrase == "" {
		passphrase = "<none>"
	}
	fmt.Printf("\n%-18s %s\n", "BIP39 Mnemonic:", km.GetMnemonic())
	fmt.Printf("%-18s %s\n", "BIP39 Passphrase:", passphrase)
	fmt.Printf("%-18s %x\n", "BIP39 Seed:", km.GetSeed())
	fmt.Printf("%-18s %s\n", "BIP32 Root Key:", masterKey.B58Serialize())
	fmt.Printf("%-18s %s\n", "Address Network:", address.Network)

	fmt.Printf("\n%-18s %-34s %-60s %s\n", "Path(BIP49)", "SegWit(nested)", "WIF(Wallet Import Format)", "P2TR")
	fmt.Println(strings.Repeat("-", 170))
	for i := 0; i < ctx.Int(utils.GetFlagName(utils.NumberFlag)); i++ {
		key, err := km.GetKey(address.PurposeBIP49, address.CoinTypeBTC, 0, 0, uint32(i))
		if err != nil {
			panic("BIP49 get key error:" + err.Error())
		}
		wif, _, _, segwitNested, p2tr, err := key.Encode(compress)
		if err != nil {
			panic("BIP49 encode key error:" + err.Error())
		}

		fmt.Printf("%-18s %s %s %s\n", key.GetPath(), segwitNested, wif, p2tr)
	}

	return nil
}

func bip84Generator(ctx *cli.Context) error {
	compress := true
	log.InitLog(config.AppConf.Logger.LogLevel, config.AppConf.Logger.LogFileDir, log.Stdout)

	address.Network = ctx.String(utils.GetFlagName(utils.NetFlag))

	mnemonic := func() string {
		if brain := ctx.String(utils.GetFlagName(utils.BrainFlag)); brain != "" {
			fmt.Printf("%-18s %s", "Brain Context:", ctx.String(utils.GetFlagName(utils.BrainFlag)))
			entropy, err := hex.DecodeString(address.Sha256(brain))
			if err != nil {
				panic("BIP84 decode sha-brain context error" + err.Error())
			}
			mnemonic, err := bip39.NewMnemonic(entropy)
			if err != nil {
				panic("BIP84 new mnemonic error" + err.Error())
			}
			return mnemonic
		}
		return ctx.String(utils.GetFlagName(utils.MnemonicFlag))
	}()

	km, err := address.NewKeyManager(128, ctx.String(utils.GetFlagName(utils.PasswordFlag)), mnemonic)
	if err != nil {
		panic("BIP84 new key manager error:" + err.Error())
	}
	masterKey, err := km.GetMasterKey()
	if err != nil {
		panic("BIP84 get master key error:" + err.Error())
	}
	passphrase := km.GetPassphrase()
	if passphrase == "" {
		passphrase = "<none>"
	}
	fmt.Printf("\n%-18s %s\n", "BIP39 Mnemonic:", km.GetMnemonic())
	fmt.Printf("%-18s %s\n", "BIP39 Passphrase:", passphrase)
	fmt.Printf("%-18s %x\n", "BIP39 Seed:", km.GetSeed())
	fmt.Printf("%-18s %s\n", "BIP32 Root Key:", masterKey.B58Serialize())
	fmt.Printf("%-18s %s\n", "Address Network:", address.Network)

	fmt.Printf("\n%-18s %-42s %-60s %s\n", "Path(BIP84)", "SegWit(bech32)", "WIF(Wallet Import Format)", "P2TR")
	fmt.Println(strings.Repeat("-", 180))
	for i := 0; i < ctx.Int(utils.GetFlagName(utils.NumberFlag)); i++ {
		key, err := km.GetKey(address.PurposeBIP84, address.CoinTypeBTC, 0, 0, uint32(i))
		if err != nil {
			panic("BIP84 get key error:" + err.Error())
		}
		wif, _, segwitBech32, _, p2tr, err := key.Encode(compress)
		if err != nil {
			panic("BIP84 encode key error:" + err.Error())
		}

		fmt.Printf("%-18s %s %s %s\n", key.GetPath(), segwitBech32, wif, p2tr)
	}
	fmt.Println()
	return nil
}
