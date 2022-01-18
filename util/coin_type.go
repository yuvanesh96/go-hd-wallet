package util

type CoinType int
type CoinName string

const (
	NameInvalid = CoinName("invalid")
	TypeInvalid = CoinType(-1)
	NameEth     = CoinName("Ethereum")
	TypeEth     = CoinType(60)
	NameSol     = CoinName("Solana")
	TypeSol     = CoinType(501)
)
