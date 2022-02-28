# Death by 1000 needles

This repo is inspired by solution provided by [disbalancer](disbalancer.com) team. I became aware of it when they offered to use it against russian infrastructure during russian invasion to Ukraine

This is a simple distributed load generation client written in go. It is able to fetch simple json config from a local or remote location. The config describes which load generation jobs should be launched in parallel. I'm not aware of internal implementation of original disbalancer but I do know that it uses a lot more sophisticated techniques to balance the load and stuff. I do not intend to copy or replace it but rather provide a simple open source option. Feel free to use it in your load tests (wink-wink)

The software is provided as is under no guarantee.
I will update both the repo and this readme as I go during following days (date of writing this is 26th of February 2022, third day into russian invasion into Ukraine)

Synflood implementation is taken from https://github.com/bilalcaliskan/syn-flood and slightly patched. I couldn't just import the package as all the functionality code was in an internal package preventing import into other modules. Will figure it out better later (sorry to the owner).

## How to install

### binary install

go to releases page and install latest version for your os

### go install

run command in your terminal

```
go install github.com/Arriven/db1000n@latest
~/go/bin/db1000n
```

### docker install

how to install docker?

https://docs.docker.com/get-docker/

make sure you've set all available resources to docker

https://docs.docker.com/desktop/windows/#resources
https://docs.docker.com/desktop/mac/#resources

run d1000n

```
docker run ghcr.io/arriven/db1000n:latest
```

### shell install

run install script directly into the shell (useful for install through ssh)

```
curl https://raw.githubusercontent.com/Arriven/db1000n/main/install.sh | sh
```

the command above will detect the os and architecture, dowload the archive, validate it, and extract db1000n executable into the working directory. You can then run it via 

```
./db1000n
```
