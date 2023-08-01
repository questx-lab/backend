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

// Transaction is an auto generated low-level Go binding around an user-defined struct.
type Transaction struct {
	Token  common.Address
	From   common.Address
	To     common.Address
	Amount *big.Int
}

// VaultContractMetaData contains all meta data concerning the VaultContract contract.
var VaultContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"Code501\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"addSpender\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_acc\",\"type\":\"string\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"balances\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_acc\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"removeSpender\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structTransaction\",\"name\":\"transaction\",\"type\":\"tuple\"}],\"name\":\"transferIn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structTransaction[]\",\"name\":\"txs\",\"type\":\"tuple[]\"}],\"name\":\"transferInMultiple\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f80fd5b50335f806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506001805f805f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f6101000a81548160ff021916908315150217905550611127806100d05f395ff3fe608060405234801561000f575f80fd5b506004361061007b575f3560e01c80638ce5877c116100595780638ce5877c146100e7578063b9b092c814610103578063e7e31e7a14610133578063f3db9fe51461014f5761007b565b80630d48b7651461007f5780636c6573901461009b5780638892ebce146100b7575b5f80fd5b610099600480360381019061009491906109e0565b61016b565b005b6100b560048036038101906100b09190610ad3565b610335565b005b6100d160048036038101906100cc9190610bca565b61040f565b6040516100de9190610c33565b60405180910390f35b61010160048036038101906100fc9190610c4c565b610447565b005b61011d60048036038101906101189190610c77565b61052b565b60405161012a9190610c33565b60405180910390f35b61014d60048036038101906101489190610c4c565b61053e565b005b61016960048036038101906101649190610cd1565b610622565b005b60015f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f9054906101000a900460ff166101f4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016101eb90610d97565b60405180910390fd5b8060600151815f015173ffffffffffffffffffffffffffffffffffffffff166370a08231306040518263ffffffff1660e01b81526004016102359190610dc4565b602060405180830381865afa158015610250573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906102749190610df1565b1061030557805f015173ffffffffffffffffffffffffffffffffffffffff1663a9059cbb826040015183606001516040518363ffffffff1660e01b81526004016102bf929190610e1c565b6020604051808303815f875af11580156102db573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906102ff9190610e78565b50610332565b7f6a036bee1b01306b61370b348f57c9a7038a7bf8cb5d8a4cfbfb197b3f329e8360405160405180910390a15b50565b60015f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f9054906101000a900460ff166103be576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016103b590610d97565b60405180910390fd5b5f5b81518163ffffffff16101561040b576103f8828263ffffffff16815181106103eb576103ea610ea3565b5b602002602001015161016b565b808061040390610f0c565b9150506103c0565b5050565b600282805160208101820180518482526020830160208501208183528095505050505050602052805f5260405f205f91509150505481565b5f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146104d4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016104cb90610f81565b60405180910390fd5b5f60015f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f6101000a81548160ff02191690831515021790555050565b5f610536838361076a565b905092915050565b5f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146105cb576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105c290610f81565b60405180910390fd5b6001805f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f6101000a81548160ff02191690831515021790555050565b808373ffffffffffffffffffffffffffffffffffffffff166370a08231336040518263ffffffff1660e01b815260040161065c9190610dc4565b602060405180830381865afa158015610677573d5f803e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061069b9190610df1565b10156106dc576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016106d390610fe9565b60405180910390fd5b8273ffffffffffffffffffffffffffffffffffffffff166323b872dd3330846040518463ffffffff1660e01b815260040161071993929190611007565b6020604051808303815f875af1158015610735573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906107599190610e78565b506107658383836107cd565b505050565b5f60028260405161077b91906110a8565b90815260200160405180910390205f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905092915050565b806002836040516107de91906110a8565b90815260200160405180910390205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825461083591906110be565b92505081905550505050565b5f604051905090565b5f80fd5b5f80fd5b5f80fd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b61089c82610856565b810181811067ffffffffffffffff821117156108bb576108ba610866565b5b80604052505050565b5f6108cd610841565b90506108d98282610893565b919050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610907826108de565b9050919050565b610917816108fd565b8114610921575f80fd5b50565b5f813590506109328161090e565b92915050565b5f819050919050565b61094a81610938565b8114610954575f80fd5b50565b5f8135905061096581610941565b92915050565b5f608082840312156109805761097f610852565b5b61098a60806108c4565b90505f61099984828501610924565b5f8301525060206109ac84828501610924565b60208301525060406109c084828501610924565b60408301525060606109d484828501610957565b60608301525092915050565b5f608082840312156109f5576109f461084a565b5b5f610a028482850161096b565b91505092915050565b5f80fd5b5f67ffffffffffffffff821115610a2957610a28610866565b5b602082029050602081019050919050565b5f80fd5b5f610a50610a4b84610a0f565b6108c4565b90508083825260208201905060808402830185811115610a7357610a72610a3a565b5b835b81811015610a9c5780610a88888261096b565b845260208401935050608081019050610a75565b5050509392505050565b5f82601f830112610aba57610ab9610a0b565b5b8135610aca848260208601610a3e565b91505092915050565b5f60208284031215610ae857610ae761084a565b5b5f82013567ffffffffffffffff811115610b0557610b0461084e565b5b610b1184828501610aa6565b91505092915050565b5f80fd5b5f67ffffffffffffffff821115610b3857610b37610866565b5b610b4182610856565b9050602081019050919050565b828183375f83830152505050565b5f610b6e610b6984610b1e565b6108c4565b905082815260208101848484011115610b8a57610b89610b1a565b5b610b95848285610b4e565b509392505050565b5f82601f830112610bb157610bb0610a0b565b5b8135610bc1848260208601610b5c565b91505092915050565b5f8060408385031215610be057610bdf61084a565b5b5f83013567ffffffffffffffff811115610bfd57610bfc61084e565b5b610c0985828601610b9d565b9250506020610c1a85828601610924565b9150509250929050565b610c2d81610938565b82525050565b5f602082019050610c465f830184610c24565b92915050565b5f60208284031215610c6157610c6061084a565b5b5f610c6e84828501610924565b91505092915050565b5f8060408385031215610c8d57610c8c61084a565b5b5f610c9a85828601610924565b925050602083013567ffffffffffffffff811115610cbb57610cba61084e565b5b610cc785828601610b9d565b9150509250929050565b5f805f60608486031215610ce857610ce761084a565b5b5f610cf586828701610924565b935050602084013567ffffffffffffffff811115610d1657610d1561084e565b5b610d2286828701610b9d565b9250506040610d3386828701610957565b9150509250925092565b5f82825260208201905092915050565b7f4e6f74207370656e6465723a20464f5242494444454e000000000000000000005f82015250565b5f610d81601683610d3d565b9150610d8c82610d4d565b602082019050919050565b5f6020820190508181035f830152610dae81610d75565b9050919050565b610dbe816108fd565b82525050565b5f602082019050610dd75f830184610db5565b92915050565b5f81519050610deb81610941565b92915050565b5f60208284031215610e0657610e0561084a565b5b5f610e1384828501610ddd565b91505092915050565b5f604082019050610e2f5f830185610db5565b610e3c6020830184610c24565b9392505050565b5f8115159050919050565b610e5781610e43565b8114610e61575f80fd5b50565b5f81519050610e7281610e4e565b92915050565b5f60208284031215610e8d57610e8c61084a565b5b5f610e9a84828501610e64565b91505092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f63ffffffff82169050919050565b5f610f1682610efd565b915063ffffffff8203610f2c57610f2b610ed0565b5b600182019050919050565b7f4e6f742061646d696e3a20464f5242494444454e0000000000000000000000005f82015250565b5f610f6b601483610d3d565b9150610f7682610f37565b602082019050919050565b5f6020820190508181035f830152610f9881610f5f565b9050919050565b7f6465706f7369743a2073656e64657220657863656564732062616c616e6365005f82015250565b5f610fd3601f83610d3d565b9150610fde82610f9f565b602082019050919050565b5f6020820190508181035f83015261100081610fc7565b9050919050565b5f60608201905061101a5f830186610db5565b6110276020830185610db5565b6110346040830184610c24565b949350505050565b5f81519050919050565b5f81905092915050565b5f5b8381101561106d578082015181840152602081019050611052565b5f8484015250505050565b5f6110828261103c565b61108c8185611046565b935061109c818560208601611050565b80840191505092915050565b5f6110b38284611078565b915081905092915050565b5f6110c882610938565b91506110d383610938565b92508282019050808211156110eb576110ea610ed0565b5b9291505056fea2646970667358221220265d42fd675d583513f205a63ec88c6b3c40c9697d00d0ac064b7f65290a604364736f6c63430008140033",
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

// BalanceOf is a free data retrieval call binding the contract method 0xb9b092c8.
//
// Solidity: function balanceOf(address _token, string _acc) view returns(uint256)
func (_VaultContract *VaultContractCaller) BalanceOf(opts *bind.CallOpts, _token common.Address, _acc string) (*big.Int, error) {
	var out []interface{}
	err := _VaultContract.contract.Call(opts, &out, "balanceOf", _token, _acc)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0xb9b092c8.
//
// Solidity: function balanceOf(address _token, string _acc) view returns(uint256)
func (_VaultContract *VaultContractSession) BalanceOf(_token common.Address, _acc string) (*big.Int, error) {
	return _VaultContract.Contract.BalanceOf(&_VaultContract.CallOpts, _token, _acc)
}

// BalanceOf is a free data retrieval call binding the contract method 0xb9b092c8.
//
// Solidity: function balanceOf(address _token, string _acc) view returns(uint256)
func (_VaultContract *VaultContractCallerSession) BalanceOf(_token common.Address, _acc string) (*big.Int, error) {
	return _VaultContract.Contract.BalanceOf(&_VaultContract.CallOpts, _token, _acc)
}

// Balances is a free data retrieval call binding the contract method 0x8892ebce.
//
// Solidity: function balances(string , address ) view returns(uint256)
func (_VaultContract *VaultContractCaller) Balances(opts *bind.CallOpts, arg0 string, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _VaultContract.contract.Call(opts, &out, "balances", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Balances is a free data retrieval call binding the contract method 0x8892ebce.
//
// Solidity: function balances(string , address ) view returns(uint256)
func (_VaultContract *VaultContractSession) Balances(arg0 string, arg1 common.Address) (*big.Int, error) {
	return _VaultContract.Contract.Balances(&_VaultContract.CallOpts, arg0, arg1)
}

// Balances is a free data retrieval call binding the contract method 0x8892ebce.
//
// Solidity: function balances(string , address ) view returns(uint256)
func (_VaultContract *VaultContractCallerSession) Balances(arg0 string, arg1 common.Address) (*big.Int, error) {
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

// Deposit is a paid mutator transaction binding the contract method 0xf3db9fe5.
//
// Solidity: function deposit(address _token, string _acc, uint256 _amount) returns()
func (_VaultContract *VaultContractTransactor) Deposit(opts *bind.TransactOpts, _token common.Address, _acc string, _amount *big.Int) (*types.Transaction, error) {
	return _VaultContract.contract.Transact(opts, "deposit", _token, _acc, _amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xf3db9fe5.
//
// Solidity: function deposit(address _token, string _acc, uint256 _amount) returns()
func (_VaultContract *VaultContractSession) Deposit(_token common.Address, _acc string, _amount *big.Int) (*types.Transaction, error) {
	return _VaultContract.Contract.Deposit(&_VaultContract.TransactOpts, _token, _acc, _amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xf3db9fe5.
//
// Solidity: function deposit(address _token, string _acc, uint256 _amount) returns()
func (_VaultContract *VaultContractTransactorSession) Deposit(_token common.Address, _acc string, _amount *big.Int) (*types.Transaction, error) {
	return _VaultContract.Contract.Deposit(&_VaultContract.TransactOpts, _token, _acc, _amount)
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

// TransferIn is a paid mutator transaction binding the contract method 0x0d48b765.
//
// Solidity: function transferIn((address,address,address,uint256) transaction) returns()
func (_VaultContract *VaultContractTransactor) TransferIn(opts *bind.TransactOpts, transaction Transaction) (*types.Transaction, error) {
	return _VaultContract.contract.Transact(opts, "transferIn", transaction)
}

// TransferIn is a paid mutator transaction binding the contract method 0x0d48b765.
//
// Solidity: function transferIn((address,address,address,uint256) transaction) returns()
func (_VaultContract *VaultContractSession) TransferIn(transaction Transaction) (*types.Transaction, error) {
	return _VaultContract.Contract.TransferIn(&_VaultContract.TransactOpts, transaction)
}

// TransferIn is a paid mutator transaction binding the contract method 0x0d48b765.
//
// Solidity: function transferIn((address,address,address,uint256) transaction) returns()
func (_VaultContract *VaultContractTransactorSession) TransferIn(transaction Transaction) (*types.Transaction, error) {
	return _VaultContract.Contract.TransferIn(&_VaultContract.TransactOpts, transaction)
}

// TransferInMultiple is a paid mutator transaction binding the contract method 0x6c657390.
//
// Solidity: function transferInMultiple((address,address,address,uint256)[] txs) returns()
func (_VaultContract *VaultContractTransactor) TransferInMultiple(opts *bind.TransactOpts, txs []Transaction) (*types.Transaction, error) {
	return _VaultContract.contract.Transact(opts, "transferInMultiple", txs)
}

// TransferInMultiple is a paid mutator transaction binding the contract method 0x6c657390.
//
// Solidity: function transferInMultiple((address,address,address,uint256)[] txs) returns()
func (_VaultContract *VaultContractSession) TransferInMultiple(txs []Transaction) (*types.Transaction, error) {
	return _VaultContract.Contract.TransferInMultiple(&_VaultContract.TransactOpts, txs)
}

// TransferInMultiple is a paid mutator transaction binding the contract method 0x6c657390.
//
// Solidity: function transferInMultiple((address,address,address,uint256)[] txs) returns()
func (_VaultContract *VaultContractTransactorSession) TransferInMultiple(txs []Transaction) (*types.Transaction, error) {
	return _VaultContract.Contract.TransferInMultiple(&_VaultContract.TransactOpts, txs)
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
