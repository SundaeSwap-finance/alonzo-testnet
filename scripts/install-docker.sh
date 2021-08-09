#!/bin/bash

set -eu

sudo amazon-linux-extras install docker
sudo service docker start
sudo usermod -a -G docker ec2-user
