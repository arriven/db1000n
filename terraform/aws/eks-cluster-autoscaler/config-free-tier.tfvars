# common
profile                   = "free-tier"
region                    = "us-east-1"
project                   = "FreeTier"

# ssh key
key_name                  = "db1000n"
public_key                = "~/.ssh/hbdm/id_rsa.pub"

# eks nodes
eks_node_instance_type    = "t3.large"
eks_node_desired_capacity = "2"
eks_node_max_size         = "3"
eks_node_min_size         = "2"