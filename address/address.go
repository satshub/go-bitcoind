package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

// Purpose BIP43 - Purpose Field for Deterministic Wallets
// https://github.com/bitcoin/bips/blob/master/bip-0043.mediawiki
//
// Purpose is a constant set to 44' (or 0x8000002C) following the BIP43 recommendation.
// It indicates that the subtree of this node is used according to this specification.
//
// What does 44' mean in BIP44?
// https://bitcoin.stackexchange.com/questions/74368/what-does-44-mean-in-bip44
//
// 44' means that hardened keys should be used. The distinguisher for whether
// a key a given index is hardened is that the index is greater than 2^31,
// which is 2147483648. In hex, that is 0x80000000. That is what the apostrophe (') means.
// The 44 comes from adding it to 2^31 to get the final hardened key index.
// In hex, 44 is 2C, so 0x80000000 + 0x2C = 0x8000002C.
type Purpose = uint32

const (
	PurposeBIP44 Purpose = 0x8000002C // 44' BIP44
	PurposeBIP49 Purpose = 0x80000031 // 49' BIP49
	PurposeBIP84 Purpose = 0x80000054 // 84' BIP84
)

// CoinType SLIP-0044 : Registered coin types for BIP-0044
// https://github.com/satoshilabs/slips/blob/master/slip-0044.md
type CoinType = uint32

const (
	CoinTypeBTC CoinType = 0x80000000
	CoinTypeLTC CoinType = 0x80000002
	CoinTypeETH CoinType = 0x8000003c
	CoinTypeEOS CoinType = 0x800000c2
)

const (
	Apostrophe uint32 = 0x80000000 // 0'
)

type Key struct {
	path     string
	bip32Key *bip32.Key
}

func (k *Key) Encode(compress bool) (wif, address, segwitBech32, segwitNested, p2tr string, err error) {
	prvKey, _ := btcec.PrivKeyFromBytes(k.bip32Key.Key)
	return GenerateFromBytes(prvKey, compress)
}

// https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki
// bip44 define the following 5 levels in BIP32 path:
// m / purpose' / coin_type' / account' / change / address_index

func (k *Key) GetPath() string {
	return k.path
}

type KeyManager struct {
	mnemonic   string
	passphrase string
	keys       map[string]*bip32.Key
	mux        sync.Mutex
}

// NewKeyManager return new key manager
// bitSize has to be a multiple 32 and be within the inclusive range of {128, 256}
// 128: 12 phrases
// 256: 24 phrases
func NewKeyManager(bitSize int, passphrase, mnemonic string) (*KeyManager, error) {

	if mnemonic == "" {
		entropy, err := bip39.NewEntropy(bitSize)
		if err != nil {
			return nil, err
		}
		mnemonic, err = bip39.NewMnemonic(entropy)
		if err != nil {
			return nil, err
		}
	}

	km := &KeyManager{
		mnemonic:   mnemonic,
		passphrase: passphrase,
		keys:       make(map[string]*bip32.Key, 0),
	}
	return km, nil
}

func (km *KeyManager) GetMnemonic() string {
	return km.mnemonic
}

func (km *KeyManager) GetPassphrase() string {
	return km.passphrase
}

func (km *KeyManager) GetSeed() []byte {
	return bip39.NewSeed(km.GetMnemonic(), km.GetPassphrase())
}

func (km *KeyManager) getKey(path string) (*bip32.Key, bool) {
	km.mux.Lock()
	defer km.mux.Unlock()

	key, ok := km.keys[path]
	return key, ok
}

func (km *KeyManager) setKey(path string, key *bip32.Key) {
	km.mux.Lock()
	defer km.mux.Unlock()

	km.keys[path] = key
}

func (km *KeyManager) GetMasterKey() (*bip32.Key, error) {
	path := "m"

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	key, err := bip32.NewMasterKey(km.GetSeed())
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

func (km *KeyManager) GetPurposeKey(purpose uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'`, purpose-Apostrophe)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	parent, err := km.GetMasterKey()
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(purpose)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

func (km *KeyManager) GetCoinTypeKey(purpose, coinType uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'`, purpose-Apostrophe, coinType-Apostrophe)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	parent, err := km.GetPurposeKey(purpose)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(coinType)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

func (km *KeyManager) GetAccountKey(purpose, coinType, account uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d'`, purpose-Apostrophe, coinType-Apostrophe, account)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	parent, err := km.GetCoinTypeKey(purpose, coinType)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(account + Apostrophe)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

// GetChangeKey ...
// https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki#change
// change constant 0 is used for external chain
// change constant 1 is used for internal chain (also known as change addresses)
func (km *KeyManager) GetChangeKey(purpose, coinType, account, change uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d'/%d`, purpose-Apostrophe, coinType-Apostrophe, account, change)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	parent, err := km.GetAccountKey(purpose, coinType, account)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(change)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

func (km *KeyManager) GetKey(purpose, coinType, account, change, index uint32) (*Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d'/%d/%d`, purpose-Apostrophe, coinType-Apostrophe, account, change, index)

	key, ok := km.getKey(path)
	if ok {
		return &Key{path: path, bip32Key: key}, nil
	}

	parent, err := km.GetChangeKey(purpose, coinType, account, change)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(index)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return &Key{path: path, bip32Key: key}, nil
}

func Generate(compress bool) (wif, address, segwitBech32, segwitNested, p2tr string, err error) {
	prvKey, err := btcec.NewPrivateKey()
	if err != nil {
		return "", "", "", "", "", err
	}
	return GenerateFromBytes(prvKey, compress)
}

func GenerateFromBytes(prvKey *btcec.PrivateKey, compress bool) (wif, address, segwitBech32, segwitNested, p2tr string, err error) {
	// generate the wif(wallet import format) string
	btcwif, err := btcutil.NewWIF(prvKey, NetworkParams(), compress)
	if err != nil {
		return "", "", "", "", "", err
	}
	wif = btcwif.String()

	// generate a normal p2pkh address
	serializedPubKey := btcwif.SerializePubKey()
	addressPubKey, err := btcutil.NewAddressPubKey(serializedPubKey, NetworkParams())
	if err != nil {
		return "", "", "", "", "", err
	}
	address = addressPubKey.EncodeAddress()

	// generate a normal p2wkh address from the pubkey hash
	witnessProg := btcutil.Hash160(serializedPubKey)
	addressWitnessPubKeyHash, err := btcutil.NewAddressWitnessPubKeyHash(witnessProg, NetworkParams())
	if err != nil {
		return "", "", "", "", "", err
	}
	segwitBech32 = addressWitnessPubKeyHash.EncodeAddress()

	// generate an address which is
	// backwards compatible to Bitcoin nodes running 0.6.0 onwards, but
	// allows us to take advantage of segwit's scripting improvments,
	// and malleability fixes.
	serializedScript, err := txscript.PayToAddrScript(addressWitnessPubKeyHash)
	if err != nil {
		return "", "", "", "", "", err
	}
	addressScriptHash, err := btcutil.NewAddressScriptHash(serializedScript, NetworkParams())
	if err != nil {
		return "", "", "", "", "", err
	}
	segwitNested = addressScriptHash.EncodeAddress()

	p2trAddr := genP2TRAddress(prvKey)
	return wif, address, segwitBech32, segwitNested, p2trAddr, nil
}

func genP2TRAddress(privateKey *btcec.PrivateKey) string {
	privateKey, err := btcec.NewPrivateKey()
	taprootAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootKeyNoScript(privateKey.PubKey())), NetworkParams())
	if err != nil {
		log.Fatal(err)
	}
	return taprootAddress.EncodeAddress()
}

func Sha256(context string) string {
	h := sha256.New()
	h.Write([]byte(context))
	bs := h.Sum(nil)

	return fmt.Sprintf("%x", bs)
}

var network string

func NetworkParams() *chaincfg.Params {
	switch network {
	case "mainnet":
		return &chaincfg.MainNetParams
	case "testnet":
		return &chaincfg.TestNet3Params
	case "signet":
		return &chaincfg.SigNetParams
	case "simnet":
		return &chaincfg.SimNetParams
	default:
		return &chaincfg.MainNetParams
	}
}
func main() {
	compress := true // generate a compressed public key
	bip39Enable := flag.Bool("bip39", false, "mnemonic code for generating deterministic keys")
	pass := flag.String("pass", "", "protect bip39 mnemonic with a passphrase")
	number := flag.Int("n", 10, "set number of keys to generate")
	mnemonic := flag.String("mnemonic", "scout shoot capable river old waste air gauge execute share loop nothing", "optional list of words to re-generate a root key")
	brain := flag.String("brain", "brain wallet context", "some words that will generate mnemonic")
	net := flag.String("net", "mainnet", "address which network user want to create on")

	flag.Parse()

	network = *net

	if *brain != "" {
		fmt.Printf("%-18s %s", "Brain Context:", *brain)
		entropy, e := hex.DecodeString(Sha256(*brain))
		if e != nil {
			fmt.Println("hex.DecodeString error")
			return
		}
		// generate a mnemomic
		*mnemonic, _ = bip39.NewMnemonic(entropy)
	}

	if !*bip39Enable {
		fmt.Printf("\n%-34s %-52s %-42s %-42s %s\n", "Bitcoin Address", "WIF(Wallet Import Format)", "SegWit(bech32)", "SegWit(nested)", "P2TR")
		fmt.Println(strings.Repeat("-", 185))

		for i := 0; i < *number; i++ {
			wif, address, segwitBech32, segwitNested, p2tr, err := Generate(compress)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%-34s %s %s %s %s\n", address, wif, segwitBech32, segwitNested, p2tr)
		}
		fmt.Println()
		return
	}

	km, err := NewKeyManager(128, *pass, *mnemonic)
	if err != nil {
		log.Fatal(err)
	}
	masterKey, err := km.GetMasterKey()
	if err != nil {
		log.Fatal(err)
	}
	passphrase := km.GetPassphrase()
	if passphrase == "" {
		passphrase = "<none>"
	}
	fmt.Printf("\n%-18s %s\n", "BIP39 Mnemonic:", km.GetMnemonic())
	fmt.Printf("%-18s %s\n", "BIP39 Passphrase:", passphrase)
	fmt.Printf("%-18s %x\n", "BIP39 Seed:", km.GetSeed())
	fmt.Printf("%-18s %s\n", "BIP32 Root Key:", masterKey.B58Serialize())

	fmt.Printf("\n%-18s %-34s %-60s  %s\n", "Path(BIP44)", "Bitcoin Address", "WIF(Wallet Import Format)", "P2TR")
	fmt.Println(strings.Repeat("-", 170))
	for i := 0; i < *number; i++ {
		key, err := km.GetKey(PurposeBIP44, CoinTypeBTC, 0, 0, uint32(i))
		if err != nil {
			log.Fatal(err)
		}
		wif, address, _, _, p2tr, err := key.Encode(compress)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%-18s %-34s %s  %s\n", key.GetPath(), address, wif, p2tr)
	}

	fmt.Printf("\n%-18s %-34s %-60s %s\n", "Path(BIP49)", "SegWit(nested)", "WIF(Wallet Import Format)", "P2TR")
	fmt.Println(strings.Repeat("-", 170))
	for i := 0; i < *number; i++ {
		key, err := km.GetKey(PurposeBIP49, CoinTypeBTC, 0, 0, uint32(i))
		if err != nil {
			log.Fatal(err)
		}
		wif, _, _, segwitNested, p2tr, err := key.Encode(compress)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%-18s %s %s %s\n", key.GetPath(), segwitNested, wif, p2tr)
	}

	fmt.Printf("\n%-18s %-42s %-60s %s\n", "Path(BIP84)", "SegWit(bech32)", "WIF(Wallet Import Format)", "P2TR")
	fmt.Println(strings.Repeat("-", 180))
	for i := 0; i < *number; i++ {
		key, err := km.GetKey(PurposeBIP84, CoinTypeBTC, 0, 0, uint32(i))
		if err != nil {
			log.Fatal(err)
		}
		wif, _, segwitBech32, _, p2tr, err := key.Encode(compress)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%-18s %s %s %s\n", key.GetPath(), segwitBech32, wif, p2tr)
	}
	fmt.Println()
}

func _main() {
	netParams := &chaincfg.SigNetParams
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		log.Fatal(err)
	}
	privateKeyHex := hex.EncodeToString(privateKey.Serialize())
	log.Printf("new priviate key %s \n", privateKeyHex)

	taprootAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootKeyNoScript(privateKey.PubKey())), netParams)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("new taproot address %s \n", taprootAddress.EncodeAddress())

	restorePrivateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		log.Fatal(err)
	}
	restorePrivateKey, _ := btcec.PrivKeyFromBytes(restorePrivateKeyBytes)

	restoreTaprootAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootKeyNoScript(restorePrivateKey.PubKey())), netParams)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("restore taproot address %s \n", restoreTaprootAddress.EncodeAddress())

	if taprootAddress.EncodeAddress() != restoreTaprootAddress.EncodeAddress() {
		log.Fatal("restore privateKey error")
	}
	/**
	test btc faucet
	https://signetfaucet.com/
	https://alt.signetfaucet.com/
	*/
}
