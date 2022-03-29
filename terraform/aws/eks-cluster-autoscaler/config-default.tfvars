# common
profile                   = "default"
region                    = "eu-north-1"
project                   = "db1000n"

# ssh key
key_name                  = "db1000n"
public_key                = "~/.ssh/hbdm/id_rsa.pub"

# eks nodes
eks_node_instance_type    = "t3.xlarge"
eks_node_desired_capacity = "3"
eks_node_max_size         = "4"
eks_node_min_size         = "2"