#!/usr/bin/env bash

set -e

if [ "$BUILD_ENV_LOADED" != "true" ]
then
   echo "Please go to your project and load your build environment. 'source build-env.sh'"
   exit 1
fi

cd $BUILD_ENV_PROJECT

BUILD_ENV=forjj-golang-env

if [ "$http_proxy" != "" ]
then
   PROXY="--build-arg http_proxy=$http_proxy --build-arg https_proxy=$http_proxy --build-arg no_proxy=$no_proxy"
fi

USER="--build-arg UID=$(id -u) --build-arg GID=$(id -g)"

set -x
sudo docker build -t $BUILD_ENV $PROXY $USER glide

if [ "$GOPATH" = "" ]
then
    echo "Unable to build without GOPATH. Please set it (build)env or your local personal '.gopath')"
    exit 1
fi

# Requires forjj to be static.
export CGO_ENABLED=0
go install

if [ -f .publish_lacws ]
then
   scp  -P5001 $GOPATH/bin/forjj lacws.emea.hpqcorp.net:/storage/install/published/larsonsh/forjj
fi
