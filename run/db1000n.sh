#!/bin/sh

set -e

while [ ! -f /tmp/tunnel ]
do
	echo "waiting for tunnel"
	sleep 1
done

/usr/src/app/main

kill 1
