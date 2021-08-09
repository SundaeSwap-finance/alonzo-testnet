#!/bin/bash

set -eu

sudo yum install -y golang

mkdir -p "${HOME}/bin"
cd "/tmp/bootstrap"
go get ./...
go build -o "${HOME}/bin/bootstrap"

cat <<EOF >> "${HOME}/.bash_profile"

export CARDANO_NODE_SOCKET_PATH=\${HOME}/alonzo-testnet/node-bft1/node.sock

EOF

if [ -f /tmp/restart-testnet.sh ] ; then
  cp /tmp/restart-testnet.sh "${HOME}/bin/restart-testnet.sh"
  chmod +x "${HOME}/bin/restart-testnet.sh"
fi


