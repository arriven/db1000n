# Define local variables
locals {
  cluster_version = "1.21"
  cluster_name    = "eks-${var.project}"
}

# Create VPC network
module "network" {
  source = "./modules/network"

  name = "vpc-${var.project}"
  tags = {
    "kubernetes.io/cluster/eks-${var.project}" = "shared"
  }
}

# Create EKS cluster
module "eks_cluster" {
  source = "./modules/eks-cluster"

  cluster_name           = local.cluster_name
  cluster_version        = local.cluster_version
  vpc_id                 = module.network.vpc_id
  subnets                = module.network.private_subnet_ids
  source_security_groups = module.eks_nodes.security_group_id
}

# Create EKS nodes
module "eks_nodes" {
  source = "./modules/eks-nodes"

  autoscale_group_name   = "eks-${var.project}-node"
  cluster_name           = local.cluster_name
  cluster_version        = local.cluster_version
  cluster_endpoint       = module.eks_cluster.endpoint
  cluster_ca_data        = module.eks_cluster.ca_data
  vpc_id                 = module.network.vpc_id
  subnets                = module.network.private_subnet_ids
  source_security_groups = module.eks_cluster.security_group_id
  key_name               = var.key_name
  public_key             = file(var.public_key)
  instance_type          = var.eks_node_instance_type
  desired_capacity       = var.eks_node_desired_capacity
  max_size               = var.eks_node_max_size
  min_size               = var.eks_node_min_size
}

# Setup kubernetes
module "kubernetes" {
  source = "./modules/kubernetes"

  profile                  = var.profile
  vpc_id                   = module.network.vpc_id
  cluster_name             = local.cluster_name
  cluster_endpoint         = module.eks_cluster.endpoint
  cluster_ca_data          = module.eks_cluster.ca_data
  worker_node_iam_role_arn = module.eks_nodes.worker_node_iam_role_arn
  autoscaler_iam_role_arn  = module.eks_cluster.autoscaler_iam_role_arn
}