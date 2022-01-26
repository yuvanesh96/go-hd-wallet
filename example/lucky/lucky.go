package main

import (
	"context"
	"fmt"
	"hdwallet"
	"hdwallet/networks"
	"hdwallet/util"
	"log"
	"math/big"
	"math/rand"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/tyler-smith/go-bip39"
)

const (
	DbPath                           = "./.data"
	CheckedDdPath                    = DbPath + "/checked"
	LuckyDbPath                      = DbPath + "/lucky"
	LocalGethIpcEndPoint             = "/home/yuvanesh/.ethereum/geth.ipc"
	BalanceWorkerPoolCount           = 100
	MnemonicGeneratorWorkerPoolCount = 25
	AnalyseBalanceWorkerPoolCount    = 25
	ChannelBufferLength              = 10000
	PrintStatusDuration              = 60 // In seconds
	BalanceBackoffDuration           = 5.0
	BalanceMaxTries                  = 20
)

var (
	globalCheckedDbInstance *leveldb.DB
	globalLuckyDbInstance   *leveldb.DB
)

func initDb() {
	var err error
	globalCheckedDbInstance, err = leveldb.OpenFile(CheckedDdPath, nil)
	if err != nil {
		log.Fatal("Unable to create DB")
	}

	globalLuckyDbInstance, err = leveldb.OpenFile(LuckyDbPath, nil)
	if err != nil {
		log.Fatal("Unable to create DB")
	}
}

type BalanceInfo struct {
	Mnemonic string
	EthAddr  string
	Balance  int64
}

type MnemonicInfo struct {
	Mnemonic string
	EthAddr  string
}

var (
	SyncWrite = &opt.WriteOptions{
		NoWriteMerge: false,
		Sync:         true,
	}
	start                 time.Time
	commChan              chan struct{}
	totalMnemonicsChecked int
	mnemonicInfoChan      = make(chan MnemonicInfo, ChannelBufferLength)
	balanceInfoChan       = make(chan BalanceInfo, ChannelBufferLength)
)

// workGenerateMnemonicInfo generates the random set of valid mnemonics and writes it
// in to the corresponding channel.
func workGenerateMnemonicInfo() {
	for true {
		entropy, err := bip39.NewEntropy(128)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		wallet, err := hdwallet.NewRandomHdWallet(entropy)
		fmt.Println("On Mnemonic : " + wallet.GetMnemonic())
		found, err := globalCheckedDbInstance.Has([]byte(wallet.GetMnemonic()), nil)
		if found == true {
			fmt.Println("Mnemonic checked already")
			continue
		}
		ethHdPath, err := util.GetBaseHdPath(util.TypeEth)
		ethHdPath = ethHdPath + "0"
		addr, err := networks.DeriveEthereumAddress(wallet, ethHdPath)
		if err != nil {
			fmt.Println("Error generating the address for " + ethHdPath)
			return
		}
		fmt.Println("Eth Address for " + ethHdPath + " : " + addr)
		mnemonicInfoChan <- MnemonicInfo{
			Mnemonic: wallet.GetMnemonic(),
			EthAddr:  addr,
		}
	}
}

func findBalanceWithExponentialBackoff(client *ethclient.Client, account common.Address) (*big.Int, error) {
	var balance *big.Int
	var err error
	delay := BalanceBackoffDuration
	for i := 0; i < BalanceMaxTries; i++ {
		balance, err = client.BalanceAt(context.Background(), account, nil)
		if err != nil {
			fmt.Println("[IMP]Error getting balance" + err.Error())
			delay = delay + rand.Float64() // To avoid thunder being heard.
			fmt.Println(fmt.Sprintf("[IMP] Waiting for %f  before next try", delay))
			//Exponential backoff
			delay *= 2
			time.Sleep(time.Duration(delay) * time.Second)
			continue
		}
		return balance, err
	}
	return nil, err
}

// workGetEthBalanceWithLocalGeth consumes the mnemonic from the corresponding channel
// and fetches the eth balance for the same. The balnce info has been feded into the balance
// channel for further processing.
func workGetEthBalanceWithLocalGeth() {
	client, err := ethclient.Dial(LocalGethIpcEndPoint)
	defer client.Close()

	if err != nil {
		fmt.Println("[IMP]Error in connecting to local geth ipc " + err.Error())
		return
	}

	for mnemonicInfo := range mnemonicInfoChan {
		account := common.HexToAddress(mnemonicInfo.EthAddr)
		balance, err := findBalanceWithExponentialBackoff(client, account)
		if err != nil {
			fmt.Println("[IMP] Not able to fecth balance after multiple tries. Quitting")
			return
		}
		balanceInfoChan <- BalanceInfo{
			Mnemonic: mnemonicInfo.Mnemonic,
			EthAddr:  mnemonicInfo.EthAddr,
			Balance:  balance.Int64(),
		}
	}

}

// workAnalyseAndRecordBalanceInfo analyses the balance info records mnemonics with
// any non zero balance.
func workAnalyseAndRecordBalanceInfo() {
	for balanceInfo := range balanceInfoChan {
		//fmt.Println(fmt.Sprintf("Balance of %s = %d  x 10^(-9) ETH", balanceInfo.EthAddr, balanceInfo.Balance))

		if balanceInfo.Balance > 0 {
			fmt.Println("[IMP]Lucky protocol Succeeded")
			balHexString := fmt.Sprintf("0x%x", balanceInfo.Balance)
			globalLuckyDbInstance.Put([]byte(balanceInfo.Mnemonic), []byte(balHexString), SyncWrite)
		}
		globalCheckedDbInstance.Put([]byte(balanceInfo.Mnemonic), []byte(balanceInfo.EthAddr), SyncWrite)
	}
}

func getDbKeysCount(db *leveldb.DB) int {
	iter := db.NewIterator(nil, nil)
	keyCount := 0
	for iter.Next() {
		keyCount++
	}
	return keyCount
}

func printDbCount(db *leveldb.DB) {
	fmt.Println("[IMP] Keys Found : " + strconv.Itoa(getDbKeysCount(db)))
}

func closeDb() {
	defer func(db *leveldb.DB) {
		err := db.Close()
		if err != nil {
			fmt.Println("err closing the db")
		}
	}(globalCheckedDbInstance)

	defer func(db *leveldb.DB) {
		err := db.Close()
		if err != nil {
			fmt.Println("err closing the db")
		}
	}(globalLuckyDbInstance)
}

func getCurrTime() string {
	dt := time.Now()
	return dt.Format("01-02-2006 15:04:05")
}

func printStatus() {

	fmt.Println("[IMP]Printing the Check DB")
	totalMnemonicsCheckedLatest := getDbKeysCount(globalCheckedDbInstance)
	fmt.Println(fmt.Sprintf("[IMP] Keys Checked so far = %e ", float64(totalMnemonicsCheckedLatest)))

	fmt.Println("[IMP]Printing the lucky DB")
	printDbCount(globalLuckyDbInstance)
	mnemonicsAttempted := totalMnemonicsCheckedLatest - totalMnemonicsChecked

	elapsed := time.Since(start)

	fmt.Printf(
		getCurrTime()+
			" [IMP] Rate of luck protocol is %f mnemonics/sec \n",
		(float64(mnemonicsAttempted) / elapsed.Seconds()))
	start = time.Now()
	totalMnemonicsChecked = totalMnemonicsCheckedLatest
}

func printStatusRoutine() {
	totalMnemonicsChecked = getDbKeysCount(globalCheckedDbInstance)
	printStatus()
	for true {
		time.Sleep(PrintStatusDuration * time.Second)
		printStatus()
	}

}

func runMnemonicWorkers() {
	for i := 0; i < MnemonicGeneratorWorkerPoolCount; i++ {
		go workGenerateMnemonicInfo()
	}
}

func runGetBalanceWorkers() {
	for i := 0; i < BalanceWorkerPoolCount; i++ {
		go workGetEthBalanceWithLocalGeth()
	}
}

func runAnalyseBalanceWorkers() {
	for i := 0; i < AnalyseBalanceWorkerPoolCount; i++ {
		go workAnalyseAndRecordBalanceInfo()
	}
}

func run() {
	fmt.Println("[IMP]<<<-------- Attempting luck protocol --------->>> ")
	start = time.Now()
	go printStatusRoutine()
	go runMnemonicWorkers()
	go runGetBalanceWorkers()
	go runAnalyseBalanceWorkers()

	// Wait until any signal from common channel or wait Indefinitely.
	<-commChan
}

func main() {
	initDb()
	defer closeDb()

	run()
}
