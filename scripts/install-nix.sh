#!/bin/bash

set -eu

export LD_LIBRARY_PATH="/lib:/usr/local/lib"


# install common packages
#
sudo yum install -y git bash curl wget make g++ libtool autoconf automake musl


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


# install nix
#
curl -L https://nixos.org/nix/install | sh


# install plutus
#
cat <<END > /tmp/nix.conf
max-jobs            = 6
cores               = 0
trusted-users       = root
keep-derivations    = true
keep-outputs        = true
substituters        = https://hydra.iohk.io https://iohk.cachix.org https://cache.nixos.org/
trusted-public-keys = hydra.iohk.io:f/Ea+s+dFdN+3Y/G+FDgSq+a5NEWhJGzdjvKNGv0/EQ= iohk.cachix.org-1:DpRUyj7h7V830dp/i6Nti+NEO2/nhblbov/8MW7Rqoo= cache.nixos.org-1:6NCHdD59X431o0gWypbMrAURkbJ16ZPMQFGspcDShjY=
END
sudo mkdir -p /etc/nix
sudo mv /tmp/nix.conf /etc/nix/nix.conf


# install cabal
#
. /home/ec2-user/.nix-profile/etc/profile.d/nix.sh
nix-env -i cabal-install
cabal update
