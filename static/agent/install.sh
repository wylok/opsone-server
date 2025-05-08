#!/bin/bash
AgentPath="/opt/opsone"
ApiUrl="http://<remote_ip>/api/v1"
[ ! -f /usr/bin/wget ] && yum -y install wget
[ -f $AgentPath/opsone-agent.pid ] && kill -9 $(cat $AgentPath/opsone-agent.pid)
[ -f $AgentPath/opsone-dog.pid ] && kill -9 $(cat $AgentPath/opsone-dog.pid)
pkill -9 opsone-dog
pkill -9 opsone-agent
[ -d $AgentPath ] && rm -rf $AgentPath
mkdir -p $AgentPath
for ((i=1; i<=5; i++))
do
  wget -O $AgentPath/opsone-dog --timeout=60  $ApiUrl/ag/opsone-dog
  wget -O $AgentPath/config.ini --timeout=60  $ApiUrl/conf/config.ini
  [ -f $AgentPath/config.ini ] && [ -f $AgentPath/opsone-dog ] && chmod +x $AgentPath/opsone-dog && break
  sleep 30
done
$AgentPath/opsone-dog start
sleep 3
ls -l $AgentPath
exit 0