// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./erc20.sol";

contract Vault {
    address private admin;
    mapping(address => bool) spenders;
    address private constant native =
        0x4000000000000000000000000000000000000000;
    mapping(address => mapping(address => uint256)) public balances;
    mapping(uint256 => bool) notPausedChains;

    // The vault does not have enough balance to transfer token to recipient. Temporarily increases
    // user's balance for later withdrawal.
    event Code501();
    // Retry transfer fails.
    event Code502();

    constructor() {
        admin = msg.sender;
    }

    //* for authentication

    function setAdmin(address newAdmin) external {
        admin = newAdmin;
    }

    function addSpender(address spender) external onlyAdmin {
        spenders[spender] = true;
    }

    function removeSpender(address spender) external onlyAdmin {
        spenders[spender] = false;
    }

    modifier onlySpender() {
        require(spenders[msg.sender], "Not spender: FORBIDDEN");
        _;
    }

    modifier onlyAdmin() {
        require(msg.sender == admin, "Not admin: FORBIDDEN");
        _;
    }

    function _inc(address token, address acc, uint256 amount) internal {
        require(acc != address(0), "inc: address is 0");
        balances[token][acc] += amount;
    }

    function _dec(address token, address account, uint256 amount) internal {
        require(account != address(0), "dec: address is 0");
        require(
            balances[token][account] >= amount,
            "dec: amount exceeds balance"
        );
        balances[token][account] -= amount;
    }

    function transferInMultiple(
        address[] memory tokens,
        address[] memory tos,
        uint256[] memory amounts
    ) external onlySpender {
        for (uint32 i = 0; i < tokens.length; i++) {
            transferIn(tokens[i], tos[i], amounts[i]);
        }
    }

    function transferIn(
        address token,
        address to,
        uint256 amount
    ) public onlySpender {
        if (IERC20(token).balanceOf(address(this)) >= amount) {
            IERC20(token).transfer(to, amount);
        } else {
            _inc(token, to, amount);
            emit Code501();
        }
    }

    function transferOut(
        address token,
        uint256 dstChain,
        uint256 amount
    ) public {
        require(!notPausedChains[dstChain], "CHAIN_IS_PAUSED");
        if (balances[token][msg.sender] >= amount) {
            _dec(token, msg.sender, amount);
        } else {
            IERC20(token).transferFrom(msg.sender, address(this), amount);
        }
    }

    function transferOutNative(uint256 dstChain) external payable {
        require(!notPausedChains[dstChain], "CHAIN_IS_PAUSED");
        _inc(native, msg.sender, msg.value);
    }
}
