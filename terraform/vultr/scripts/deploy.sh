#! /bin/bash

apt install -y vnstat

docker run -d -it --rm --name=db1000n --pull always ghcr.io/arriven/db1000n

sudo echo "0 */2 * * * docker kill db1000n || true && docker run -d -it --rm --name=db1000n --pull always ghcr.io/arriven/db1000n" >> cronjob
crontab cronjob
