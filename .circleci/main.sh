#!/bin/bash -xe

PROJ_PATH=`pwd`

# First let's install Go 1.11
echo "Installing Go 1.11"
cd /tmp

wget https://dl.google.com/go/go1.11.linux-amd64.tar.gz
sudo tar -xvf go1.11.linux-amd64.tar.gz
# Uninstall older version of Go
sudo rm -rf /usr/local/go
sudo mv go /usr/local

export GOROOT=/usr/local/go
export GOPATH=$PROJ_PATH
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

go version

cd $PROJ_PATH

# This allows our mv calls to move also hidden files and folders
# for example /.circleci ;-)
shopt -s dotglob
mkdir -p /tmp/project
mv ../project/* /tmp/project/
mkdir -p src/github.com/orbs-network/lean-helix-go
mv /tmp/project/* ./src/github.com/orbs-network/lean-helix-go/

cd src/github.com/orbs-network/lean-helix-go/

./test.sh
