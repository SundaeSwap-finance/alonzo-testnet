#!/bin/bash

set -eu

export DEBIAN_FRONTEND="noninteractive"

sudo apt-get install && sudo apt-get install -y tzdata curl

curl -L --output /tmp/go1.17.linux-amd64.tar.gz https://golang.org/dl/go1.17.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go1.17.linux-amd64.tar.gz
rm -f /tmp/go1.17.linux-amd64.tar.gz
sudo ln -s /usr/local/go/bin/go /usr/local/bin/go

mkdir -p "${HOME}/bin"
cd "/tmp/bootstrap"
go get ./...
go build -o "${HOME}/bin/bootstrap"

cat <<EOF >> "${HOME}/.bash_profile"

# set CARDANO_NODE_SOCKET_PATH to cardano-cli
#
export CARDANO_NODE_SOCKET_PATH=\${HOME}/alonzo-testnet/node-bft1/node.sock

# add ${HOME}/bin to path
#
export PATH="\${PATH}:\${HOME}/bin"

# aliases
#
alias ls="ls -sF --color"

EOF

if [ -f /tmp/restart-testnet.sh ] ; then
  cp /tmp/restart-testnet.sh "${HOME}/bin/restart-testnet.sh"
  chmod +x "${HOME}/bin/restart-testnet.sh"
fi


