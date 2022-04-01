data "aws_iam_policy_document" "eks_worker_node_assume_role_policy" {
  statement {
    sid = "EKSWorkerAssumeRole"

    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions   = ["sts:AssumeRole"]
    resources = ["*"]
  }
}

data "aws_ami" "amazon_eks_nodes" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amazon-eks-node-${var.cluster_version}-*"]
  }

  filter {
    name = "owner-alias"
    values = [
      "amazon",
    ]
  }
}

data "template_file" "user_data" {
  template = file("${path.root}/scripts/node-user-data.sh")

  vars = {
    CLUSTER_NAME     = var.cluster_name
    CLUSTER_ENDPOINT = var.cluster_endpoint
    CLUSTER_CA_DATA  = var.cluster_ca_data
  }
}