package util

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
)

type CoinInfo struct {
	name     CoinName
	coinType CoinType
}

var (
	InvalidCoinInfo = CoinInfo{
		name:     NameInvalid,
		coinType: TypeInvalid,
	}
)

func (c *CoinInfo) getName() CoinName {
	return c.name
}

func (c *CoinInfo) getCoinType() CoinType {
	return c.coinType
}

func (c *CoinInfo) ToString() string {
	return fmt.Sprintf("[CoinName:%s,CoinType:%d]", c.getName(), c.getCoinType())
}

func GetCoinInfoFromHdPath(hdPath string) (CoinInfo, error) {
	coinType, err := getCoinType(hdPath)
	if err != nil {
		return InvalidCoinInfo, err
	}
	switch coinType {
	case TypeEth:
		return CoinInfo{
			name:     NameEth,
			coinType: TypeEth,
		}, nil
	case TypeSol:
		return CoinInfo{
			name:     NameSol,
			coinType: TypeSol,
		}, nil
	}
	return InvalidCoinInfo, err
}

func getCoinType(hdPath string) (CoinType, error) {
	components, err := accounts.ParseDerivationPath(hdPath)
	if err != nil {
		return TypeInvalid, err
	}
	// m / purpose' / coin_type' / account' / change / address_index
	componentCoinType := components[1]
	coinType := componentCoinType ^ 0x80000000
	fmt.Println("CoinType:", coinType)
	return CoinType(coinType), nil
}
