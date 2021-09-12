#!/bin/bash

set -eu

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
. /home/ubuntu/.nix-profile/etc/profile.d/nix.sh
nix-env -i cabal-install


# set upper limit on GHC compiler
#
cat <<EOF >> "${HOME}/.bash_profile"

# limit ghc max heap
#
export GHCRTS='-M12g'

EOF
