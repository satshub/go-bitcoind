package ordinals

import "fmt"

const GENESIS_CARD_SVG_HEAD = `<svg width="720" height="390" xmlns="http://www.w3.org/2000/svg"><defs><linearGradient x1="0%" y1="50%" x2="100%" y2="50%" id="a"><stop stop-color="#F18F19" offset="0%"/><stop stop-color="#FFD29C" offset="48.304%"/><stop stop-color="#F18F19" offset="99.913%"/></linearGradient></defs><g fill="none" fill-rule="evenodd"><rect fill="#000" width="720" height="390" rx="24"/><text fill="url(#a)" fill-rule="nonzero" font-family="Montserrat-SemiBold, Montserrat" font-size="60" font-weight="500"><tspan x="223" y="223">OG PASS</tspan></text><text font-family="Arial-Black, Arial Black" font-size="30" font-weight="700"><tspan x="40" y="73" fill="#FFF">De</tspan><tspan x="83.345" y="73" fill="#F18F19">index</tspan></text><text fill="#F18F19" x="595" y="340.15" font-size="24">`

const GENESIS_CARD_SVG_TAIL = `</text></g></svg>`

func GenesisCard(num int32) string {
	return fmt.Sprintf("%s#%04d%s", GENESIS_CARD_SVG_HEAD, num, GENESIS_CARD_SVG_TAIL)
}
