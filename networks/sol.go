package networks

import (
	"github.com/btcsuite/btcutil/base58"
	"hdwallet"
)

func DeriveSolanaAddress(h *hdwallet.HdWallet, hdPath string) (string, error) {
	pub, err := h.DerivePublicKeyForEd25519(hdPath)
	if err != nil {
		return "", err
	}
	return base58.Encode(pub), nil
}
