package util

type Curve string

const (
	Secp256k1 = Curve("secp256k1")
	Ed25519   = Curve("ed25519")
)
