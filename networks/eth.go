package networks

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"hdwallet"
)

// DeriveEthereumAddress DeriveAddress derives the account address of the derivation path.
func DeriveEthereumAddress(h *hdwallet.HdWallet, path string) (string, error) {

	publicKeyECDSA, err := h.DerivePublicKey(path)
	if err != nil {
		return "", err
	}
	pubBytes := crypto.FromECDSAPub(publicKeyECDSA)

	fmt.Println("Public Key = " + hexutil.Encode(pubBytes))
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address.Hex(), nil
}
