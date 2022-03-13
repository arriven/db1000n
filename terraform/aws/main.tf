terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

provider "aws" {
  region = var.region
}

data "aws_ami" "latest_amazon_linux" {
  owners      = ["amazon"]
  most_recent = true
  filter {
    name   = "name"
    values = ["amzn2-ami-kernel-*-hvm-*-x86_64-gp2"]
  }
}

# Create an IAM role for the Web Servers.
resource "aws_iam_role" "web_iam_role" {
  name               = "${var.name}_role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": ["ssm.amazonaws.com", "ec2.amazonaws.com"]
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "instance_profile" {
  name = "${var.name}_ip"
  role = aws_iam_role.web_iam_role.name
}

resource "aws_iam_role_policy_attachment" "instance_connect" {
  role       = aws_iam_role.web_iam_role.id
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2RoleforSSM"
}

resource "aws_iam_role_policy" "web_iam_role_policy" {
  name   = "${var.name}_policy"
  role   = aws_iam_role.web_iam_role.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Effect": "Allow",
        "Action": "ec2-instance-connect:SendSSHPublicKey",
        "Resource": "*",
        "Condition": {
            "StringEquals": {
                "ec2:osuser": "ec2-user"
            }
        }
    },
    {
        "Effect": "Allow",
        "Action": [
            "ec2:DescribeInstances"
        ],
        "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_security_group" "instance_connect" {
  vpc_id      = aws_vpc.main.id
  name_prefix = "instance_connect"
  description = "allow ssh"

  ingress {
    cidr_blocks      = ["0.0.0.0/0", ]
    description      = ""
    from_port        = 22
    ipv6_cidr_blocks = []
    prefix_list_ids  = []
    protocol         = "tcp"
    security_groups  = []
    self             = false
    to_port          = 22
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_internet_gateway" "test-env-gw" {
  vpc_id = aws_vpc.main.id
}

resource "aws_route_table" "route-table-test-env" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test-env-gw.id
  }
}

resource "aws_route_table_association" "subnet-association" {
  subnet_id      = aws_subnet.main.id
  route_table_id = aws_route_table.route-table-test-env.id
}

resource "aws_launch_template" "example" {
  name                                 = var.name
  image_id                             = data.aws_ami.latest_amazon_linux.id
  instance_initiated_shutdown_behavior = "terminate"
  instance_type                        = var.instance_type
  instance_market_options {
    market_type = "spot"
  }
  user_data = base64encode(<<EOF
#!/bin/bash -xe
exec > >(tee /var/log/user-data.log|logger -t user-data -s 2>/dev/console) 2>&1
    yum update -y
    amazon-linux-extras install docker
    service docker start
    usermod -a -G docker ec2-user
    chkconfig docker on
    docker run  -ti -d --restart always ghcr.io/arriven/db1000n-advanced

EOF
  )
  iam_instance_profile {
    name = aws_iam_instance_profile.instance_profile.name
  }
  vpc_security_group_ids = [aws_security_group.instance_connect.id]
  tag_specifications {
    resource_type = "instance"
    tags = {
      Name = "db1000n-server"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      Name = "db1000n-server"
    }
  }
  tag_specifications {
    resource_type = "network-interface"
    tags = {
      Name = "db1000n-server"
    }
  }
}

resource "aws_autoscaling_group" "example" {
  name                      = var.name
  capacity_rebalance        = true
  desired_capacity          = var.desired_capacity
  max_size                  = var.max_size
  min_size                  = var.min_size
  vpc_zone_identifier       = [aws_subnet.main.id]
  health_check_grace_period = 180
  launch_template {
    id      = aws_launch_template.example.id
    version = aws_launch_template.example.latest_version
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_subnet" "main" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true
}
