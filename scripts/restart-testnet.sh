#!/bin/bash

# kill all existing cardano-node instances
#
killall -9 cardano-node

# restart instances
#
(nohup "${HOME}/alonzo-testnet/run/all.sh" 2>&1) > /dev/null &
