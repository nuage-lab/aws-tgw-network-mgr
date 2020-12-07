
## setup environment

export AWS_ACCESS_KEY_ID="AKIAY4GHMD5ASE2QHTET"
export AWS_SECRET_ACCESS_KEY="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"
export AWS_DEFAULT_REGION="eu-central-1"
export AWS_VPC_NAME="eks-nokia-paco-vpc"

## deploy configuration

go run *.go deploy -c conf/nuage-aws-tgw.yaml