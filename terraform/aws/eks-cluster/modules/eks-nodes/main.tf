# Create EKS node IAM role
resource "aws_iam_role" "eks_worker_node" {
  name = "AWSEKSWorkerNodeRole"
  # name = "AmazonEKSWorkerNodeRole"
  assume_role_policy = data.aws_iam_policy_document.eks_worker_node_assume_role_policy.json

  tags = {
    Cluster = var.cluster_name
  }

  lifecycle {
    ignore_changes = [name, name_prefix]
  }
}

# Create EKS node role IAM policies
resource "aws_iam_role_policy" "assume_role_policy" {
  name = "AWSEKSWorkerNodeAssumeRolePolicy"
  # name   = "AmazonEKSWorkerNodeAssumeRolePolicy"
  role   = aws_iam_role.eks_worker_node.name
  policy = data.aws_iam_policy_document.assume_role_policy.json
}

# Create EKS node role IAM policy attachments
resource "aws_iam_role_policy_attachment" "eks_worker_node_AmazonEKSWorkerNodePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.eks_worker_node.name
}

resource "aws_iam_role_policy_attachment" "eks_worker_node_AmazonEKS_CNI_Policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.eks_worker_node.name
}

resource "aws_iam_role_policy_attachment" "eks_worker_node_AmazonEC2ContainerRegistryReadOnly" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.eks_worker_node.name
}

# Create EKS node instance profile
resource "aws_iam_instance_profile" "eks_worker_node" {
  name = "AWSEKSWorkerNodeInstanceProfile"
  # name = "AmazonEKSWorkerNodeInstanceProfile"
  role = aws_iam_role.eks_worker_node.name

  lifecycle {
    ignore_changes = [name, name_prefix]
  }
}

# Create SSH key-pair
resource "aws_key_pair" "instance" {
  key_name   = var.key_name
  public_key = var.public_key
}

# Create EKS node launch template
resource "aws_launch_template" "eks_nodes" {
  name          = var.autoscale_group_name
  image_id      = var.ami_id != "" ? var.ami_id : data.aws_ami.amazon_eks_nodes.id
  instance_type = var.instance_type
  key_name      = var.key_name
  user_data     = base64encode(data.template_file.user_data.rendered)

  iam_instance_profile {
    arn = aws_iam_instance_profile.eks_worker_node.arn
  }

  network_interfaces {
    delete_on_termination = true
    security_groups       = [aws_security_group.eks_nodes.id]
  }

  block_device_mappings {
    device_name = var.device_name

    ebs {
      volume_size           = var.volume_size
      volume_type           = var.volume_type
      encrypted             = var.encrypted
      delete_on_termination = true
    }
  }

  # Should be enabled for production
  monitoring {
    enabled = false
  }

  lifecycle {
    create_before_destroy = true
  }

  tags = {
    Name    = var.autoscale_group_name
    Cluster = var.cluster_name
  }
}

# Create EKS node autoscaling group
resource "aws_autoscaling_group" "eks_nodes" {
  name                = var.autoscale_group_name
  desired_capacity    = var.desired_capacity
  max_size            = var.max_size
  min_size            = var.min_size
  vpc_zone_identifier = var.subnets

  force_delete         = true
  termination_policies = var.termination_policies

  dynamic "launch_template" {
    for_each = var.use_spot_instances ? [] : [1]

    content {
      id      = aws_launch_template.eks_nodes.id
      version = "$Latest"
    }
  }

  dynamic "mixed_instances_policy" {
    for_each = var.use_spot_instances ? [1] : []

    content {
      instances_distribution {
        on_demand_base_capacity                  = 0
        on_demand_percentage_above_base_capacity = 0
      }

      launch_template {
        launch_template_specification {
          launch_template_id = aws_launch_template.eks_nodes.id
          version            = "$Latest"
        }

        dynamic "override" {
          for_each = var.spot_overrides

          content {
            instance_type     = override.value["instance_type"]
            weighted_capacity = override.value["weighted_capacity"]
          }
        }
      }
    }
  }

  timeouts {
    delete = "15m"
  }

  lifecycle {
    create_before_destroy = true
  }

  tags = [
    {
      key                 = "Name"
      value               = var.autoscale_group_name
      propagate_at_launch = true
    },
    {
      key                 = "Cluster"
      value               = var.cluster_name
      propagate_at_launch = true
    },
    {
      key                 = "kubernetes.io/cluster/${var.cluster_name}"
      value               = "owned"
      propagate_at_launch = true
    },
    {
      key                 = "k8s.io/cluster-autoscaler/enabled"
      value               = "true"
      propagate_at_launch = true
    },
    {
      key                 = "k8s.io/cluster-autoscaler/${var.cluster_name}"
      value               = "owned"
      propagate_at_launch = true
    }
  ]
}

# Create EKS node security group
resource "aws_security_group" "eks_nodes" {
  name        = "${var.cluster_name}-node"
  description = "Security group for all nodes in the cluster"
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
    Name                                        = "${var.cluster_name}-node"
    Cluster                                     = var.cluster_name
    "kubernetes.io/cluster/${var.cluster_name}" = "owned"
  }
}

# Create EKS node security group rules
resource "aws_security_group_rule" "ingress_self" {
  description              = "Allow node to communicate with each other"
  type                     = "ingress"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.eks_nodes.id
  source_security_group_id = aws_security_group.eks_nodes.id
}

resource "aws_security_group_rule" "ingress_cluster" {
  description              = "Allow worker Kubelets and pods to receive communication from the cluster control plane"
  type                     = "ingress"
  from_port                = 1025
  to_port                  = 65535
  protocol                 = "tcp"
  security_group_id        = aws_security_group.eks_nodes.id
  source_security_group_id = var.source_security_groups
}

resource "aws_security_group_rule" "ingress_control_plane" {
  description              = "Allow pods to communicate with the cluster API Server"
  type                     = "ingress"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.eks_nodes.id
  source_security_group_id = var.source_security_groups
}