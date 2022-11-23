// SPDX-License-Identifier: MIT
pragma solidity ^0.8.17;

contract GasGuzzle {
    mapping(bytes32 => uint256) private pacificOcean; // where we store all the guzzled gas

    function guzzle(uint256 gasToBurn) external {
        uint256 startGas = gasleft();
        // burn gas
        while (startGas - gasleft() < gasToBurn) {
            pacificOcean[blockhash(block.number)] = block.difficulty;
        }
    }
}
