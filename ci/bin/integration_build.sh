#!/bin/bash

set -ex

: ${ARCH:=amd64}
: ${PKG_PREFIX:=tyk-identity-broker}

DESCRIPTION="TBD"
BUILD_DIR="build"

function test {
    "$@"
    local status=$?
    if [ $status -ne 0 ]; then
        echo "error with $1" >&2
        exit 1
    fi
    return $status
}

# ---- APP BUILD START ---
echo "Building application"
test go build
# ---- APP BUILD END ---

mkdir $BUILD_DIR
# ---- CREATE TARGET FOLDER ---
echo "Copying Dashboard files"
cp -R app $BUILD_DIR/
cp -R public $BUILD_DIR/
cp README.md $BUILD_DIR/

echo "Creating $arch Tarball"
mv tyk-identity-broker $BUILD_DIR
tar -C $BUILD_DIR -pczf ${PKG_PREFIX}-$ARCH-$VERSION.tar.gz .
