// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package guzzle

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// GuzzleMetaData contains all meta data concerning the Guzzle contract.
var GuzzleMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"gasToBurn\",\"type\":\"uint256\"}],\"name\":\"guzzle\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"pacificOcean\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b50610267806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c8063c54d02ed1461003b578063cb1ed5181461006b575b600080fd5b61005560048036038101906100509190610114565b610087565b604051610062919061015a565b60405180910390f35b610085600480360381019061008091906101a1565b61009f565b005b60006020528060005260406000206000915090505481565b60005a90505b815a826100b291906101fd565b10156100d5574460008043408152602001908152602001600020819055506100a5565b5050565b600080fd5b6000819050919050565b6100f1816100de565b81146100fc57600080fd5b50565b60008135905061010e816100e8565b92915050565b60006020828403121561012a576101296100d9565b5b6000610138848285016100ff565b91505092915050565b6000819050919050565b61015481610141565b82525050565b600060208201905061016f600083018461014b565b92915050565b61017e81610141565b811461018957600080fd5b50565b60008135905061019b81610175565b92915050565b6000602082840312156101b7576101b66100d9565b5b60006101c58482850161018c565b91505092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061020882610141565b915061021383610141565b925082820390508181111561022b5761022a6101ce565b5b9291505056fea26469706673582212203da552c2e1dc2083d44abedb68f2bb91f844e02739c4435357f4992d30000dce64736f6c63430008110033",
}

// GuzzleABI is the input ABI used to generate the binding from.
// Deprecated: Use GuzzleMetaData.ABI instead.
var GuzzleABI = GuzzleMetaData.ABI

// GuzzleBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use GuzzleMetaData.Bin instead.
var GuzzleBin = GuzzleMetaData.Bin

// DeployGuzzle deploys a new Ethereum contract, binding an instance of Guzzle to it.
func DeployGuzzle(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Guzzle, error) {
	parsed, err := GuzzleMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(GuzzleBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Guzzle{GuzzleCaller: GuzzleCaller{contract: contract}, GuzzleTransactor: GuzzleTransactor{contract: contract}, GuzzleFilterer: GuzzleFilterer{contract: contract}}, nil
}

// Guzzle is an auto generated Go binding around an Ethereum contract.
type Guzzle struct {
	GuzzleCaller     // Read-only binding to the contract
	GuzzleTransactor // Write-only binding to the contract
	GuzzleFilterer   // Log filterer for contract events
}

// GuzzleCaller is an auto generated read-only Go binding around an Ethereum contract.
type GuzzleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GuzzleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GuzzleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GuzzleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GuzzleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GuzzleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GuzzleSession struct {
	Contract     *Guzzle           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GuzzleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GuzzleCallerSession struct {
	Contract *GuzzleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// GuzzleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GuzzleTransactorSession struct {
	Contract     *GuzzleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GuzzleRaw is an auto generated low-level Go binding around an Ethereum contract.
type GuzzleRaw struct {
	Contract *Guzzle // Generic contract binding to access the raw methods on
}

// GuzzleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GuzzleCallerRaw struct {
	Contract *GuzzleCaller // Generic read-only contract binding to access the raw methods on
}

// GuzzleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GuzzleTransactorRaw struct {
	Contract *GuzzleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGuzzle creates a new instance of Guzzle, bound to a specific deployed contract.
func NewGuzzle(address common.Address, backend bind.ContractBackend) (*Guzzle, error) {
	contract, err := bindGuzzle(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Guzzle{GuzzleCaller: GuzzleCaller{contract: contract}, GuzzleTransactor: GuzzleTransactor{contract: contract}, GuzzleFilterer: GuzzleFilterer{contract: contract}}, nil
}

// NewGuzzleCaller creates a new read-only instance of Guzzle, bound to a specific deployed contract.
func NewGuzzleCaller(address common.Address, caller bind.ContractCaller) (*GuzzleCaller, error) {
	contract, err := bindGuzzle(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GuzzleCaller{contract: contract}, nil
}

// NewGuzzleTransactor creates a new write-only instance of Guzzle, bound to a specific deployed contract.
func NewGuzzleTransactor(address common.Address, transactor bind.ContractTransactor) (*GuzzleTransactor, error) {
	contract, err := bindGuzzle(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GuzzleTransactor{contract: contract}, nil
}

// NewGuzzleFilterer creates a new log filterer instance of Guzzle, bound to a specific deployed contract.
func NewGuzzleFilterer(address common.Address, filterer bind.ContractFilterer) (*GuzzleFilterer, error) {
	contract, err := bindGuzzle(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GuzzleFilterer{contract: contract}, nil
}

// bindGuzzle binds a generic wrapper to an already deployed contract.
func bindGuzzle(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(GuzzleABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Guzzle *GuzzleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Guzzle.Contract.GuzzleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Guzzle *GuzzleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Guzzle.Contract.GuzzleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Guzzle *GuzzleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Guzzle.Contract.GuzzleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Guzzle *GuzzleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Guzzle.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Guzzle *GuzzleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Guzzle.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Guzzle *GuzzleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Guzzle.Contract.contract.Transact(opts, method, params...)
}

// PacificOcean is a free data retrieval call binding the contract method 0xc54d02ed.
//
// Solidity: function pacificOcean(bytes32 ) view returns(uint256)
func (_Guzzle *GuzzleCaller) PacificOcean(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Guzzle.contract.Call(opts, &out, "pacificOcean", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PacificOcean is a free data retrieval call binding the contract method 0xc54d02ed.
//
// Solidity: function pacificOcean(bytes32 ) view returns(uint256)
func (_Guzzle *GuzzleSession) PacificOcean(arg0 [32]byte) (*big.Int, error) {
	return _Guzzle.Contract.PacificOcean(&_Guzzle.CallOpts, arg0)
}

// PacificOcean is a free data retrieval call binding the contract method 0xc54d02ed.
//
// Solidity: function pacificOcean(bytes32 ) view returns(uint256)
func (_Guzzle *GuzzleCallerSession) PacificOcean(arg0 [32]byte) (*big.Int, error) {
	return _Guzzle.Contract.PacificOcean(&_Guzzle.CallOpts, arg0)
}

// Guzzle is a paid mutator transaction binding the contract method 0xcb1ed518.
//
// Solidity: function guzzle(uint256 gasToBurn) returns()
func (_Guzzle *GuzzleTransactor) Guzzle(opts *bind.TransactOpts, gasToBurn *big.Int) (*types.Transaction, error) {
	return _Guzzle.contract.Transact(opts, "guzzle", gasToBurn)
}

// Guzzle is a paid mutator transaction binding the contract method 0xcb1ed518.
//
// Solidity: function guzzle(uint256 gasToBurn) returns()
func (_Guzzle *GuzzleSession) Guzzle(gasToBurn *big.Int) (*types.Transaction, error) {
	return _Guzzle.Contract.Guzzle(&_Guzzle.TransactOpts, gasToBurn)
}

// Guzzle is a paid mutator transaction binding the contract method 0xcb1ed518.
//
// Solidity: function guzzle(uint256 gasToBurn) returns()
func (_Guzzle *GuzzleTransactorSession) Guzzle(gasToBurn *big.Int) (*types.Transaction, error) {
	return _Guzzle.Contract.Guzzle(&_Guzzle.TransactOpts, gasToBurn)
}
