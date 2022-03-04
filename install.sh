#!/usr/bin/env bash
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

INSTALL_ARCH="unknown"
OSARCH=$(uname -m)
case "$OSARCH" in
  x86_64*)  INSTALL_ARCH="amd64" ;;
  i386*)    INSTALL_ARCH="386" ;; 
  arm*)     INSTALL_ARCH="arm64" ;;
  *)        echo "unknown: $OSARCH"; exit 1 ;;
esac

INSTALL_VERSION="${INSTALL_OS}-${INSTALL_ARCH}"

curl https://api.github.com/repos/${REPO}/releases/latest | grep "${INSTALL_VERSION}" | grep -Eo 'https://[^\"]*' | xargs -n 1 curl -L -O

INSTALL_TAG=$(curl --silent "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

ARCHIVE="db1000n-${INSTALL_TAG}-${INSTALL_VERSION}.tar.gz"

if ! command -v md5sum &> /dev/null
then
    echo "Warning: md5sum is not be found. Could not check arhive integrity. Please be cautious when launching the executable."
else
    cat ${ARCHIVE} | md5sum | awk '{ print $1 }' > md5sum.txt
      if ! cmp --silent "${ARCHIVE}.md5" "md5sum.txt"; then
      echo "md5sum for ${ARCHIVE} failed. Please check the md5sum. File may possibly be corrupted."
      exit 1
    fi
fi

tar xvf ${ARCHIVE}