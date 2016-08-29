#!/usr/bin/env bash
set -eo pipefail
set +H

if [ -s "$1" ]; then
    echo "No cloud specified"
    exit 1
fi

CLOUD=$1

echo -e "\n1. Checking that the binary is present and runs successfully..."
./carina --version

echo -e "\n2. Creating a kubernetes cluster named ci on the $CLOUD cloud..."
./carina --cloud=$CLOUD create --wait --template kubernetes-dev ci

echo -e "\n3. Downloading the cluster credentials..."
./carina --cloud=$CLOUD credentials ci

echo -e "\n4. Loading the cluster credentials..."
eval $(./carina --cloud=$CLOUD env ci)

echo -e "\n6. Testing the cluster connection with kubectl..."
kubectl cluster-info

echo -e "\n7. Removing the cluster..."
./carina --cloud=$CLOUD delete --wait ci

echo -e "\n#######\nAll done!\n#######\n"