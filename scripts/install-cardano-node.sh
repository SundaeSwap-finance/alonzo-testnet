#!/bin/bash

set -eu

export GIT_TAG=${GIT_TAG:=1.30.1}
export LD_LIBRARY_PATH="/lib:/usr/local/lib"


# install cardano-node
#
mkdir -p src
git clone https://github.com/input-output-hk/cardano-node src/cardano-node
cd src/cardano-node
git checkout ${GIT_TAG}

# build
#
. /home/ubuntu/.nix-profile/etc/profile.d/nix.sh
nix-shell --run "cabal build cardano-node"
nix-shell --run "cabal build cardano-cli"

# create symlinks to simplify execution
#
mkdir -p "${HOME}/bin"
ln -s "$(find "${HOME}/src/cardano-node/dist-newstyle" -type f -executable | grep -E 'cardano-cli$')" "${HOME}/bin/cardano-cli"
ln -s "$(find "${HOME}/src/cardano-node/dist-newstyle" -type f -executable | grep -E 'cardano-node$')" "${HOME}/bin/cardano-node"

