#!/bin/bash

dir=`pwd`
echo $dir
export GOPATH=$dir

go test golang/packet -v
