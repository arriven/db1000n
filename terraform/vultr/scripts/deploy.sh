#! /bin/bash

apt install -y vnstat

container_name="db1000n"
docker_cmd="docker run -d -it --rm --name=${container_name} -e ENABLE_PRIMITIVE=false --pull always ghcr.io/arriven/db1000n"

${docker_cmd}

echo "0 */2 * * * docker kill ${container_name} || true && ${docker_cmd}" >> cronjob
crontab cronjob
