#!/bin/bash

set -eu

sleep 15

export DEBIAN_FRONTEND="noninteractive"

sudo apt-get update -y
sudo apt-get install -y tree curl wget vim unzip


# install awscli
#
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o /tmp/awscliv2.zip
(cd /tmp && unzip awscliv2.zip)
(cd /tmp && sudo ./aws/install)
rm -rf /tmp/aws /tmp/awscliv2.zip


# vim tab spacing to 2
#
cat <<EOF > "${HOME}/.exrc"
set ts=2
EOF

