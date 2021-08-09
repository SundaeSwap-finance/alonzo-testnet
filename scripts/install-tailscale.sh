#!/bin/bash

set -eu

# Some versions of Amazon Linux 2 are unable to verify signatures from Tailscale,
# due to a bug in the gnupg2 package.
sudo yum update -y gnupg2

sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://pkgs.tailscale.com/stable/amazon-linux/2/tailscale.repo
sudo yum install -y tailscale

sudo systemctl enable --now tailscaled
