alonzo-testnet
-------------

`alonzo-testnet` simplifies deploying a private alonzo testnet.

SundaeSwap heavily leverages AWS and consequently, this tooling has 
been optimized to run in AWS.

### Launch Stacks

Launch AWS Cloudformation stack in specified region.  Each stack will spin up a single ec2 instance with a
private alonzo testnet.

Region Name               | Region         |         |
:---                      | :---           | :---   |
US East (N. Virginia)     | us-east-1      | [![Launch Stack](https://cdn.sundaeswap.finance/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/new?stackName=alonzo-testnet&templateURL=https://s3.amazonaws.com/sundaeswap-oss/alonzo-testnet/alonzo-testnet-us-east-1.template)
US East (Ohio)            | us-east-2      | [![Launch Stack](https://cdn.sundaeswap.finance/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=us-east-2#/stacks/new?stackName=alonzo-testnet&templateURL=https://s3.amazonaws.com/sundaeswap-oss/alonzo-testnet/alonzo-testnet-us-east-2.template)
US West (N. California)   | us-west-1      | [![Launch Stack](https://cdn.sundaeswap.finance/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=us-west-1#/stacks/new?stackName=alonzo-testnet&templateURL=https://s3.amazonaws.com/sundaeswap-oss/alonzo-testnet/alonzo-testnet-us-west-1.template)
US West (Oregon)          | us-west-2      | [![Launch Stack](https://cdn.sundaeswap.finance/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=us-west-2#/stacks/new?stackName=alonzo-testnet&templateURL=https://s3.amazonaws.com/sundaeswap-oss/alonzo-testnet/alonzo-testnet-us-west-2.template)
Asia Pacific (Mumbai)     | ap-south-1     | [![Launch Stack](https://cdn.sundaeswap.finance/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-1#/stacks/new?stackName=alonzo-testnet&templateURL=https://s3.amazonaws.com/sundaeswap-oss/alonzo-testnet/alonzo-testnet-ap-south-1.template)
Asia Pacific (Osaka)      | ap-northeast-3 | [![Launch Stack](https://cdn.sundaeswap.finance/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-1#/stacks/new?stackName=alonzo-testnet&templateURL=https://s3.amazonaws.com/sundaeswap-oss/alonzo-testnet/alonzo-testnet-ap-northeast-3.template)
Asia Pacific (Seoul)      | ap-northeast-2 | [![Launch Stack](https://cdn.sundaeswap.finance/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-1#/stacks/new?stackName=alonzo-testnet&templateURL=https://s3.amazonaws.com/sundaeswap-oss/alonzo-testnet/alonzo-testnet-ap-northeast-2.template)
Asia Pacific (Tokyo)      | ap-northeast-1 | [![Launch Stack](https://cdn.sundaeswap.finance/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-1#/stacks/new?stackName=alonzo-testnet&templateURL=https://s3.amazonaws.com/sundaeswap-oss/alonzo-testnet/alonzo-testnet-ap-northeast-1.template)
Europe (Ireland)          | eu-west-1      | [![Launch Stack](https://cdn.sundaeswap.finance/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-1#/stacks/new?stackName=alonzo-testnet&templateURL=https://s3.amazonaws.com/sundaeswap-oss/alonzo-testnet/alonzo-testnet-eu-west-1.template)
Europe (London)           | eu-west-2      | [![Launch Stack](https://cdn.sundaeswap.finance/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-1#/stacks/new?stackName=alonzo-testnet&templateURL=https://s3.amazonaws.com/sundaeswap-oss/alonzo-testnet/alonzo-testnet-eu-west-2.template)
Europe (Paris)            | eu-west-3      | [![Launch Stack](https://cdn.sundaeswap.finance/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-1#/stacks/new?stackName=alonzo-testnet&templateURL=https://s3.amazonaws.com/sundaeswap-oss/alonzo-testnet/alonzo-testnet-eu-west-3.template)
Europe (Stockholm)        | eu-north-1	   | [![Launch Stack](https://cdn.sundaeswap.finance/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-1#/stacks/new?stackName=alonzo-testnet&templateURL=https://s3.amazonaws.com/sundaeswap-oss/alonzo-testnet/alonzo-testnet-eu-north-1.template)
South America (São Paulo) | sa-east-1      | [![Launch Stack](https://cdn.sundaeswap.finance/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-1#/stacks/new?stackName=alonzo-testnet&templateURL=https://s3.amazonaws.com/sundaeswap-oss/alonzo-testnet/alonzo-testnet-sa-east-1.template)


### Build the alonzo-testnet AMI

To create your testnet, begin by creating a custom AWS AMI for the testnet
instance.  First, create your variables file, `variables.pkrvars.hcl`.  In
it, you'll need

```hcl
region    = "your-region-here"
subnet_id = "your-subnet-here"
vpc_id    = "your-vpc-here"
```

```bash
packer build -var-file="variables.pkrvars.hcl" cardano-node.pkr.hcl
```

The AMI contains:

* aws cli
* docker
* a complete nix development environment
* cardano-node built from source using the 1.29.0

### Quickstart

Once the AMI has been built, launch an instance that has sufficient memory and CPU 
to run 3 nodes.  It is recommended to use `t3.xlarge` at a minimum.  The instance
is built using `Ubuntu 20.04` so the login will be `ubuntu`.

After you log in, you can start your testnet as follows:

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

#### alonzo - no fee env

```bash
bootstrap mkfiles -m 31415 -n maxLovelaceSupply=10000000000000000 -n minFeeA=0
```

#### alonzo with 10m epoch

```bash
bootstrap mkfiles -m 31415 \
  -n networkMagic=31415 \
  -n maxLovelaceSupply=10000000000000000 \
  -n maxTxSize=16384 \
  -n minUTxOValue=1000000 \
  -n minFeeB=155381 \
  -n minFeeA=44 \
  -n epochLength=600 \
  --alonzo-del executionPrices \
  --alonzo-set collateralPercentage=150 \
  --alonzo-set executionPrices.prMem.denominator=10000 \
  --alonzo-set executionPrices.prMem.numerator=577 \
  --alonzo-set executionPrices.prSteps.denominator=10000000 \
  --alonzo-set executionPrices.prSteps.numerator=721 \
  --alonzo-set lovelacePerUTxOWord=34482 \
  --alonzo-set maxBlockExUnits.exUnitsMem=50000000 \
  --alonzo-set maxBlockExUnits.exUnitsSteps=40000000000 \
  --alonzo-set maxCollateralInputs=3 \
  --alonzo-set maxTxExUnits.exUnitsMem=10000000 \
  --alonzo-set maxTxExUnits.exUnitsSteps=10000000000 \
  --alonzo-set maxValueSize=5000
```

### Additional Packages

The following additional packages have also been installed to facilitate
integration into the development environment

* tailscale client, https://tailscale.com
* AWS EFS client, https://aws.amazon.com/efs/
