variable "name" {}
variable "private_subnet_ids" {}
variable "public_subnet_ids" {}
variable "vpc" {}
variable "allow_ssh" {}
variable "arch_ami" {}
variable "instance_type" {}
variable "extra_startup_script" {}
variable "instance_profile" {}
variable "desired_capacity" {}
variable "min_size" {}
variable "max_size" {}

output "lb" {
  value = aws_lb.proxy-lb
}

resource "aws_lb" "proxy-lb" {
  name               = "${var.name}-proxy"
  internal           = true
  load_balancer_type = "network"
  subnets            = var.private_subnet_ids
}

resource "aws_lb_target_group" "proxy-lb-tg" {
  name     = "${var.name}-proxy"
  port     = 9050
  protocol = "TCP"
  vpc_id   = var.vpc.id

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
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.proxy-lb-tg.arn
  }
}

resource "aws_security_group" "proxy_instance" {
  vpc_id      = var.vpc.id
  name_prefix = "proxy_instance"
  description = "access to/from proxy instances"

  dynamic "ingress" {
    for_each = var.allow_ssh ? ["ssh"] : []
    content {
      cidr_blocks = ["0.0.0.0/0", ]
      description = "ssh"
      protocol    = "tcp"
      from_port   = 0
      to_port     = 22
    }
  }

  ingress {
    cidr_blocks = [var.vpc.cidr_block]
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
  vpc_id      = var.vpc.id

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
  vpc_id      = var.vpc.id

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

data "aws_ami" "latest_amazon_linux" {
  owners      = ["amazon"]
  most_recent = true
  filter {
    name   = "name"
    values = ["amzn2-ami-kernel-*-hvm-*-${var.arch_ami}-gp2"]
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
  user_data = base64encode(join("\n", [<<EOF
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
for pid in \$(pgrep tor); do /bin/kill -1 \$pid ; done
EOSC
    chmod +x /tmp/hup
    (crontab -l 2>/dev/null || true; echo "* * * * * /tmp/hup >> /var/log/messages") | crontab -
EOF
    , var.extra_startup_script
  ]))
  iam_instance_profile {
    name = var.instance_profile.name
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
  vpc_zone_identifier       = var.public_subnet_ids
  health_check_grace_period = 180
  target_group_arns         = [aws_lb_target_group.proxy-lb-tg.arn]
  launch_template {
    id      = aws_launch_template.proxy-instance-template.id
    version = aws_launch_template.proxy-instance-template.latest_version
  }
}
