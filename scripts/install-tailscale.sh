#!/bin/bash

set -eu

# Some versions of Amazon Linux 2 are unable to verify signatures from Tailscale,
# due to a bug in the gnupg2 package.
curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/focal.gpg | sudo apt-key add -
curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/focal.list | sudo tee /etc/apt/sources.list.d/tailscale.list

sudo apt-get update && sudo apt-get install -y tailscale net-tools
