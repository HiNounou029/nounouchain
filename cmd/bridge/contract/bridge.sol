// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

pragma solidity ^0.4.25;

contract Bridge {

    // message sent to other chains
    // srcChain is implicit (this chain), destChain is map key
    struct SendMsg {
        address srcAddr;
        address destAddr;
        uint256 value;
        string reqMsg; // request message
        // final response status from other chains, -1 unknown, 1 success, 0 fail
        int8 status;
        string resMsg; // response message
    }

    // message received from other chains
    // srcChain is map key, destChain is implicit (this chain)
    struct RecvMsg {
        address srcAddr;
        address destAddr;
        uint256 value;
        bool success;   // result of the message
        string resMsg; // response message
    }

    // msgs sent to other chains, key is destChain
    mapping (uint256 => SendMsg[]) public sendMsgsByDestChain;
    mapping (uint256 => uint256) public lastSentMsgIdByDestChain;
    mapping (uint256 => uint256) public lastAckedMsgIdByDestChain;

    // received messages from other chains, by srcChain
    mapping (uint256 => RecvMsg[]) public recvMsgsBySrcChain;
    mapping (uint256 => uint256) public lastMsgIdBySrcChain;

    // request event, contains only ID information to save storage
    event Request(uint256 destChain, uint256 id);
    event Response(uint256 srcChain, uint256 id, bool success, string resMsg);

    address owner;

    constructor(address _owner) public {
        owner = _owner;
    }

    // send message  to destChain/destAddr from srcAddr
    // srcAddr could be different from msg.sender, if different, usually called by
    // other contract to delegate their user
    // called by normal user or other contract in the same chain
    function send(address srcAddr, uint256 destChain, address destAddr,
        string reqMsg) public payable {
        require(srcAddr != address(0) && destChain>0 && destAddr!=address(0));
        require(srcAddr.balance >= msg.value, "no enough money");
        sendMsgsByDestChain[destChain].push(SendMsg({srcAddr: srcAddr,
            destAddr: destAddr, value: msg.value, reqMsg: reqMsg, status: -1, resMsg: ""}));
        uint256 lastMsgId = lastSentMsgIdByDestChain[destChain];
        lastSentMsgIdByDestChain[destChain] = lastMsgId + 1; // first message id is 1
        emit Request(destChain, lastMsgId + 1);
    }

    // ack request message to destChain/id, result is success
    // ack id must be expected
    // called by relay server
    function ack(uint256 destChain, uint256 id, bool success, string resMsg) public {
        require(msg.sender == owner, "only owner is allowed to call ack");
        require(destChain>0);
        uint256 expectedId = lastAckedMsgIdByDestChain[destChain] + 1;
        require(id == expectedId, "id is not expected");
        int8 status = success ? int8(1) : int8(0);
        sendMsgsByDestChain[destChain][id-1].status = status;
        sendMsgsByDestChain[destChain][id-1].resMsg = resMsg;
        lastAckedMsgIdByDestChain[destChain] = id;
        //TOOD: add general message handler
        if(!success){
            // TODO: if failed, rollback value transer
        }
    }

    // forward message from relay server to dest chain
    // called by relay server
    function forward(uint256 srcChain, uint256 id,
        address srcAddr,  address destAddr, string reqMsg) public payable {
        require(msg.sender == owner, "only owner is allowed to relay message");
        // ensure id is expected, prevent duplicate message and double spent
        uint256 expectedMsgId = lastMsgIdBySrcChain[srcChain] + 1;
        require(id == expectedMsgId, "id is not expected");

        bool success = true;
        //TODO: add general message handler
        if(msg.value>0){
            //TODO: limit the gas value to prevent reentry attack
            success = destAddr.send(msg.value);
        }
        lastMsgIdBySrcChain[srcChain] = id;
        string memory resMsg  = reqMsg; //TODO: to get real response message
        recvMsgsBySrcChain[srcChain].push(RecvMsg({srcAddr: srcAddr,
            destAddr: destAddr, value: msg.value, success: success, resMsg: resMsg}));

        emit Response(srcChain, id, success, resMsg);
    }

    // called by relay server to withdraw money from contract to himself
    function withdraw(uint256 value) public {
        require(value>0);
        require(msg.sender == owner, "only owner is allowed to withdraw");
        require(address(this).balance>=value, "no enough money");

        owner.transfer(address(this).balance);
    }
}