#!/bin/bash

set -eu

# install common packages
#
sudo apt-get update && sudo apt-get install -y git bash curl wget vim libtool build-essential


# install stack
#
curl -sSL https://get.haskellstack.org/ | sh

cat <<END >> "${HOME}/.bash_profile"

# include stack binaries in path
export PATH="\${PATH}:\${HOME}/.local/bin"

END