#!/usr/bin/bash

cmd="dlqt"
os="linux"
arch="amd64"
v="v0.3.2"
bin="$cmd-$os-$arch"
 
echo "downloading $bin $v"
wget https://github.com/emerconn/$cmd/releases/download/$v/$bin -O $bin -q --show-progress
 
echo "installing $cmd"
set -x
chmod +x $bin
sudo mv $bin /usr/local/bin/$cmd
 
$cmd -v