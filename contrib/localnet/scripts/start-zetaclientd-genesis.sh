#!/bin/bash

/usr/sbin/sshd

HOSTNAME=$(hostname)

cp  /root/preparams/PreParams_$HOSTNAME.json /root/preParams.json
num=$(echo $HOSTNAME | tr -dc '0-9')
node="zetacore$num"
#mv  /root/zetacored/zetacored_$node /root/.zetacored
#mv /root/tss/$HOSTNAME /root/.tss

echo "Wait for zetacore to exchange genesis file"
sleep 30
operator=$(cat $HOME/.zetacored/os.json | jq '.ObserverAddress' )
operatorAddress=$(echo "$operator" | tr -d '"')
echo "operatorAddress: $operatorAddress"
echo "Start zetaclientd"
if [ $HOSTNAME == "zetaclient0" ]
then
    rm ~/.tss/*
    MYIP=$(/sbin/ip -o -4 addr list eth0 | awk '{print $4}' | cut -d/ -f1)
    zetaclientd init  --zetacore-url zetacore0 --chain-id athens_101-1 --operator "$operatorAddress"  --log-format=text --public-ip "$MYIP"
    zetaclientd start
else
  num=$(echo $HOSTNAME | tr -dc '0-9')
  node="zetacore$num"
  MYIP=$(/sbin/ip -o -4 addr list eth0 | awk '{print $4}' | cut -d/ -f1)
  SEED=$(curl --retry 10 --retry-delay 5 --retry-connrefused  -s zetaclient0:8123/p2p)
  rm ~/.tss/*
  zetaclientd init --peer /ip4/172.20.0.21/tcp/6668/p2p/"$SEED" --zetacore-url "$node" --chain-id athens_101-1 --operator "$operatorAddress" --log-format=text --public-ip "$MYIP"
  zetaclientd start
fi
