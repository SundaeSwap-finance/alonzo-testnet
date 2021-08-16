packer {
  required_plugins {
    amazon = {
      version = ">= 0.0.2"
      source = "github.com/hashicorp/amazon"
    }
  }
}

source "amazon-ebs" "amazon" {
  ami_name = "alonzo-purple-{{timestamp}}"

  associate_public_ip_address = true
  instance_type = "m5.2xlarge"
  region = var.region
  ssh_username = "ec2-user"
  subnet_id = var.subnet_id
  vpc_id = var.vpc_id

  launch_block_device_mappings {
    volume_type = "gp2"
    device_name = "/dev/xvda"
    volume_size = 250
    delete_on_termination = true
  }

  source_ami_filter {
    filters = {
      name = "*amzn2-ami-hvm-*"
      root-device-type = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners = [
      "amazon"
    ]
  }
}

build {
  name = "alonzo-purple-1.0.1"
  sources = [
    "source.amazon-ebs.amazon"
  ]

  provisioner "file" {
    source = "cmd/bootstrap"
    destination = "/tmp/bootstrap"
  }
  provisioner "file" {
    source = "scripts/restart-testnet.sh"
    destination = "/tmp/restart-testnet.sh"
  }

  provisioner "shell" {
    environment_vars = [
      "LD_LIBRARY_PATH=/lib:/usr/local/lib",
    ]
    scripts = [
      "scripts/install-tailscale.sh",
      "scripts/install-nix.sh",
      "scripts/install-cardano-node.sh",
      "scripts/install-docker.sh",
      "scripts/install-bootstrap.sh",
      "scripts/install-jq.sh"
    ]
  }
}

variable "region" {
  type = string
  default = "us-east-2"
}

variable "subnet_id" {
  type = string
  default = "subnet-0575ab4529d4a3128"
}

variable "vpc_id" {
  type = string
  default = "vpc-0e91af5d62d4fd92c"
}

