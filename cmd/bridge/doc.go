// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

/*

跨链合约与网关

bridge_server: 跨链网关程序，配置文件为 config.json，不需要配置ClientKeyPath，
	ServerKeyPath表示以该key对应的身份运行网关，该身份在两个链上都有余额，可以作为
	转账中间账户

cross_chain_client: 发送跨链请求的客户端，配置文件同样为config.json，不需要配置ServerKeyPath
	ClientKeyPath表示以该key对应的身份发起跨链请求

contract: 合约solidity代码bridge.sol，合约ABI接口bridge.json, 以及go客户端程序bridge_contract.go

*/
package bridge
