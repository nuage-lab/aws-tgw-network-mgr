# AWS Nuage-TGW Network Manager

AWS TGW network manager is a tool that automates the connectivity of Nuage SD-WAN sites with AWS Network Manager and Transit Gateway. A Nuage SD-WAN site is connected via IPSEC tunnels to the TGW in the region.

## Pre-requisite

[install AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html) and [configure AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html) on the machine where you run the binary.

## install

awstgwnetwmgr can be installed using yum or apt package managers

```
sudo yum install https://github.com/nuage-lab/aws-tgw-network-mgr/releases/download/v0.1.0/awstgwnetworkmgr_0.1.0_linux_386.rpm
```

awstgwnetwmgr package can be installed using the installation script which detects the operating system type and installs the relevant package:

```
sudo curl -sL https://raw.githubusercontent.com/nuage-lab/aws-tgw-network-mgr/master/get.sh | sudo bash

sudo curl -sL https://raw.githubusercontent.com/nuage-lab/aws-tgw-network-mgr/master/get.sh | sudo bash -s -- -v 0.1.0
```

## configuration