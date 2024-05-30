package jsonrpc

// Represents a block
type Block struct {
	// The block hash
	Hash string `json:"hash"`

	// The number of confirmations
	Confirmations uint64 `json:"confirmations"`

	// The block size
	Size uint64 `json:"size"`

	// The block height or index
	Height uint64 `json:"height"`

	// The block version
	Version uint32 `json:"version"`

	// The merkle root
	Merkleroot string `json:"merkleroot"`

	// Slice on transaction ids
	Tx []string `json:"tx"`

	// The block time in seconds since epoch (Jan 1 1970 GMT)
	Time int64 `json:"time"`

	// The nonce
	Nonce uint64 `json:"nonce"`

	// The bits
	Bits string `json:"bits"`

	// The difficulty
	Difficulty float64 `json:"difficulty"`

	// Total amount of work in active chain, in hexadecimal
	Chainwork string `json:"chainwork,omitempty"`

	// The hash of the previous block
	Previousblockhash string `json:"previousblockhash"`

	// The hash of the next block
	Nextblockhash string `json:"nextblockhash"`
}

// Represents a verbose 2 block
type BlockV2 struct {
	Block
	Tx []RawTransaction `json:"tx"`
}

// An Info represent a response to getmininginfo
type Info struct {
	// The server version
	Version uint32 `json:"version"`

	// The protocol version
	Protocolversion uint32 `json:"protocolversion"`

	// The wallet version
	Walletversion uint32 `json:"walletversion"`

	// The total bitcoin balance of the wallet
	Balance float64 `json:"balance"`

	// The current number of blocks processed in the server
	Blocks uint32 `json:"blocks"`

	// The time offset
	Timeoffset int32 `json:"timeoffset"`

	// The number of connections
	Connections uint32 `json:"connections"`

	// Tthe proxy used by the server
	Proxy string `json:"proxy,omitempty"`

	// Tthe current difficulty
	Difficulty float64 `json:"difficulty"`

	// If the server is using testnet or not
	Testnet bool `json:"testnet"`

	// The timestamp (seconds since GMT epoch) of the oldest pre-generated key in the key pool
	Keypoololdest uint64 `json:"keypoololdest"`

	// How many new keys are pre-generated
	KeypoolSize uint32 `json:"keypoolsize,omitempty"`

	// The timestamp in seconds since epoch (midnight Jan 1 1970 GMT) that the wallet is unlocked for transfers, or 0 if the wallet is locked
	UnlockedUntil int64 `json:"unlocked_until,omitempty"`

	// the transaction fee set in btc/kb
	Paytxfee float64 `json:"paytxfee"`

	// Minimum relay fee for non-free transactions in btc/kb
	Relayfee float64 `json:"relayfee"`

	//  Any error messages
	Errors string `json:"errors"`
}

// A MiningInfo represents a mininginfo response
type MiningInfo struct {
	// The current block
	Blocks uint64 `json:"blocks"`

	// The last block size
	CurrentBlocksize uint64 `json:"currentblocksize"`

	// The last block transaction
	CurrentBlockTx uint64 `json:"currentblocktx"`

	// The current difficulty
	Difficulty float64 `json:"difficulty"`

	// Current errors
	Errors string `json:"errors"`

	// The processor limit for generation. -1 if no generation. (see getgenerate or setgenerate calls)
	GenProcLimit int32 `json:"genproclimit"`

	// The size of the mem pool
	PooledtTx uint64 `json:"pooledtx"`

	// If using testnet or not
	Testnet bool `json:"testnet"`

	// If the generation is on or off (see getgenerate or setgenerate calls)
	Generate bool `json:"generate"`

	// The network hashrate
	NetworkHashps uint64 `json:"networkhashps"`

	// Node hashrate
	HashesPersec uint64 `json:"hashespersec"`
}

type Peer struct {
	// The ip address and port of the peer
	Addr string `json:"addr"`

	// Local address
	Addrlocal string `json:"addrlocal"`

	// The services
	Services string `json:"services"`

	// The time in seconds since epoch (Jan 1 1970 GMT) of the last send
	Lastsend uint64 `json:"lastsend"`

	// The time in seconds since epoch (Jan 1 1970 GMT) of the last receive
	Lastrecv uint64 `json:"lastrecv"`

	// The total bytes sent
	Bytessent uint64 `json:"bytessent"`

	// The total bytes received
	Bytesrecv uint64 `json:"bytesrecv"`

	// The connection time in seconds since epoch (Jan 1 1970 GMT)
	Conntime uint64 `json:"conntime"`

	// Ping time
	Pingtime float64 `json:"pingtime"`

	// Ping Wait
	Pingwait float64 `json:"pingwait"`

	// The peer version, such as 7001
	Version uint32 `json:"version"`

	// The string version
	Subver string `json:"subver"`

	// Inbound (true) or Outbound (false)
	Inbound bool `json:"inbound"`

	//  The starting height (block) of the peer
	Startingheight int32 `json:"startingheight"`

	// The ban score (stats.nMisbehavior)
	Banscore int32 `json:"banscore"`

	// If sync node
	Syncnode bool `json:"syncnode"`
}

// A ScriptSig represents a scriptsyg
type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

// Vin represent an IN value
type Vin struct {
	Coinbase  string    `json:"coinbase"`
	Txid      string    `json:"txid"`
	Vout      int       `json:"vout"`
	ScriptSig ScriptSig `json:"scriptSig"`
	Sequence  uint32    `json:"sequence"`
}

type ScriptPubKey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	ReqSigs   int      `json:"reqSigs,omitempty"`
	Type      string   `json:"type"`
	Address   string   `json:"address,omitempty"`
	Addresses []string `json:"addresses,omitempty"`
}

// Vout represent an OUT value
type Vout struct {
	Value        float64      `json:"value"`
	N            int          `json:"n"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

// RawTx represents a raw transaction
type RawTransaction struct {
	Hex           string `json:"hex"`
	Txid          string `json:"txid"`
	Version       uint32 `json:"version"`
	LockTime      uint32 `json:"locktime"`
	Vin           []Vin  `json:"vin"`
	Vout          []Vout `json:"vout"`
	BlockHash     string `json:"blockhash,omitempty"`
	Confirmations uint64 `json:"confirmations,omitempty"`
	Time          int64  `json:"time,omitempty"`
	Blocktime     int64  `json:"blocktime,omitempty"`
}

// TransactionDetails represents details about a transaction
type TransactionDetails struct {
	Account  string  `json:"account"`
	Address  string  `json:"address,omitempty"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Fee      float64 `json:"fee,omitempty"`
	Label    string  `json:"label,omitempty"`
}

// Transaction represents a transaction
type Transaction struct {
	Amount          float64              `json:"amount"`
	Account         string               `json:"account,omitempty"`
	Address         string               `json:"address,omitempty"`
	Category        string               `json:"category,omitempty"`
	Fee             float64              `json:"fee,omitempty"`
	Confirmations   int64                `json:"confirmations"`
	BlockHash       string               `json:"blockhash"`
	BlockIndex      int64                `json:"blockindex"`
	BlockTime       int64                `json:"blocktime"`
	TxID            string               `json:"txid"`
	WalletConflicts []string             `json:"walletconflicts"`
	Time            int64                `json:"time"`
	TimeReceived    int64                `json:"timereceived"`
	Details         []TransactionDetails `json:"details,omitempty"`
	Hex             string               `json:"hex,omitempty"`
}

// UTransactionOut represents a unspent transaction out (UTXO)
type UTransactionOut struct {
	Bestblock     string       `json:"bestblock"`
	Confirmations uint32       `json:"confirmations"`
	Value         float64      `json:"value"`
	ScriptPubKey  ScriptPubKey `json:"scriptPubKey"`
	Version       uint32       `json:"version"`
	Coinbase      bool         `json:"coinbase"`
}

// TransactionOutSet represents statistics about the unspent transaction output database
type TransactionOutSet struct {
	Height          uint32  `json:"height"`
	Bestblock       string  `json:"bestblock"`
	Transactions    float64 `json:"transactions"`
	TxOuts          float64 `json:"txouts"`
	BytesSerialized float64 `json:"bytes_serialized"`
	HashSerialized  string  `json:"hash_serialized"`
	TotalAmount     float64 `json:"total_amount"`
}

// A Work represents a formatted hash data to work on
type Work struct {
	Midstate string `json:"midstate"`
	Data     string `json:"data"`
	Hash1    string `json:"hash1"`
	Target   string `json:"target"`
}
