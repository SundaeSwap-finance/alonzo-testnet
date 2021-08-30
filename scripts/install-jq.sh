#!/bin/bash

set -eu

# install jq
#
sudo apt-get update && sudo apt-get install -y jq

# install yq
#
sudo apt install -y python3-pip
sudo pip install yq
