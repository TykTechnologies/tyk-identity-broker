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

TEST_TIMEOUT=10m
db="file"

if [[ -n $1 ]]; then
    db=$1
fi

if [[ $db = "mongo-mgo" ]]; then
    export TYK_IB_STORAGE_STORAGETYPE="mongo"
    export TYK_IB_STORAGE_MONGOCONF_MONGOURL="mongodb://localhost/tyk_identity_broker"
    export TYK_IB_STORAGE_MONGOCONF_DRIVER="mgo"
    : # do nothing
elif [[ $db = "mongo-official" ]]; then
    export TYK_IB_STORAGE_STORAGETYPE="mongo"
    export TYK_IB_STORAGE_MONGOCONF_MONGOURL="mongodb://localhost/tyk_identity_broker"
    export TYK_IB_STORAGE_MONGOCONF_DRIVER="mongo-go"
elif [[ $db = "file" ]]; then
    export TYK_IB_STORAGE_STORAGETYPE="file"
    export TYK_IB_PROFILEDIR="./"
else
    fatal "unsupported database: $db"
fi

echo "Running with $TYK_IB_STORAGE_STORAGETYPE database using $TYK_IB_STORAGE_MONGOCONF_DRIVER driver"

show go vet . || fatal "go vet errored"

GO_FILES=$(find * -name '*.go' )

echo "Formatting checks..."

FMT_FILES="$(gofmt -s -l ${GO_FILES})"
if [[ -n ${FMT_FILES} ]]; then
	fatal "Run 'gofmt -s -w' on these files:\n$FMT_FILES"
fi

echo "gofmt check is ok!"

IMP_FILES="$(goimports -l ${GO_FILES})"
if [[ -n ${IMP_FILES} ]]; then
	fatal "Run 'goimports -w' on these files:\n$IMP_FILES"
fi

echo "goimports check is ok!"

for pkg in $(go list github.com/TykTechnologies/tyk-identity-broker/...);
do
    race="-race"
    echo "Testing... $pkg"
    coveragefile=`echo "$pkg-$db" | awk -F/ '{print $NF}'`
    show go test -timeout ${TEST_TIMEOUT} ${race} --coverprofile=${coveragefile}.cov -v ${pkg}
done
