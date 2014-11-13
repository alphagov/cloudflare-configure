#!/bin/bash
set -eu

# Isolated GOPATH for Jenkins.
export GOPATH="$(pwd)/gopath"
export PATH="${GOPATH}/bin:${PATH}"

ORG="github.com/alphagov"
REPO="${ORG}/cloudflare-configure"
mkdir -p ${GOPATH}/src/${ORG}
[ -h ${GOPATH}/src/${REPO} ] || ln -s ../../../.. ${GOPATH}/src/${REPO}

make
