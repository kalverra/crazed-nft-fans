package client

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

// PrivateKeyToAddress is a handy converter for an ecdsa private key to a usable eth address
func PrivateKeyToAddress(privateKey *ecdsa.PrivateKey) (common.Address, error) {
	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, fmt.Errorf(
			"error converting public key to ecdsa format. private key: %s public key: %s", privateKey, privateKey.Public())
	}
	return crypto.PubkeyToAddress(*publicKeyECDSA), nil
}

// Credit to kimxilxyong: https://github.com/ethereum/go-ethereum/issues/21221#issuecomment-802092592

// EtherToWei converts an ETH amount to wei
func EtherToWei(eth *big.Float) *big.Int {
	truncInt, _ := eth.Int(nil)
	truncInt = new(big.Int).Mul(truncInt, big.NewInt(params.Ether))
	fracStr := strings.Split(fmt.Sprintf("%.18f", eth), ".")[1]
	fracStr += strings.Repeat("0", 18-len(fracStr))
	fracInt, _ := new(big.Int).SetString(fracStr, 10)
	wei := new(big.Int).Add(truncInt, fracInt)
	return wei
}

// WeiToEther converts a wei amount to eth
func WeiToEther(wei *big.Int) *big.Float {
	f := new(big.Float)
	f.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	f.SetMode(big.ToNearestEven)
	fWei := new(big.Float)
	fWei.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	fWei.SetMode(big.ToNearestEven)
	return f.Quo(fWei.SetInt(wei), big.NewFloat(params.Ether))
}

// WeiToGwei converts Wei amounts to GWei
func WeiToGwei(wei *big.Int) *big.Float {
	floatWei := new(big.Float).SetInt(wei)
	floatGWei := big.NewFloat(params.GWei)
	return new(big.Float).Quo(floatWei, floatGWei)
}
