#!/bin/bash

set -e

export GO111MODULE=on

# print a command and execute it
show() {
 echo "$@" >&2
 eval "$@"
}

fatal() {
 echo "$@" >&2
 exit 1
}

for pkg in $(go list github.com/TykTechnologies/tyk-identity-broker/...);
do
    coveragefile=`echo "$pkg.cov" | awk -F/ '{print $NF}'`
    mgo_cov=`echo "$pkg-mongo-mgo.cov" | awk -F/ '{print $NF}'`
    mongo_cov=`echo "$pkg-mongo-official.cov" | awk -F/ '{print $NF}'`
    file_cov=`echo "$pkg-file.cov" | awk -F/ '{print $NF}'`
    show gocovmerge $mongo_cov $mgo_cov $file_cov > $coveragefile
    rm $mongo_cov $mgo_cov $file_cov
done