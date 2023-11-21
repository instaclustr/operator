#!/bin/bash

IPADDR=1.1.1.1
CONFIGFILE=./tehst/test.yaml

#sed -i "s/.*broadcast_rpc_address.*/broadcast_rpc_address: ${IPADDR}/g" $CONFIGFILE
#sed -i "s/.*broadcast_address.*/broadcast_address: ${IPADDR}/g" $CONFIGFILE
#sed -i "s/.*listen_address.*/listen_address: ${IPADDR}/g" $CONFIGFILE
#sed -i "s/.*{seeds.*/ - {seeds: 'cassandra-headless.default.svc.cluster.local'}/g" $CONFIGFILE

while true
do
  echo "IN WHILE"
  if [ -f "$CONFIGFILE" ]; then
    sed -i "s/.*{seeds.*/ - {seeds: 'cassandra-headless.default.svc.cluster.local'}/g" $CONFIGFILE
    echo "BREAK"
    break
  fi
  echo "SLEEP"
  sleep 5
done

echo "DONE"







































