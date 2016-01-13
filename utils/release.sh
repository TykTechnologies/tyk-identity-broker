#!/bin/bash
VERSION=$1
FOLDERNAME=tib-$VERSION
RELEASEDIR=release/$FOLDERNAME

mkdir -p $RELEASEDIR

gox -os="linux"

cp tib_sample.conf $RELEASEDIR/tib.conf
cp sample_profile.json $RELEASEDIR/profiles.json
cp LICENSE.md $RELEASEDIR/
cp README.md $RELEASEDIR/

ARCH=amd64
cp tyk-auth-proxy_linux_$ARCH $RELEASEDIR/tib
cd release/
tar -pczf ./tib-linux-$ARCH-$VERSION.tar.gz ./$FOLDERNAME
cd ..

ARCH=386
cp tyk-auth-proxy_linux_$ARCH $RELEASEDIR/tib
cd release/
tar -pczf ./tib-linux-$ARCH-$VERSION.tar.gz ./$FOLDERNAME
cd ..

ARCH=arm
cp tyk-auth-proxy_linux_$ARCH $RELEASEDIR/tib
cd release/
tar -pczf ./tib-linux-$ARCH-$VERSION.tar.gz ./$FOLDERNAME
cd ..

rm tyk-auth-proxy_linux_*
