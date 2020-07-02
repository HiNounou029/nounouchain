# nounou

#### 介绍
macro nounou chain

#### build and install

* cd src/github.com/nounouchain
* go test  ./api/... ./nounou/... ./cmd/nounou/... ./consensus/... ./network/... ./storage/... ./vm/... ./trie/... ./state/... ./miner/... ./core/... ./common/...
* go install ./cmd/... 如果使用国密算法 go install -tags "gm" ./cmd/...

安装在 $GOPATH/bin/下

* ./genesis --datadir `配置文件目录` --approvercount `授权节点数量，默认值3` --authcount `出块节点数量，默认值3` --password `设置节点keystore文件对应的密码文件` 

运行genesis命令创建启动节点所需要的配置文件

配置目录 ./conf/下需要有一个node.json和genesis_cfg.json，node.json为一个authority的keystore文件，可由genesis命令产生

nodes目录下对每个authority生成了一个目录，里面有node.json和genesis_cfg.json

运行种子节点

./seed --keyfile seed.key  (seed.key 为 src/github.com/nounouchain/data/seed.key)

运行一个节点，示例：

polo --config-dir ./nodes/node1/ --data-dir ./nodes/node1/ --api-addr localhost:8669 --p2p-port 11237 --seed enode://adbe99b8382dff3235b5571007a39b949755afeb4e1b777c1c4684177a419616f2bee53139721681cf258d640c8ff21eed1d0b0e7344c9ef160ef99ff6f8e74a@127.0.0.1:55555 --accpwd 123456 --api-cors "*"

当前目录下需要有一个 conf.json （示例文件为 src/github.com/nounouchain/data/conf.json)


运行每个节点，比如：

polo --config-dir ./nodes/node1/ --data-dir ./nodes/node1/ --api-addr localhost:8669 --p2p-port 11237 --seed enode://adbe99b8382dff3235b5571007a39b949755afeb4e1b777c1c4684177a419616f2bee53139721681cf258d640c8ff21eed1d0b0e7344c9ef160ef99ff6f8e74a@127.0.0.1:55555 --accpwd 123456 --api-cors "*"

polo --config-dir ./nodes/node2/ --data-dir ./nodes/node2/ --api-addr localhost:8670 --p2p-port 11238 --seed enode://adbe99b8382dff3235b5571007a39b949755afeb4e1b777c1c4684177a419616f2bee53139721681cf258d640c8ff21eed1d0b0e7344c9ef160ef99ff6f8e74a@127.0.0.1:55555 --accpwd 123456 --api-cors "*"

polo --config-dir ./nodes/node3/ --data-dir ./nodes/node3/ --api-addr localhost:8671 --p2p-port 11239 --seed enode://adbe99b8382dff3235b5571007a39b949755afeb4e1b777c1c4684177a419616f2bee53139721681cf258d640c8ff21eed1d0b0e7344c9ef160ef99ff6f8e74a@127.0.0.1:55555 --accpwd 123456 --api-cors "*"



* 出块节点通过授权方式增加或者撤销
节点启动的配置文件目录，必须包含：genesis_cfg.json文件和node.json文件

./auth propose add --keystore "./nodes/approvers/approver1.json" --addr "0x83c5f4c311e94a07c95e83ec3bb5ced2bd4b4a2b" --id "daniel1983" --password "./data/pwd.txt"

./auth approve 0x3b7365511b4736b045f71af31cf63c80b3b294419a4c92cd786ed6211fb57c10 --keystore "./nodes/approvers/approver2.json" --password "./data/pwd.txt"

./auth execute 0x45eaae2872676e5f27589d6e880cb30cf342d0910862f04ed81fa5ce38bf042e --keystore "./nodes/approvers/approver2.json" --password "./data/pwd.txt"