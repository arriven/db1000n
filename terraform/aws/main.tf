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
    values = ["amzn2-ami-kernel-*-hvm-*-${var.arch_ami}-gp2"]
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
  for_each       = { for az, subnet in aws_subnet.main : az => subnet.id }
  subnet_id      = each.value
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
    PIPS=$(host -4 ${aws_lb.proxy-lb.dns_name} | egrep -o '[0-9]+(\.[0-9]+){3}$' | awk '{printf("socks5://%s:9050\n", $0)}' | paste -d',' -s -)
    docker run -ti -d --restart always ghcr.io/arriven/db1000n-advanced ./db1000n -proxy $PIPS

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
  depends_on = [aws_lb.proxy-lb]
}

resource "aws_autoscaling_group" "example" {
  name                      = var.name
  capacity_rebalance        = true
  desired_capacity          = var.desired_capacity
  max_size                  = var.max_size
  min_size                  = var.min_size
  vpc_zone_identifier       = [for subnet in aws_subnet.main : subnet.id]
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
  for_each                = { for azid, zone in slice(data.aws_availability_zones.azs.names, 0, var.zones) : zone => azid }
  availability_zone       = each.key
  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(aws_vpc.main.cidr_block, 8, var.zones + each.value)
  map_public_ip_on_launch = true
}

# >>> proxy

data "aws_availability_zones" "azs" {
  state = "available"
}

resource "aws_subnet" "private" {
  for_each                = { for azid, zone in slice(data.aws_availability_zones.azs.names, 0, var.zones) : zone => azid }
  availability_zone       = each.key
  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(aws_vpc.main.cidr_block, 8, 0 + each.value)
  map_public_ip_on_launch = false
}

resource "aws_lb" "proxy-lb" {
  name               = "${var.name}-proxy"
  internal           = true
  load_balancer_type = "network"
  #  security_groups    = [aws_security_group.lb.id]
  subnets = [for subnet in aws_subnet.private : subnet.id]
}

resource "aws_lb_target_group" "proxy-lb-tg" {
  name     = "${var.name}-proxy"
  port     = 9050
  protocol = "TCP" //"HTTP"
  vpc_id   = aws_vpc.main.id

  health_check {
    path                = "/"
    healthy_threshold   = 3
    unhealthy_threshold = 3
    interval            = 30
    protocol            = "HTTP"
    port                = 8080
  }
}

resource "aws_lb_listener" "proxy-lb-listener" {
  load_balancer_arn = aws_lb.proxy-lb.arn
  port              = "9050"
  protocol          = "TCP" //"HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.proxy-lb-tg.arn
  }
}

resource "aws_security_group" "proxy_instance" {
  vpc_id      = aws_vpc.main.id
  name_prefix = "proxy_instance"
  description = "access to/from proxy instances"

  ingress {
    cidr_blocks = ["0.0.0.0/0", ]
    description = "ssh"
    protocol    = "tcp"
    from_port   = 0
    to_port     = 22
  }

  ingress {
    cidr_blocks = [aws_vpc.main.cidr_block]
    description = "socks5"
    protocol    = "tcp"
    from_port   = 0
    to_port     = 9050
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "lb" {
  name        = "${var.name}-proxy-lb-security-group"
  description = "controls access to the proxy load balancer"
  vpc_id      = aws_vpc.main.id

  ingress {
    protocol        = -1
    from_port       = 0
    to_port         = 0
    security_groups = [aws_security_group.proxy_instance.id]
  }
}

resource "aws_security_group_rule" "lb-egress" {
  security_group_id        = aws_security_group.lb.id
  type                     = "egress"
  protocol                 = -1
  from_port                = 0
  to_port                  = 0
  source_security_group_id = aws_security_group.proxy-target-group.id
}

resource "aws_security_group" "proxy-target-group" {
  name        = "${var.name}-proxy-target-group"
  description = "controls access to the proxy containers"
  vpc_id      = aws_vpc.main.id

  ingress {
    protocol        = -1
    from_port       = 0
    to_port         = 0
    security_groups = [aws_security_group.lb.id]
  }

  egress {
    protocol        = -1
    from_port       = 0
    to_port         = 0
    security_groups = [aws_security_group.lb.id]
  }
}

resource "aws_launch_template" "proxy-instance-template" {
  name                                 = "${var.name}-proxy"
  image_id                             = data.aws_ami.latest_amazon_linux.id
  instance_initiated_shutdown_behavior = "terminate"
  instance_type                        = var.instance_type
  instance_market_options {
    market_type = "spot"
  }
  user_data = base64encode(<<EOF
#!/bin/bash
    yum update -y
    amazon-linux-extras install epel -y
    yum-config-manager --enable epel
    yum install tor nc -y
    echo "SOCKSPort 0.0.0.0:9050" >> /etc/tor/torrc
    service tor start
    chkconfig tor on
    while true; do echo -e 'HTTP/1.1 200 OK\r\n' | nc -lp 8080 > /dev/null; echo Healthcheck >> /var/log/messages; done &
    systemctl start crond
    systemctl enable crond
    cat << EOSC > /tmp/hup
#!/bin/bash
echo "Sending hup to tor processes"
for pid in \$(pgrep tor); do /bin/kill -1 $pid ; done
EOSC
    chmod +x /tmp/hup
    (crontab -l 2>/dev/null || true; echo "* * * * * /tmp/hup >> /var/log/messages") | crontab -
EOF
  )
  iam_instance_profile {
    name = aws_iam_instance_profile.instance_profile.name
  }
  vpc_security_group_ids = [aws_security_group.proxy_instance.id]
  tag_specifications {
    resource_type = "instance"
    tags = {
      Name = "db1000n-proxy"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      Name = "db1000n-proxy"
    }
  }
  tag_specifications {
    resource_type = "network-interface"
    tags = {
      Name = "db1000n-proxy"
    }
  }
}

resource "aws_autoscaling_group" "proxy" {
  name                      = "${var.name}-proxy"
  capacity_rebalance        = true
  desired_capacity          = var.desired_capacity
  max_size                  = var.max_size
  min_size                  = var.min_size
  vpc_zone_identifier       = [for subnet in aws_subnet.main : subnet.id]
  health_check_grace_period = 180
  target_group_arns         = [aws_lb_target_group.proxy-lb-tg.arn]
  launch_template {
    id      = aws_launch_template.proxy-instance-template.id
    version = aws_launch_template.proxy-instance-template.latest_version
  }
}
