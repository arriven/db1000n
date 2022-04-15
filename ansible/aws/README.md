# Description

## Prerequisites

1. Ansible 2.9
1. Collection amazon.aws (can be installed by command `ansible-galaxy collection install amazon.aws`)
1. Created key pair and public key stored in `AWS.pub` in same folder as the `aws-provisioning.yaml` playbook. The key pair can be created by utility `ssh-keygen` for Linux/MacOS or `puttygen` for Windows.

## How to provision EC2 instances

1. Create AWS unless created according to [AWS manual](https://docs.google.com/document/d/e/2PACX-1vTeCirL7ANTcX9vKXniKTjKkxGEE9Ftd1xBc0bHKPoSrd2aj5fNeresltDUEp6ZYNgM3EZF5csNj_R4/pub)
1. Set environment variables according to your operating system:
   - `AWS_ACCESS_KEY` - access key of the API user
   - `AWS_SECRET_KEY` - secret key of the API user
   - `AWS_REGION` - region where EC2 instances will be provisioned
1. Run `ansible-playbook aws-provisioning.yaml`. 

For POSIX systems it can look like that: `env AWS_ACCESS_KEY=_your-access-key-id_ AWS_SECRET_KEY=_your-secret-key_ AWS_REGION=us-east-1 ansible-playbook aws-provisioning.yaml`

Result should look like this:
```
ralfeus@trench:~/ddos$ env AWS_ACCESS_KEY=access-key-id AWS_SECRET_KEY=secret-key AWS_REGION=us-east-1 ansible-playbook aws-provisioning.yaml 
[WARNING]: provided hosts list is empty, only localhost is available. Note that the implicit localhost does not match 'all'

PLAY [Prepare AWS account and create instances] ****************************************************************************************************************

TASK [Add SSH key] *********************************************************************************************************************************************
changed: [localhost]

TASK [Create SG with allowed management traffic] ***************************************************************************************************************
changed: [localhost]

TASK [Create Linux instance] ***********************************************************************************************************************************
changed: [localhost]

TASK [Create Windows instance] *********************************************************************************************************************************
changed: [localhost]

PLAY RECAP *****************************************************************************************************************************************************
localhost                  : ok=4    changed=4    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0   


ralfeus@trench:~/ddos$ 
```

This playbook creates one Linux and one Windows EC2 instances from customized AMIs so whole setup fits into Free Tier limits. This means it can run for free 1 year from AWS account creation.

The Windows AMI has OpenSSH server installed. The public key is copied to the server if it's created using Ansible playbook provided. Otherwise it would have to be done manually after logging in to the server.

The Administrator password is reset during startup and can be retrieved in a standard AWS way.
