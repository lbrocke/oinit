#!/bin/sh

BASEDIR=/go/src/github.com/lbrocke/oinit
# check version of goreleaser
docker images | grep goreleaser
# update goreleaser
docker pull goreleaser/goreleaser
# run goreleaser to build packages
docker run --rm --privileged \
  -v "$PWD":"$BASEDIR" \
  -w "$BASEDIR" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  goreleaser/goreleaser release --skip-publish --skip-docker
# do not add commands here, script exists with status of last command