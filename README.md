# AWS Nuage-TGW Network Manager

AWS TGW network manager is a service that allows to automates the connectivity of Nuage SD-WAN sites with AWS Network Manager and Transit Gateway. A Nuage SD-WAN site is connected via IPSEC tunnels to the TGW in the region. Through the AWS Network manager we can manage this connectivity in a global setting.

More information can be found [here](https://aws.amazon.com/transit-gateway/network-manager/)

## Pre-requisite

[install AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html) and [configure AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html) on the machine where you run the binary.

## install

awsnuagenmgr can be installed using yum or apt package managers

```
sudo yum install https://github.com/nuage-lab/aws-tgw-network-mgr/releases/download/v0.1.0/awsnuagenetwmgr_0.1.0_linux_386.rpm
```

awsnuagenmgr package can be installed using the installation script which detects the operating system type and installs the relevant package:

```
sudo curl -sL https://raw.githubusercontent.com/nuage-lab/aws-tgw-network-mgr/master/get.sh | sudo bash

sudo curl -sL https://raw.githubusercontent.com/nuage-lab/aws-tgw-network-mgr/master/get.sh | sudo bash -s -- -v 0.1.0
```

## Configuration

The deployment is guided through a configuration file with the following parameters:

- name: provides the name of the global network in aws
- aws parameters
    - profile you want to use for authentication through the aws API, if not specified we will use the default profile that is configured through **aws configure**
- nuage parameters
    - url: VSD url IP address and port
    - enterprise: the enterprise name in VSD that is used to connect to the AWS network manager/TGW
- topology parameters: describe the topology configuration for the connectivity between AWS and nuage
    - sites: describes the site configuration, like address information
        - name: the name of the site where the sdwan device is deployed
            - street
            - number
            - city
            - state
            - country
    - device-kinds: provide information that is global to the devices like, model and vendor
        - vendor
        - model
    - devices: each device is presented by a name, multiple devices can be added and should match the connection configuration
        - name: name of the device that is configured in Nuage VSD or name of TGW in AWS
            - kind: sdwan or tgw
            - serial: serial number of the device
            - region: this is mandatory for the tgw kind and inidcates where the tgw will be deployed
    - connections:
        - endpoints: represent the connectivity from sdwan to tgw in a list, the first element represents the sdwan endpoint, through <site-name>:<device-name>:<port-name>, the second element represnts the TGW through the name configured in the device section
            - labels are used to describe the connection attributed like:
                - provider: the name of the underlay provider
                - bwup: the amount of BW available upstream on the connection
                - bwdown: the amount of BW available downstream on the connection
                - kind: what kind of connection broadband or lte
                - public ip of the sd-wan uplink
                - asn: the AS number if GP would be used
                - cidr: that gets connected from the sd-wan appliance

An example is shown below:

```yaml
name: NuageTestNetwork

aws:
  profile: admin

nuage:
  enterprise: goPublic
  url:  "https://<ip address>:<port>"

topology:
  sites:
    home1:
      street: Copernicuslaan
      number: 50
      city: Antwerp
      state:
      country: Belgium 
  device-kinds:
    sdwan:
      vendor: nuage
      model: e300
  devices:
    goE300WifiLTE:
      kind: sdwan
      serial: 0123456789
#    nsg2:
#      kind: sdwan
#      serial: 0123456789
    tgw-euc1:
      kind: tgw
      region: eu-central-1
#    tgw-use1:
#      kind: tgw
#      region: us-east-1
        
  connections:
    - endpoints: ["home1:goE300WifiLTE:port1", "tgw-euc1"]
      labels: {"provider": "Telenet", "bwdown": "500", "bwup": "100", "kind": "broadband", "public-ip": "81.82.181.214", "asn": "65000", "cidr": "172.0.0.0/24"}
    - endpoints: ["home1:goE300WifiLTE:lte0", "tgw-euc1"]
      labels: {"provider": "Proximus", "bwdown": "200", "bwup": "50", "kind": "lte", "public-ip": "194.78.106.219", "asn": "65000", "cidr": "172.0.0.0/24"}
```

## deploy and destroy

We assume that the SD-WAN appliances are already configured in VSD, the tool is focussed on configuring the IPSEC connectivity between the SD-WAN appliances in nuage and the AWS network manager/TGW

The deployment is handled in 2 steps:

1. The deployment of the global network and the TGW(s) in AWS
2. The deployment/configuration of the SD-WAN sites in Nuage and the TGW/CGW and VPN connections in AWS