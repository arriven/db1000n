# Description

## Prerequisites

1. Ansible 2.9
1. Collection amazon.aws (can be installed by command `ansible-galaxy collection install amazon.aws`)
1. Created key pair and public key stored in `AWS.pub` in same folder as the `aws-provisioning.yaml` playbook

Here you can read a manual on AWS account creation: [AWS manual](https://docs.google.com/document/d/e/2PACX-1vTeCirL7ANTcX9vKXniKTjKkxGEE9Ftd1xBc0bHKPoSrd2aj5fNeresltDUEp6ZYNgM3EZF5csNj_R4/pub)

This playbook creates one Linux and one Windows EC2 instances from customized AMIs so whole setup fits into Free Tier limits. This means it can run for free 1 year from AWS account creation.

The AMI has OpenSSH server installed. The public key is copied to the server if it's created using Ansible playbook provided. Otherwise it would have to be done manually after logging in to the server.

The Administrator password is reset during startup and can be retrieved in a standard AWS way.
