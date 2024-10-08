package script

import (
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
)

type Key interface {
	Bytes() []byte
	String() string
}

type PubKey struct {
	key []byte
}

func NewPubKey(s string) (*PubKey, error) {
	if s == "" {
		return nil, errors.New("pubkey can't be empty")
	}

	key, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid key format: %w", err)
	}
	return &PubKey{key}, nil
}

func (pk *PubKey) Bytes() []byte  { return pk.key }
func (pk *PubKey) String() string { return hex.EncodeToString(pk.key) }

type XPub struct {
	key *hdkeychain.ExtendedKey
}

var (
	// keyOriginRegexp matches a key origin
	keyOriginRegexp = regexp.MustCompile(`\[[0-9a-fA-F]{8}.*?\]`)
	// xpubExprRegexp matches a Xpub and it's children derivation path
	xpubExprRegexp = regexp.MustCompile(`(\w+)(\/.+)?`)
)

func trimKeyOrigin(s string) string {
	return keyOriginRegexp.ReplaceAllString(s, "")
}

type xpubExpr struct {
	xpub     string
	children string
}

// parseXpubExpr returns a xpubExpr given a string
func parseXpubExpr(s string) (*xpubExpr, error) {
	submatch := xpubExprRegexp.FindStringSubmatch(s)
	if len(submatch) == 0 {
		// no match
		return nil, errors.New("invalid xpub format")
	}

	return &xpubExpr{
		xpub:     submatch[1],
		children: submatch[2],
	}, nil
}

// IsXPub returns if a string looks like an XPub or not
func IsXPub(s string) bool {
	marks := []string{"xpub", "vpub", "xpriv", "tprv", "ypub", "yprv", "zpub", "zprv"}
	for _, mark := range marks {
		if strings.Contains(s, mark) {
			return true
		}
	}
	return false
}

func NewXPub(s string) (*XPub, error) {
	expr, err := parseXpubExpr(s)
	if err != nil {
		return nil, err
	}

	xpub, err := newXPub(expr.xpub)
	if err != nil {
		return nil, err
	}

	if path := expr.children; path != "" {
		return xpub.Derive("m" + path)
	}

	return xpub, nil
}

func newXPub(s string) (*XPub, error) {
	key, err := hdkeychain.NewKeyFromString(s)
	if err != nil {
		return nil, err
	}

	return &XPub{key: key}, nil
}

func parsePath(path string, fn func(uint32) error) error {
	if !strings.HasPrefix(path, "m/") {
		return errors.New("xpub: invalid path prefix")
	}
	path = strings.TrimPrefix(path, "m/")

	// Hardened levels can be defined as `'`, `h` or `H` so unify them into `'`
	path = strings.Map(func(r rune) rune {
		if r == 'h' || r == 'H' {
			return '\''
		}
		return r
	}, path)

	levels := strings.Split(path, "/")
	for _, level := range levels {
		var v uint32

		// Verify if the level is hardened
		if strings.HasSuffix(level, "'") {
			v = 0x80000000
			level = strings.TrimSuffix(level, "'")
		}

		if level == "*" {
			continue
		}

		atoi, err := strconv.Atoi(level)
		if err != nil {
			return err
		}

		v += uint32(atoi)

		if err := fn(v); err != nil {
			return err
		}
	}

	return nil
}

func (xpub *XPub) Derive(path string) (*XPub, error) {
	key := xpub.key
	err := parsePath(path, func(i uint32) error {
		var err error
		key, err = key.Derive(i)
		return err
	})
	if err != nil {
		return xpub, err
	}

	return &XPub{key: key}, nil
}

func (xpub *XPub) Child(i uint32) (Key, error) {
	child, err := xpub.key.Derive(i)
	if err != nil {
		return nil, err
	}

	pub, err := child.ECPubKey()
	if err != nil {
		return nil, err
	}

	return &PubKey{key: pub.SerializeCompressed()}, nil
}

func (xpub *XPub) String() string { return xpub.key.String() }
func (xpub *XPub) PubKey() (string, error) {
	pub, err := xpub.key.ECPubKey()
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(pub.SerializeCompressed()), nil
}
