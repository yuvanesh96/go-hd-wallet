package main

import (
	"fmt"
	"hdwallet"
	"hdwallet/networks"
	"hdwallet/util"
	"log"
)

func main() {

	mnemonic := "canvas mule stage intact brick account category observe hybrid napkin claim ensure"
	wallet, err := hdwallet.NewHdWalletFromMnemonic(mnemonic)
	if err != nil {
		log.Fatal("error while setting up wallet")
	}

	// 1. Eth HD Path
	ethHdPath := "m/44'/60'/0'/0/0"
	addrs, err := networks.DeriveEthereumAddress(wallet, ethHdPath)
	fmt.Println("Address for the Hd Path", ethHdPath, "is ", addrs)

	// 2. Solana HD Path
	solanaHdPath := "m/44'/501'/0'/0'"
	addrs, err = networks.DeriveSolanaAddress(wallet, solanaHdPath)
	fmt.Println("Address for the Hd Path", solanaHdPath, "is ", addrs)

	// 3. CoinInfo
	info, err := util.GetCoinInfoFromHdPath(solanaHdPath)
	fmt.Println(info.ToString())
}
