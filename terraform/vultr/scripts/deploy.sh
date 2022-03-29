#! /bin/bash

apt install -y vnstat

docker run -d -it --rm --name=d1000bn --pull always ghcr.io/arriven/db1000n-advanced

sudo echo "0 */2 * * * docker kill d1000bn || true && docker run -d -it --rm --name=d1000bn --pull always ghcr.io/arriven/db1000n-advanced" >> cronjob
crontab cronjob
