# Create VPC
resource "aws_vpc" "vpc" {
  cidr_block           = var.vpc_cidr_block
  enable_dns_support   = true
  enable_dns_hostnames = true


  tags = merge(
    {
      Name = "${var.name}"
    },
    var.tags
  )
}

# Create Internet gateway (associated with public subnets)
resource "aws_internet_gateway" "internetgw" {
  vpc_id = aws_vpc.vpc.id


  tags = {
    Name = "${var.name}-internetgw"
  }
}

# Create elastic IPs (associated with NAT gateways)
resource "aws_eip" "natgw" {
  count = var.amount_az
  vpc   = true

  depends_on = [aws_internet_gateway.internetgw]
}

# Create NAT gateway (for each AZ)
resource "aws_nat_gateway" "natgw" {
  count         = var.amount_az
  allocation_id = aws_eip.natgw[count.index].id
  subnet_id     = aws_subnet.public[count.index].id


  tags = {
    Name = "${var.name}-natgw-${data.aws_availability_zones.available.names[count.index]}"
  }

  depends_on = [aws_internet_gateway.internetgw]
}

# Create public subnets
resource "aws_subnet" "public" {
  count                   = var.amount_az
  vpc_id                  = aws_vpc.vpc.id
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  cidr_block              = cidrsubnet(var.vpc_cidr_block, 8, count.index + 1)
  map_public_ip_on_launch = true


  tags = merge(
    {
      "kubernetes.io/role/elb" = 1
    },
    var.tags
  )
}

# Create private subnets
resource "aws_subnet" "private" {
  count             = var.amount_az
  vpc_id            = aws_vpc.vpc.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(var.vpc_cidr_block, 8, count.index + 3)


  tags = merge(
    {
      "kubernetes.io/role/internal-elb" = 1
    },
    var.tags
  )
}

# Create Internet gateway route table
resource "aws_route_table" "internetgw" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.internetgw.id
  }


  tags = {
    Name = "${var.name}-internetgw"
  }
}

# Create NAT gateway route table (one for each a-z)
resource "aws_route_table" "natgw" {
  count  = var.amount_az
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.natgw[count.index].id
  }


  tags = {
    Name = "${var.name}-natgw-${data.aws_availability_zones.available.names[count.index]}"
  }
}

# Create Internet gateway route table association (associated with public subnets)
resource "aws_route_table_association" "internetgw" {
  count          = var.amount_az
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.internetgw.id
}

# Create NAT gateway route table association (associated with private subnets)
resource "aws_route_table_association" "natgw" {
  count          = var.amount_az
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.natgw[count.index].id
}