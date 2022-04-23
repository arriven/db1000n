#!/usr/bin/env bash

set -euo pipefail

REPO=${REPO:-"Arriven/db1000n"}
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

if [ -z "${OSARCH+x}" ];
then
  OSARCH=$(uname -m);
fi

INSTALL_ARCH="unknown"
case "$OSARCH" in
  x86_64*)  INSTALL_ARCH="amd64" ;;
  i386*)    INSTALL_ARCH="386" ;; 
  armv6l)   INSTALL_ARCH="arm" ;;
  armv7l)   INSTALL_ARCH="arm" ;;
  arm*)     INSTALL_ARCH="arm64" ;;
  aarch64*) INSTALL_ARCH="arm64" ;;
  *)        echo "unknown: $OSARCH"; exit 1 ;;
esac

INSTALL_VERSION="${INSTALL_OS}_${INSTALL_ARCH}"

BROWSER_DOWNLOAD_URL=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep "${INSTALL_VERSION}" | grep -Eo 'https://[^\"]*')
CHECKSUM_DOWNLOAD_URL=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep "checksums" | grep -Eo 'https://[^\"]*')

ARCHIVE=${BROWSER_DOWNLOAD_URL##*/}
CHECKSUMS_FILE=${CHECKSUM_DOWNLOAD_URL##*/}

echo "Downloading an archive..."
echo "${BROWSER_DOWNLOAD_URL}" | xargs -n 1 curl -s -L -O
echo "Downloading checksums..."
echo "${CHECKSUM_DOWNLOAD_URL}" | xargs -n 1 curl -s -L -O

if [ "${INSTALL_OS}" = "darwin" ]
then
  SHA256_BINARY="shasum"
  SHA256_SUFFIX="-a 256"
else
  SHA256_BINARY="sha256sum"
  SHA256_SUFFIX=""
fi

echo "Checking sha256 hash..."
if ! command -v "${SHA256_BINARY}" &> /dev/null
then
  echo "Warning: sha256sum/shasum not found. Could not check archive integrity. Please be careful when launching the executable."
else
  # shellcheck disable=SC2086
  SHA256SUM=$(${SHA256_BINARY} ${SHA256_SUFFIX} ${ARCHIVE})
  if ! grep -q "${SHA256SUM}" "${CHECKSUMS_FILE}"; then
    echo "shasum for ${ARCHIVE} failed. Please check the shasum. File may possibly be corrupted."
    exit 1
  fi
fi

tar xvf "${ARCHIVE}"
echo "Successfully installed db1000n"
