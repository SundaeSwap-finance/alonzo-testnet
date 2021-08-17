#!/bin/bash

set -eu

# install common packages
#
sudo apt-get update && sudo apt-get install -y git bash curl wget vim libtool build-essential


# install IOHK version of libsodium
#
mkdir -p src
git clone https://github.com/input-output-hk/libsodium src/libsodium
cd src/libsodium
git checkout 66f017f1
./autogen.sh
./configure
make
sudo make install

cat <<END >> "${HOME}/.bash_profile"

# include libsodium in path
export LD_LIBRARY_PATH="\${LD_LIBRARY_PATH}:/usr/local/lib"

END
