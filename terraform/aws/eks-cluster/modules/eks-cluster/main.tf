# Create EKS cluster role
resource "aws_iam_role" "eks_cluster" {
  name = "AWSEKSClusterRole"
  # name = "AmazonEKSClusterRole"
  assume_role_policy = data.aws_iam_policy_document.eks_cluster_assume_role_policy.json

  tags = {
    Cluster = var.cluster_name
  }

  lifecycle {
    ignore_changes = [name, name_prefix]
  }
}

# Create EKS cluster role policy attachment
resource "aws_iam_role_policy_attachment" "eks_cluster_AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.eks_cluster.name
}

resource "aws_iam_role_policy_attachment" "eks_cluster_AmazonEKSServicePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSServicePolicy"
  role       = aws_iam_role.eks_cluster.name
}

# Create EKS cluster autoscaler role
resource "aws_iam_role" "eks_cluster_autoscaler" {
  name               = "AmazonEKSClusterAutoscalerRole"
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy_web_identity.json

  inline_policy {
    name   = "AmazonEKSClusterAutoscalerPolicy"
    policy = data.aws_iam_policy_document.eks_cluster_autoscaler_policy.json
  }

  tags = {
    Cluster = var.cluster_name
  }

  lifecycle {
    ignore_changes = [name, name_prefix]
  }

  depends_on = [aws_iam_openid_connect_provider.oidc_provider]
}

# Create EKS cluster
resource "aws_eks_cluster" "eks_cluster" {
  name                      = var.cluster_name
  version                   = var.cluster_version
  role_arn                  = aws_iam_role.eks_cluster.arn
  enabled_cluster_log_types = ["api", "audit", "authenticator", "controllerManager", "scheduler"]

  vpc_config {
    security_group_ids = [aws_security_group.control_plane.id]
    subnet_ids         = var.subnets
  }

  depends_on = [
    aws_iam_role_policy_attachment.eks_cluster_AmazonEKSClusterPolicy,
    aws_iam_role_policy_attachment.eks_cluster_AmazonEKSServicePolicy
  ]
}

# create OIDC provider
resource "aws_iam_openid_connect_provider" "oidc_provider" {
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.cert.certificates[0].sha1_fingerprint]
  url             = aws_eks_cluster.eks_cluster.identity[0].oidc[0].issuer
}

# Create EKS cluster security group
resource "aws_security_group" "control_plane" {
  name        = "${var.cluster_name}-control-plane"
  description = "Cluster communication with worker nodes"
  vpc_id      = var.vpc_id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  revoke_rules_on_delete = true

  lifecycle {
    create_before_destroy = true
  }

  tags = {
    Name    = "${var.cluster_name}-control-plane"
    Cluster = var.cluster_name
  }
}

# Create EKS cluster security group rules
resource "aws_security_group_rule" "control_plane_ingress_nodes" {
  description              = "Allow cluster control plane to receive communication from the worker Kubelets"
  type                     = "ingress"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.control_plane.id
  source_security_group_id = var.source_security_groups
}