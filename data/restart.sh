#!/bin/bash


ssh cefa_2 /root/go/bin/stop.sh
ssh cefa_3 /root/go/bin/stop.sh

scp /root/go/bin/polo root@cefa_2:/root/go/bin/polo
scp /root/go/bin/polo root@cefa_3:/root/go/bin/polo

## cefa 2
## 登陆cefa 2
screen
polo --config-dir /root/go/bin/node2/ --data-dir /root/go/bin/node2/ --api-addr cefa_2:8670 --p2p-port 11238 --seed enode://adbe99b8382dff3235b5571007a39b949755afeb4e1b777c1c4684177a419616f2bee53139721681cf258d640c8ff21eed1d0b0e7344c9ef160ef99ff6f8e74a@172.17.132.133:55555
输入Ctrl A+D 将命令放入后台运行


## cefa 1
polo --config-dir /root/go/bin/node1/ --data-dir /root/go/bin/node1/ --api-addr cefa_1:8669 --p2p-port 11237 --seed enode://adbe99b8382dff3235b5571007a39b949755afeb4e1b777c1c4684177a419616f2bee53139721681cf258d640c8ff21eed1d0b0e7344c9ef160ef99ff6f8e74a@172.17.132.133:55555

## cefa 3
polo --config-dir /root/go/bin/node3/ --data-dir /root/go/bin/node3/ --api-addr cefa_3:8671 --p2p-port 11239 --seed enode://adbe99b8382dff3235b5571007a39b949755afeb4e1b777c1c4684177a419616f2bee53139721681cf258d640c8ff21eed1d0b0e7344c9ef160ef99ff6f8e74a@172.17.132.133:55555

ppllmm12



