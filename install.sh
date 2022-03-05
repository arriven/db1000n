#!/usr/bin/env bash

set -euo pipefail

REPO="Arriven/db1000n"
INSTALL_OS="unknown"

case "$OSTYPE" in
  solaris*) INSTALL_OS="solaris" ;;
  darwin*)  INSTALL_OS="darwin" ;; 
  linux*)   INSTALL_OS="linux" ;;
  bsd*)     INSTALL_OS="bsd" ;;
  msys*)    INSTALL_OS="windows" ;;
  cygwin*)  INSTALL_OS="windows" ;;
  *)        echo "unknown: $OSTYPE"; exit 1 ;;
esac

if [ -z ${OSARCH+x} ];
then
  OSARCH=$(uname -m);
fi

INSTALL_ARCH="unknown"
case "$OSARCH" in
  x86_64*)  INSTALL_ARCH="amd64" ;;
  i386*)    INSTALL_ARCH="386" ;; 
  arm*)     INSTALL_ARCH="arm64" ;;
  *)        echo "unknown: $OSARCH"; exit 1 ;;
esac

INSTALL_VERSION="${INSTALL_OS}-${INSTALL_ARCH}"

echo "Downloading an archive..."
curl -s https://api.github.com/repos/${REPO}/releases/latest | grep "${INSTALL_VERSION}" | grep -Eo 'https://[^\"]*' | xargs -n 1 curl -s -L -O

INSTALL_TAG=$(curl --silent "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

ARCHIVE="db1000n-${INSTALL_TAG}-${INSTALL_VERSION}.tar.gz"

if [ "${INSTALL_OS}" = "darwin" ]
then
  MD5_BINARY="md5"
  MD5_SUFFIX="-q"
else
  MD5_BINARY="md5sum"
  MD5_SUFFIX="-t"
fi

echo "Checking md5 hash..."
if ! command -v "${MD5_BINARY}" &> /dev/null
then
  echo "Warning: md5sum/md5 not found. Could not check archive integrity. Please be careful when launching the executable."
else
  "${MD5_BINARY}" "${MD5_SUFFIX}" "${ARCHIVE}" | awk '{ print $1 }' > md5sum.txt
  if ! cmp --silent "${ARCHIVE}.md5" "md5sum.txt"; then
    echo "md5sum for ${ARCHIVE} failed. Please check the md5sum. File may possibly be corrupted."
    exit 1
  fi
fi

tar xvf "${ARCHIVE}"
echo "Successfully installed db1000n"
