package taproot

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	tx "github.com/satshub/go-bitcoind/transaction/taproot/internal"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

const (
	// bitcoin-cli -regtest sendtoaddress <script_address> 0.0001
	// previous outpoint
	p2trPrevHashStr   = "cc6bbc55755d2b3fc3a55bcb3fc9505804960a239abc0db9098c752aabd11003"
	p2trPrevIndex     = uint32(1)
	p2trPrevAmountSat = int64(10_000)

	// send address: bitcoin-cli -regtest getnewaddress "" bech32
	p2trSendAddrStr = "bcrt1quqqccct6wqpq9tp7qqw0j74cy4wkmrc5mt3d3t"
	p2trFeeSat      = int64(330)
)

var (
	//  <<signature>>
	//  <<preimage>>
	//
	//  OP_SHA256 <payment_hash> OP_EQUAL
	//  OP_IF
	//     <alicePubkey>
	//  OP_ELSE
	//     <bobPubkey>
	//  OP_ENDIF
	//  OP_CHKSIG
	//
	//  ↓↓
	//
	//  1)  OP_SHA256 <payment_hash> OP_EQUALVERIFY <alicePubkey> OP_CHKSIG
	//  2)  <bobPubkey> OP_CHECKSIG
	p2trPreimage, _ = hex.DecodeString("00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff")
	p2trPaymentHash = sha256.Sum256(p2trPreimage)

	p2trPrivkeyHexAlice, _ = hex.DecodeString("00112233445566778899aabbccddee0000112233445566778899aabbccddee00")
	p2trPrivkeyHexBob, _   = hex.DecodeString("00112233445566778899aabbccddee0100112233445566778899aabbccddee01")
	p2trKeyAlice           = tx.NewKey(p2trPrivkeyHexAlice, Network)
	p2trKeyBob             = tx.NewKey(p2trPrivkeyHexBob, Network)
)

func createP2trScript(pubkeyA []byte, pubkeyB []byte) [][]byte {
	const (
		OP_IF          = 0x63
		OP_ELSE        = 0x67
		OP_ENDIF       = 0x68
		OP_DROP        = 0x75
		OP_EQUAL       = 0x87
		OP_EQUALVERIFY = 0x88
		OP_SHA256      = 0xa8
		OP_CHKSIG      = 0xac
		OP_CSV         = 0xb2
	)

	part1a := []byte{OP_SHA256, byte(len(p2trPaymentHash))}
	// paymentHash[:]
	part1b := []byte{OP_EQUALVERIFY, byte(len(pubkeyA))}
	// pubkeyA
	part1c := []byte{OP_CHKSIG}
	script1 := make(
		[]byte,
		0,
		len(part1a)+
			len(p2trPaymentHash)+
			len(part1b)+
			len(pubkeyA)+
			len(part1c))
	script1 = append(script1, part1a...)
	script1 = append(script1, p2trPaymentHash[:]...)
	script1 = append(script1, part1b...)
	script1 = append(script1, pubkeyA...)
	script1 = append(script1, part1c...)

	part2a := []byte{byte(len(pubkeyB))}
	// pubkeyB
	part2b := []byte{OP_CHKSIG}
	script2 := make(
		[]byte,
		0,
		len(part2a)+
			len(pubkeyB)+
			len(part2b))
	script2 = append(script2, part2a...)
	script2 = append(script2, pubkeyB...)
	script2 = append(script2, part2b...)
	fmt.Printf("script1= %x\n", script1)
	fmt.Printf("script2= %x\n", script2)

	return [][]byte{script1, script2}
}

func ScriptPath() {
	scripts := createP2trScript(
		p2trKeyAlice.SerializeSchnorrPubKey(),
		p2trKeyBob.SerializeSchnorrPubKey())
	ts := tx.NewTapScript(p2trKeyBob, scripts, Network)
	p2tr, err := ts.CreateP2tr()
	if err != nil {
		fmt.Printf("fail CreateP2TR(): %v\n", err)
		return
	}
	fmt.Printf("send to this address: %s\n\n", p2tr)

	// redeem
	prevHash, _ := chainhash.NewHashFromStr(p2trPrevHashStr)
	tx, txid, err := ts.CreateRawTxP2TR(
		// previous output
		prevHash, p2trPrevIndex, p2trPrevAmountSat,
		// current output
		p2trSendAddrStr, p2trFeeSat,
		// unlock
		0, // script1 = scripts[0]
		[][]byte{p2trPreimage},
		p2trKeyAlice,
	)
	if err != nil {
		fmt.Printf("fail CreateRawTxP2TR(): %v\n", err)
		return
	}
	fmt.Printf("txid=%s\n", txid)
	fmt.Printf("tx= %x\n", tx)
}
