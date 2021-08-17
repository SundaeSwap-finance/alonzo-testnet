#!/bin/bash

set -eu

sudo apt-get update && sudo apt-get install -y git binutils

mkdir -p src
git clone https://github.com/aws/efs-utils src/efs-utils
(cd src/efs-utils && ./build-deb.sh)

(cd src/efs-utils && sudo apt-get -y install ./build/amazon-efs-utils*deb)
