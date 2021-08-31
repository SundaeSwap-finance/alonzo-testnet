packer {
  required_plugins {
    amazon = {
      version = ">= 0.0.2"
      source = "github.com/hashicorp/amazon"
    }
  }
}

source "amazon-ebs" "amazon" {
  ami_name = "alonzo-{{timestamp}}"

  associate_public_ip_address = true
  instance_type = "m5.2xlarge"
  region = var.region
  ssh_username = "ubuntu"
  subnet_id = var.subnet_id
  vpc_id = var.vpc_id

  ami_groups = ["all"]

  launch_block_device_mappings {
    volume_type = "gp2"
    device_name = "/dev/sda1"
    volume_size = 500
    delete_on_termination = true
  }

  run_tag {
    key = "Name"
    value = "alonzo-testnet"
  }
  
  run_tag {
    key = "sundaeswap:name"
    value = "alonzo-testnet"
  }
  
  run_tag {
    key = "sundaeswap:ami_id"
    value = "{{ .SourceAMI }}"
  }

  run_tag {
    key = "sundaeswap:ami_name"
    value = "{{ .SourceAMIName }}"
  }

  run_tag {
    key = "sundaeswap:version"
    value = var.version
  }

  run_volume_tag {
    name = "sundaeswap:name"
    value = "alonzo-testnet"
  }
  
  run_volume_tag {
    name = "sundaeswap:ami_id"
    value = "{{ .SourceAMI }}"
  }

  run_volume_tag {
    name = "sundaeswap:ami_name"
    value = "{{ .SourceAMIName }}"
  }

  run_volume_tag {
    name = "sundaeswap:version"
    value = var.version
  }

  source_ami_filter {
    filters = {
      "virtualization-type": "hvm",
      "architecture": "x86_64",
      "name": "ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*",
      "block-device-mapping.volume-type": "gp2",
      "root-device-type": "ebs"
    }
    owners = [
      "099720109477"
    ]
    most_recent = true
  }
}

build {
  name = "alonzo"
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
      "scripts/install-common.sh",
      "scripts/install-tailscale.sh",
      "scripts/install-libsodium.sh",
      "scripts/install-nix.sh",
      "scripts/install-stack.sh",
      "scripts/install-docker.sh",
      "scripts/install-cardano-node.sh",
      "scripts/install-bootstrap.sh",
      "scripts/install-jq.sh",
      "scripts/install-postgresql.sh",
    ]
  }
}

variable "region" {
  type = string
}

variable "subnet_id" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "version" {
  type = string
  default = "latest"
}


