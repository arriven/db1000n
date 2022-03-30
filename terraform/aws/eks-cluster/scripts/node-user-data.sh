#!/bin/bash -xe

sudo /etc/eks/bootstrap.sh --apiserver-endpoint ${CLUSTER_ENDPOINT} --b64-cluster-ca ${CLUSTER_CA_DATA} ${CLUSTER_NAME}

