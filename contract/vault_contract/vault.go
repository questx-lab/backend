// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package vault_contract

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
	_ = abi.ConvertType
)

// VaultContractMetaData contains all meta data concerning the VaultContract contract.
var VaultContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"Code501\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"Code502\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"addSpender\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"balances\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"removeSpender\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"setAdmin\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferIn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"tokens\",\"type\":\"address[]\"},{\"internalType\":\"address[]\",\"name\":\"tos\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"amounts\",\"type\":\"uint256[]\"}],\"name\":\"transferInMultiple\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"dstChain\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferOut\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"dstChain\",\"type\":\"uint256\"}],\"name\":\"transferOutNative\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f80fd5b50335f806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506115028061005c5f395ff3fe60806040526004361061007a575f3560e01c8063c23f001f1161004d578063c23f001f1461011e578063cc54128a1461015a578063e4652f4914610176578063e7e31e7a1461019e5761007a565b8063704b6c021461007e578063754ff725146100a6578063814785eb146100ce5780638ce5877c146100f6575b5f80fd5b348015610089575f80fd5b506100a4600480360381019061009f9190610bec565b6101c6565b005b3480156100b1575f80fd5b506100cc60048036038101906100c79190610e5a565b610208565b005b3480156100d9575f80fd5b506100f460048036038101906100ef9190610efe565b610326565b005b348015610101575f80fd5b5061011c60048036038101906101179190610bec565b610497565b005b348015610129575f80fd5b50610144600480360381019061013f9190610f4e565b61057b565b6040516101519190610f9b565b60405180910390f35b610174600480360381019061016f9190610fb4565b61059b565b005b348015610181575f80fd5b5061019c60048036038101906101979190610fdf565b61061b565b005b3480156101a9575f80fd5b506101c460048036038101906101bf9190610bec565b6107e0565b005b805f806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050565b60015f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f9054906101000a900460ff16610291576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161028890611089565b60405180910390fd5b5f5b83518163ffffffff1610156103205761030d848263ffffffff16815181106102be576102bd6110a7565b5b6020026020010151848363ffffffff16815181106102df576102de6110a7565b5b6020026020010151848463ffffffff1681518110610300576102ff6110a7565b5b602002602001015161061b565b808061031890611110565b915050610293565b50505050565b60035f8381526020019081526020015f205f9054906101000a900460ff1615610384576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161037b90611185565b60405180910390fd5b8060025f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054106104135761040e8333836108c4565b610492565b8273ffffffffffffffffffffffffffffffffffffffff166323b872dd3330846040518463ffffffff1660e01b8152600401610450939291906111b2565b6020604051808303815f875af115801561046c573d5f803e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610490919061121c565b505b505050565b5f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610524576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161051b90611291565b60405180910390fd5b5f60015f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f6101000a81548160ff02191690831515021790555050565b6002602052815f5260405f20602052805f5260405f205f91509150505481565b60035f8281526020019081526020015f205f9054906101000a900460ff16156105f9576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105f090611185565b60405180910390fd5b6106187340000000000000000000000000000000000000003334610a80565b50565b60015f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f9054906101000a900460ff166106a4576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161069b90611089565b60405180910390fd5b808373ffffffffffffffffffffffffffffffffffffffff166370a08231306040518263ffffffff1660e01b81526004016106de91906112af565b602060405180830381865afa1580156106f9573d5f803e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061071d91906112dc565b106107a3578273ffffffffffffffffffffffffffffffffffffffff1663a9059cbb83836040518363ffffffff1660e01b815260040161075d929190611307565b6020604051808303815f875af1158015610779573d5f803e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061079d919061121c565b506107db565b6107ae838383610a80565b7f6a036bee1b01306b61370b348f57c9a7038a7bf8cb5d8a4cfbfb197b3f329e8360405160405180910390a15b505050565b5f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461086d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161086490611291565b60405180910390fd5b6001805f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f6101000a81548160ff02191690831515021790555050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610932576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161092990611378565b60405180910390fd5b8060025f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205410156109ed576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016109e4906113e0565b60405180910390fd5b8060025f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f828254610a7491906113fe565b92505081905550505050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610aee576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610ae59061147b565b60405180910390fd5b8060025f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f828254610b759190611499565b92505081905550505050565b5f604051905090565b5f80fd5b5f80fd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610bbb82610b92565b9050919050565b610bcb81610bb1565b8114610bd5575f80fd5b50565b5f81359050610be681610bc2565b92915050565b5f60208284031215610c0157610c00610b8a565b5b5f610c0e84828501610bd8565b91505092915050565b5f80fd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b610c6182610c1b565b810181811067ffffffffffffffff82111715610c8057610c7f610c2b565b5b80604052505050565b5f610c92610b81565b9050610c9e8282610c58565b919050565b5f67ffffffffffffffff821115610cbd57610cbc610c2b565b5b602082029050602081019050919050565b5f80fd5b5f610ce4610cdf84610ca3565b610c89565b90508083825260208201905060208402830185811115610d0757610d06610cce565b5b835b81811015610d305780610d1c8882610bd8565b845260208401935050602081019050610d09565b5050509392505050565b5f82601f830112610d4e57610d4d610c17565b5b8135610d5e848260208601610cd2565b91505092915050565b5f67ffffffffffffffff821115610d8157610d80610c2b565b5b602082029050602081019050919050565b5f819050919050565b610da481610d92565b8114610dae575f80fd5b50565b5f81359050610dbf81610d9b565b92915050565b5f610dd7610dd284610d67565b610c89565b90508083825260208201905060208402830185811115610dfa57610df9610cce565b5b835b81811015610e235780610e0f8882610db1565b845260208401935050602081019050610dfc565b5050509392505050565b5f82601f830112610e4157610e40610c17565b5b8135610e51848260208601610dc5565b91505092915050565b5f805f60608486031215610e7157610e70610b8a565b5b5f84013567ffffffffffffffff811115610e8e57610e8d610b8e565b5b610e9a86828701610d3a565b935050602084013567ffffffffffffffff811115610ebb57610eba610b8e565b5b610ec786828701610d3a565b925050604084013567ffffffffffffffff811115610ee857610ee7610b8e565b5b610ef486828701610e2d565b9150509250925092565b5f805f60608486031215610f1557610f14610b8a565b5b5f610f2286828701610bd8565b9350506020610f3386828701610db1565b9250506040610f4486828701610db1565b9150509250925092565b5f8060408385031215610f6457610f63610b8a565b5b5f610f7185828601610bd8565b9250506020610f8285828601610bd8565b9150509250929050565b610f9581610d92565b82525050565b5f602082019050610fae5f830184610f8c565b92915050565b5f60208284031215610fc957610fc8610b8a565b5b5f610fd684828501610db1565b91505092915050565b5f805f60608486031215610ff657610ff5610b8a565b5b5f61100386828701610bd8565b935050602061101486828701610bd8565b925050604061102586828701610db1565b9150509250925092565b5f82825260208201905092915050565b7f4e6f74207370656e6465723a20464f5242494444454e000000000000000000005f82015250565b5f61107360168361102f565b915061107e8261103f565b602082019050919050565b5f6020820190508181035f8301526110a081611067565b9050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f63ffffffff82169050919050565b5f61111a82611101565b915063ffffffff82036111305761112f6110d4565b5b600182019050919050565b7f434841494e5f49535f50415553454400000000000000000000000000000000005f82015250565b5f61116f600f8361102f565b915061117a8261113b565b602082019050919050565b5f6020820190508181035f83015261119c81611163565b9050919050565b6111ac81610bb1565b82525050565b5f6060820190506111c55f8301866111a3565b6111d260208301856111a3565b6111df6040830184610f8c565b949350505050565b5f8115159050919050565b6111fb816111e7565b8114611205575f80fd5b50565b5f81519050611216816111f2565b92915050565b5f6020828403121561123157611230610b8a565b5b5f61123e84828501611208565b91505092915050565b7f4e6f742061646d696e3a20464f5242494444454e0000000000000000000000005f82015250565b5f61127b60148361102f565b915061128682611247565b602082019050919050565b5f6020820190508181035f8301526112a88161126f565b9050919050565b5f6020820190506112c25f8301846111a3565b92915050565b5f815190506112d681610d9b565b92915050565b5f602082840312156112f1576112f0610b8a565b5b5f6112fe848285016112c8565b91505092915050565b5f60408201905061131a5f8301856111a3565b6113276020830184610f8c565b9392505050565b7f6465633a206164647265737320697320300000000000000000000000000000005f82015250565b5f61136260118361102f565b915061136d8261132e565b602082019050919050565b5f6020820190508181035f83015261138f81611356565b9050919050565b7f6465633a20616d6f756e7420657863656564732062616c616e636500000000005f82015250565b5f6113ca601b8361102f565b91506113d582611396565b602082019050919050565b5f6020820190508181035f8301526113f7816113be565b9050919050565b5f61140882610d92565b915061141383610d92565b925082820390508181111561142b5761142a6110d4565b5b92915050565b7f696e633a206164647265737320697320300000000000000000000000000000005f82015250565b5f61146560118361102f565b915061147082611431565b602082019050919050565b5f6020820190508181035f83015261149281611459565b9050919050565b5f6114a382610d92565b91506114ae83610d92565b92508282019050808211156114c6576114c56110d4565b5b9291505056fea264697066735822122094775ff5f9f9fcf7550280720216f8f37d5117492cc150d9659f529788cb997d64736f6c63430008140033",
}

// VaultContractABI is the input ABI used to generate the binding from.
// Deprecated: Use VaultContractMetaData.ABI instead.
var VaultContractABI = VaultContractMetaData.ABI

// VaultContractBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use VaultContractMetaData.Bin instead.
var VaultContractBin = VaultContractMetaData.Bin

// DeployVaultContract deploys a new Ethereum contract, binding an instance of VaultContract to it.
func DeployVaultContract(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *VaultContract, error) {
	parsed, err := VaultContractMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(VaultContractBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &VaultContract{VaultContractCaller: VaultContractCaller{contract: contract}, VaultContractTransactor: VaultContractTransactor{contract: contract}, VaultContractFilterer: VaultContractFilterer{contract: contract}}, nil
}

// VaultContract is an auto generated Go binding around an Ethereum contract.
type VaultContract struct {
	VaultContractCaller     // Read-only binding to the contract
	VaultContractTransactor // Write-only binding to the contract
	VaultContractFilterer   // Log filterer for contract events
}

// VaultContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type VaultContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VaultContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type VaultContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VaultContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type VaultContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VaultContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type VaultContractSession struct {
	Contract     *VaultContract    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// VaultContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type VaultContractCallerSession struct {
	Contract *VaultContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// VaultContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type VaultContractTransactorSession struct {
	Contract     *VaultContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// VaultContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type VaultContractRaw struct {
	Contract *VaultContract // Generic contract binding to access the raw methods on
}

// VaultContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type VaultContractCallerRaw struct {
	Contract *VaultContractCaller // Generic read-only contract binding to access the raw methods on
}

// VaultContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type VaultContractTransactorRaw struct {
	Contract *VaultContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewVaultContract creates a new instance of VaultContract, bound to a specific deployed contract.
func NewVaultContract(address common.Address, backend bind.ContractBackend) (*VaultContract, error) {
	contract, err := bindVaultContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &VaultContract{VaultContractCaller: VaultContractCaller{contract: contract}, VaultContractTransactor: VaultContractTransactor{contract: contract}, VaultContractFilterer: VaultContractFilterer{contract: contract}}, nil
}

// NewVaultContractCaller creates a new read-only instance of VaultContract, bound to a specific deployed contract.
func NewVaultContractCaller(address common.Address, caller bind.ContractCaller) (*VaultContractCaller, error) {
	contract, err := bindVaultContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &VaultContractCaller{contract: contract}, nil
}

// NewVaultContractTransactor creates a new write-only instance of VaultContract, bound to a specific deployed contract.
func NewVaultContractTransactor(address common.Address, transactor bind.ContractTransactor) (*VaultContractTransactor, error) {
	contract, err := bindVaultContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &VaultContractTransactor{contract: contract}, nil
}

// NewVaultContractFilterer creates a new log filterer instance of VaultContract, bound to a specific deployed contract.
func NewVaultContractFilterer(address common.Address, filterer bind.ContractFilterer) (*VaultContractFilterer, error) {
	contract, err := bindVaultContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &VaultContractFilterer{contract: contract}, nil
}

// bindVaultContract binds a generic wrapper to an already deployed contract.
func bindVaultContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := VaultContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_VaultContract *VaultContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _VaultContract.Contract.VaultContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_VaultContract *VaultContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _VaultContract.Contract.VaultContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_VaultContract *VaultContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _VaultContract.Contract.VaultContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_VaultContract *VaultContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _VaultContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_VaultContract *VaultContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _VaultContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_VaultContract *VaultContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _VaultContract.Contract.contract.Transact(opts, method, params...)
}

// Balances is a free data retrieval call binding the contract method 0xc23f001f.
//
// Solidity: function balances(address , address ) view returns(uint256)
func (_VaultContract *VaultContractCaller) Balances(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _VaultContract.contract.Call(opts, &out, "balances", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Balances is a free data retrieval call binding the contract method 0xc23f001f.
//
// Solidity: function balances(address , address ) view returns(uint256)
func (_VaultContract *VaultContractSession) Balances(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _VaultContract.Contract.Balances(&_VaultContract.CallOpts, arg0, arg1)
}

// Balances is a free data retrieval call binding the contract method 0xc23f001f.
//
// Solidity: function balances(address , address ) view returns(uint256)
func (_VaultContract *VaultContractCallerSession) Balances(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _VaultContract.Contract.Balances(&_VaultContract.CallOpts, arg0, arg1)
}

// AddSpender is a paid mutator transaction binding the contract method 0xe7e31e7a.
//
// Solidity: function addSpender(address spender) returns()
func (_VaultContract *VaultContractTransactor) AddSpender(opts *bind.TransactOpts, spender common.Address) (*types.Transaction, error) {
	return _VaultContract.contract.Transact(opts, "addSpender", spender)
}

// AddSpender is a paid mutator transaction binding the contract method 0xe7e31e7a.
//
// Solidity: function addSpender(address spender) returns()
func (_VaultContract *VaultContractSession) AddSpender(spender common.Address) (*types.Transaction, error) {
	return _VaultContract.Contract.AddSpender(&_VaultContract.TransactOpts, spender)
}

// AddSpender is a paid mutator transaction binding the contract method 0xe7e31e7a.
//
// Solidity: function addSpender(address spender) returns()
func (_VaultContract *VaultContractTransactorSession) AddSpender(spender common.Address) (*types.Transaction, error) {
	return _VaultContract.Contract.AddSpender(&_VaultContract.TransactOpts, spender)
}

// RemoveSpender is a paid mutator transaction binding the contract method 0x8ce5877c.
//
// Solidity: function removeSpender(address spender) returns()
func (_VaultContract *VaultContractTransactor) RemoveSpender(opts *bind.TransactOpts, spender common.Address) (*types.Transaction, error) {
	return _VaultContract.contract.Transact(opts, "removeSpender", spender)
}

// RemoveSpender is a paid mutator transaction binding the contract method 0x8ce5877c.
//
// Solidity: function removeSpender(address spender) returns()
func (_VaultContract *VaultContractSession) RemoveSpender(spender common.Address) (*types.Transaction, error) {
	return _VaultContract.Contract.RemoveSpender(&_VaultContract.TransactOpts, spender)
}

// RemoveSpender is a paid mutator transaction binding the contract method 0x8ce5877c.
//
// Solidity: function removeSpender(address spender) returns()
func (_VaultContract *VaultContractTransactorSession) RemoveSpender(spender common.Address) (*types.Transaction, error) {
	return _VaultContract.Contract.RemoveSpender(&_VaultContract.TransactOpts, spender)
}

// SetAdmin is a paid mutator transaction binding the contract method 0x704b6c02.
//
// Solidity: function setAdmin(address newAdmin) returns()
func (_VaultContract *VaultContractTransactor) SetAdmin(opts *bind.TransactOpts, newAdmin common.Address) (*types.Transaction, error) {
	return _VaultContract.contract.Transact(opts, "setAdmin", newAdmin)
}

// SetAdmin is a paid mutator transaction binding the contract method 0x704b6c02.
//
// Solidity: function setAdmin(address newAdmin) returns()
func (_VaultContract *VaultContractSession) SetAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _VaultContract.Contract.SetAdmin(&_VaultContract.TransactOpts, newAdmin)
}

// SetAdmin is a paid mutator transaction binding the contract method 0x704b6c02.
//
// Solidity: function setAdmin(address newAdmin) returns()
func (_VaultContract *VaultContractTransactorSession) SetAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _VaultContract.Contract.SetAdmin(&_VaultContract.TransactOpts, newAdmin)
}

// TransferIn is a paid mutator transaction binding the contract method 0xe4652f49.
//
// Solidity: function transferIn(address token, address to, uint256 amount) returns()
func (_VaultContract *VaultContractTransactor) TransferIn(opts *bind.TransactOpts, token common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _VaultContract.contract.Transact(opts, "transferIn", token, to, amount)
}

// TransferIn is a paid mutator transaction binding the contract method 0xe4652f49.
//
// Solidity: function transferIn(address token, address to, uint256 amount) returns()
func (_VaultContract *VaultContractSession) TransferIn(token common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _VaultContract.Contract.TransferIn(&_VaultContract.TransactOpts, token, to, amount)
}

// TransferIn is a paid mutator transaction binding the contract method 0xe4652f49.
//
// Solidity: function transferIn(address token, address to, uint256 amount) returns()
func (_VaultContract *VaultContractTransactorSession) TransferIn(token common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _VaultContract.Contract.TransferIn(&_VaultContract.TransactOpts, token, to, amount)
}

// TransferInMultiple is a paid mutator transaction binding the contract method 0x754ff725.
//
// Solidity: function transferInMultiple(address[] tokens, address[] tos, uint256[] amounts) returns()
func (_VaultContract *VaultContractTransactor) TransferInMultiple(opts *bind.TransactOpts, tokens []common.Address, tos []common.Address, amounts []*big.Int) (*types.Transaction, error) {
	return _VaultContract.contract.Transact(opts, "transferInMultiple", tokens, tos, amounts)
}

// TransferInMultiple is a paid mutator transaction binding the contract method 0x754ff725.
//
// Solidity: function transferInMultiple(address[] tokens, address[] tos, uint256[] amounts) returns()
func (_VaultContract *VaultContractSession) TransferInMultiple(tokens []common.Address, tos []common.Address, amounts []*big.Int) (*types.Transaction, error) {
	return _VaultContract.Contract.TransferInMultiple(&_VaultContract.TransactOpts, tokens, tos, amounts)
}

// TransferInMultiple is a paid mutator transaction binding the contract method 0x754ff725.
//
// Solidity: function transferInMultiple(address[] tokens, address[] tos, uint256[] amounts) returns()
func (_VaultContract *VaultContractTransactorSession) TransferInMultiple(tokens []common.Address, tos []common.Address, amounts []*big.Int) (*types.Transaction, error) {
	return _VaultContract.Contract.TransferInMultiple(&_VaultContract.TransactOpts, tokens, tos, amounts)
}

// TransferOut is a paid mutator transaction binding the contract method 0x814785eb.
//
// Solidity: function transferOut(address token, uint256 dstChain, uint256 amount) returns()
func (_VaultContract *VaultContractTransactor) TransferOut(opts *bind.TransactOpts, token common.Address, dstChain *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _VaultContract.contract.Transact(opts, "transferOut", token, dstChain, amount)
}

// TransferOut is a paid mutator transaction binding the contract method 0x814785eb.
//
// Solidity: function transferOut(address token, uint256 dstChain, uint256 amount) returns()
func (_VaultContract *VaultContractSession) TransferOut(token common.Address, dstChain *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _VaultContract.Contract.TransferOut(&_VaultContract.TransactOpts, token, dstChain, amount)
}

// TransferOut is a paid mutator transaction binding the contract method 0x814785eb.
//
// Solidity: function transferOut(address token, uint256 dstChain, uint256 amount) returns()
func (_VaultContract *VaultContractTransactorSession) TransferOut(token common.Address, dstChain *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _VaultContract.Contract.TransferOut(&_VaultContract.TransactOpts, token, dstChain, amount)
}

// TransferOutNative is a paid mutator transaction binding the contract method 0xcc54128a.
//
// Solidity: function transferOutNative(uint256 dstChain) payable returns()
func (_VaultContract *VaultContractTransactor) TransferOutNative(opts *bind.TransactOpts, dstChain *big.Int) (*types.Transaction, error) {
	return _VaultContract.contract.Transact(opts, "transferOutNative", dstChain)
}

// TransferOutNative is a paid mutator transaction binding the contract method 0xcc54128a.
//
// Solidity: function transferOutNative(uint256 dstChain) payable returns()
func (_VaultContract *VaultContractSession) TransferOutNative(dstChain *big.Int) (*types.Transaction, error) {
	return _VaultContract.Contract.TransferOutNative(&_VaultContract.TransactOpts, dstChain)
}

// TransferOutNative is a paid mutator transaction binding the contract method 0xcc54128a.
//
// Solidity: function transferOutNative(uint256 dstChain) payable returns()
func (_VaultContract *VaultContractTransactorSession) TransferOutNative(dstChain *big.Int) (*types.Transaction, error) {
	return _VaultContract.Contract.TransferOutNative(&_VaultContract.TransactOpts, dstChain)
}

// VaultContractCode501Iterator is returned from FilterCode501 and is used to iterate over the raw logs and unpacked data for Code501 events raised by the VaultContract contract.
type VaultContractCode501Iterator struct {
	Event *VaultContractCode501 // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VaultContractCode501Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VaultContractCode501)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VaultContractCode501)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VaultContractCode501Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VaultContractCode501Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VaultContractCode501 represents a Code501 event raised by the VaultContract contract.
type VaultContractCode501 struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterCode501 is a free log retrieval operation binding the contract event 0x6a036bee1b01306b61370b348f57c9a7038a7bf8cb5d8a4cfbfb197b3f329e83.
//
// Solidity: event Code501()
func (_VaultContract *VaultContractFilterer) FilterCode501(opts *bind.FilterOpts) (*VaultContractCode501Iterator, error) {

	logs, sub, err := _VaultContract.contract.FilterLogs(opts, "Code501")
	if err != nil {
		return nil, err
	}
	return &VaultContractCode501Iterator{contract: _VaultContract.contract, event: "Code501", logs: logs, sub: sub}, nil
}

// WatchCode501 is a free log subscription operation binding the contract event 0x6a036bee1b01306b61370b348f57c9a7038a7bf8cb5d8a4cfbfb197b3f329e83.
//
// Solidity: event Code501()
func (_VaultContract *VaultContractFilterer) WatchCode501(opts *bind.WatchOpts, sink chan<- *VaultContractCode501) (event.Subscription, error) {

	logs, sub, err := _VaultContract.contract.WatchLogs(opts, "Code501")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VaultContractCode501)
				if err := _VaultContract.contract.UnpackLog(event, "Code501", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCode501 is a log parse operation binding the contract event 0x6a036bee1b01306b61370b348f57c9a7038a7bf8cb5d8a4cfbfb197b3f329e83.
//
// Solidity: event Code501()
func (_VaultContract *VaultContractFilterer) ParseCode501(log types.Log) (*VaultContractCode501, error) {
	event := new(VaultContractCode501)
	if err := _VaultContract.contract.UnpackLog(event, "Code501", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// VaultContractCode502Iterator is returned from FilterCode502 and is used to iterate over the raw logs and unpacked data for Code502 events raised by the VaultContract contract.
type VaultContractCode502Iterator struct {
	Event *VaultContractCode502 // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VaultContractCode502Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VaultContractCode502)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VaultContractCode502)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VaultContractCode502Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VaultContractCode502Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VaultContractCode502 represents a Code502 event raised by the VaultContract contract.
type VaultContractCode502 struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterCode502 is a free log retrieval operation binding the contract event 0xe1ef72f1796988294cc9a59a2e3073a1079dc556e353a1ebf43b456b191161a9.
//
// Solidity: event Code502()
func (_VaultContract *VaultContractFilterer) FilterCode502(opts *bind.FilterOpts) (*VaultContractCode502Iterator, error) {

	logs, sub, err := _VaultContract.contract.FilterLogs(opts, "Code502")
	if err != nil {
		return nil, err
	}
	return &VaultContractCode502Iterator{contract: _VaultContract.contract, event: "Code502", logs: logs, sub: sub}, nil
}

// WatchCode502 is a free log subscription operation binding the contract event 0xe1ef72f1796988294cc9a59a2e3073a1079dc556e353a1ebf43b456b191161a9.
//
// Solidity: event Code502()
func (_VaultContract *VaultContractFilterer) WatchCode502(opts *bind.WatchOpts, sink chan<- *VaultContractCode502) (event.Subscription, error) {

	logs, sub, err := _VaultContract.contract.WatchLogs(opts, "Code502")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VaultContractCode502)
				if err := _VaultContract.contract.UnpackLog(event, "Code502", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCode502 is a log parse operation binding the contract event 0xe1ef72f1796988294cc9a59a2e3073a1079dc556e353a1ebf43b456b191161a9.
//
// Solidity: event Code502()
func (_VaultContract *VaultContractFilterer) ParseCode502(log types.Log) (*VaultContractCode502, error) {
	event := new(VaultContractCode502)
	if err := _VaultContract.contract.UnpackLog(event, "Code502", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
