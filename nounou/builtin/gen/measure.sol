// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

pragma solidity 0.4.24;

/// @title to measure basic gas usage for external call.
contract Measure {
    function outer() public view {
        this.inner();
    }
    function inner() public pure {}    
}