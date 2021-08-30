alonzo-testnet
-------------

`alonzo-testnet` simplifies deploying a private alonzo testnet.

SundaeSwap heavily leverages AWS and consequently, this tooling has 
been optimized to run in AWS.

### Build the alonzo-purple AMI

To create your testnet, begin by creating a custom AWS AMI for the testnet
instance.  

```bash
packer build cardano-node.pkr.hcl
```

The AMI contains:

* a complete nix development environment
* cardano-node built from source using the alonzo-purple-1.0.1 tag

### Quickstart

Once the AMI has been built, launch an instance that has sufficient memory and CPU 
to run 3 nodes.  It is recommended to use `t3.xlarge` at a minimum.  The instance
is built using `Ubuntu 20.04` so the login will be `ubuntu`.

After you login, you can start your testnet as follows:

```bash
bootstrap mkfiles
(nohup "${HOME}/alonzo-testnet/run/all.sh" 2>&1) > /dev/null &
```

### Options

#### Use a custom testnet-magic

```bash
bootstrap mkfiles -m 31415
```

#### Increase the max transaction size

```bash
bootstrap mkfiles -n maxTxSize=32768
```

#### Increase the transaction fees

```bash
bootstrap mkfiles -n minFeeA=100 -n minFeeB=12
```

#### SundaeSwap - no fee env

```bash
bootstrap mkfiles -m 31415 -n maxTxSize=65384 -n maxLovelaceSupply=10000000000000000 -n minFeeA=0
```

#### alonzo-purple rc3

```bash
bootstrap mkfiles -m 31415 -n maxTxSize=16384 -n maxLovelaceSupply=10000000000000000 -n minFeeA=44 -n minFeeB=155381
```

#### alonzo-purple rc2 with 10m epoch

```bash
bootstrap mkfiles -m 31415 -n maxTxSize=16384 -n maxLovelaceSupply=10000000000000000 -n minFeeA=44 -n minFeeB=155381 -n epochLength=600
```

### Additional Packages

The following additional packages have also been installed to facilitate
integration into the development environment

* tailscale client, https://tailscale.com
* AWS EFS client, https://aws.amazon.com/efs/
