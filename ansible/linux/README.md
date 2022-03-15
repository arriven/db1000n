# Ansible playbook to setup plain VM

Script will be  useful if you have some plain VM and want to quickly setup it. Playbook tested on Ubuntu 20.04

## Requirements

- [Ansible](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html)
- Any VM with SSH access
- OPTIONAL: Subscription on [expressvpn.com](https://www.expressvpn.com) (get the activation code)

## Init

```sh
cp inventory.cfg.example inventory.cfg
```

Fill inventory.cfg with values.
Example with [ExpressVPN](https://www.expressvpn.com):

```sh
[target]
11.22.33.44 ansible_connection=ssh ansible_ssh_user=user

[target:vars]
ansible_python_interpreter=/usr/bin/python3
setup_vpn=True
expressvpn_activation_code=JFJKAMV7171
```

Example without VPN(not recommended):

```sh
[target]
11.22.33.44 ansible_connection=ssh ansible_ssh_user=user

[target:vars]
ansible_python_interpreter=/usr/bin/python3
setup_vpn=False
```

## Deploy

```sh
ansible-playbook -i inventory.cfg setup.yaml
```
