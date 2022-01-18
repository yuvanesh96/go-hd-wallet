package hdwallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"errors"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/stellar/go/exp/crypto/derivation"
	"github.com/tyler-smith/go-bip39"
)

// HdWallet is the underlying wallet struct.
type HdWallet struct {
	mnemonic string
	seed     []byte
}

// NewHdWalletFromMnemonic returns a new hd wallet from a BIP-39 mnemonic.
func NewHdWalletFromMnemonic(mnemonic string) (*HdWallet, error) {
	if mnemonic == "" {
		return nil, errors.New("mnemonic is required")
	}

	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("mnemonic is invalid")
	}

	seed, err := getSeedFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}
	wallet := &HdWallet{
		mnemonic: mnemonic,
		seed:     seed,
	}

	return wallet, nil
}

// getSeedFromMnemonic returns a BIP-39 seed based on a BIP-39 mnemonic.
func getSeedFromMnemonic(mnemonic string) ([]byte, error) {
	if mnemonic == "" {
		return nil, errors.New("mnemonic is required")
	}
	return bip39.NewSeedWithErrorChecking(mnemonic, "")
}

// derivePrivateKey derives the private key of the Hd path.
func (h *HdWallet) derivePrivateKey(path accounts.DerivationPath) (*ecdsa.PrivateKey, error) {

	master, err := hdkeychain.NewMaster(h.seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	key := master
	for _, n := range path {
		key, err = key.Child(n)
		if err != nil {
			return nil, err
		}
	}
	privateKey, err := key.ECPrivKey()
	privateKeyECDSA := privateKey.ToECDSA()
	if err != nil {
		return nil, err
	}
	return privateKeyECDSA, nil
}

// DerivePublicKey derives the public key of the Hd path.
func (h *HdWallet) DerivePublicKey(path string) (*ecdsa.PublicKey, error) {

	derivationPath, err := accounts.ParseDerivationPath(path)
	if err != nil {
		return nil, errors.New("error in parsing the hd path")
	}

	privateKeyECDSA, err := h.derivePrivateKey(derivationPath)
	if err != nil {
		return nil, err
	}

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error getting public key")
	}

	return publicKeyECDSA, nil
}

func (h *HdWallet) DerivePublicKeyForEd25519(path string) (ed25519.PublicKey, error) {
	key, err := derivation.DeriveForPath(path, h.seed)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	rawSeed := key.RawSeed()
	reader := bytes.NewReader(rawSeed[:])
	pub, _, err := ed25519.GenerateKey(reader)
	return pub, nil
}
