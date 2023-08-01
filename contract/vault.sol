// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./erc20.sol";

struct Transaction {
    address token;
    address from;
    address to;
    uint256 amount;
}

struct Balance {
    address token;
    uint256 amount;
}

contract Vault {
    address private admin;
    mapping(address => bool) spenders;
    mapping(string => mapping(address => uint256)) public balances;
    event Code501();

    constructor() {
        admin = msg.sender;
        spenders[admin] = true;
    }

    //////// start manage contract ////////

    function addSpender(address spender) external onlyAdmin {
        spenders[spender] = true;
    }

    function removeSpender(address spender) external onlyAdmin {
        spenders[spender] = false;
    }

    //////// end manage contract ////////

    //////// start authenticate ////////
    modifier onlySpender() {
        require(spenders[msg.sender], "Not spender: FORBIDDEN");
        _;
    }

    modifier onlyAdmin() {
        require(msg.sender == admin, "Not admin: FORBIDDEN");
        _;
    }

    ////////// end authenticate ////////

    //////// start internal function ////////
    function _inc(
        address _token,
        string memory _acc,
        uint256 _amount
    ) internal {
        balances[_acc][_token] += _amount;
    }

    function _dec(
        address _token,
        string memory _acc,
        uint256 _amount
    ) internal {
        require(
            balances[_acc][_token] >= _amount,
            "dec: amount exceeds balance"
        );
        balances[_acc][_token] -= _amount;
    }

    function _balanceOf(
        address _token,
        string memory _acc
    ) internal view returns (uint256) {
        return balances[_acc][_token];
    }

    //////// end internal function ////////

    //////// start transfer function ////////
    function transferInMultiple(Transaction[] memory txs) external onlySpender {
        for (uint32 i = 0; i < txs.length; i++) {
            transferIn(txs[i]);
        }
    }

    function transferIn(Transaction memory transaction) public onlySpender {
        if (
            IERC20(transaction.token).balanceOf(address(this)) >=
            transaction.amount
        ) {
            IERC20(transaction.token).transfer(
                transaction.to,
                transaction.amount
            );
        } else {
            emit Code501();
        }
    }

    //////// end transfer function ////////

    //////// working with balance ////////
    function deposit(
        address _token,
        string memory _acc,
        uint256 _amount
    ) external {
        require(
            IERC20(_token).balanceOf(msg.sender) >= _amount,
            "deposit: sender exceeds balance"
        );
        IERC20(_token).transferFrom(msg.sender, address(this), _amount);
        _inc(_token, _acc, _amount);
    }

    function balanceOf(
        address _token,
        string memory _acc
    ) external view returns (uint256) {
        return _balanceOf(_token, _acc);
    }
}
