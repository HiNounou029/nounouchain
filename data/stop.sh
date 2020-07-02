#!/bin/bash

ps -ef | grep polo | grep -v grep| awk '{print $2}' | xargs kill -2

sleep 2


#scp /root/go/bin/stop.sh root@cefa_2:/root/go/bin/stop.sh
#scp /root/go/bin/stop.sh root@cefa_3:/root/go/bin/stop.sh
