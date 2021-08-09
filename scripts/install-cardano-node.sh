#!/bin/bash

set -eu

export GIT_TAG=${GIT_TAG:=alonzo-purple-1.0.1}
export LD_LIBRARY_PATH="/lib:/usr/local/lib"


# install cardano-node
#
git clone https://github.com/input-output-hk/cardano-node "${HOME}/src/cardano-node"
cd "${HOME}/src/cardano-node"
git checkout ${GIT_TAG}

# build
#
. /home/ec2-user/.nix-profile/etc/profile.d/nix.sh
nix-shell --run "cabal build cardano-node"
nix-shell --run "cabal build cardano-cli"

# create symlinks to simplify execution
#
mkdir -p "${HOME}/bin"
ln -s "$(find "${HOME}/src/cardano-node/dist-newstyle" -type f -executable | grep -E 'cardano-cli$')" "${HOME}/bin/cardano-cli"
ln -s "$(find "${HOME}/src/cardano-node/dist-newstyle" -type f -executable | grep -E 'cardano-node$')" "${HOME}/bin/cardano-node"

