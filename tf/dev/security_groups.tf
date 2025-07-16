# A security group acts as a firewall for AWS resources
# Egress = outbound data
# Ingress = inbound data

# Security group for ECS
resource "aws_security_group" "sg_ecs" {
  name   = "${local.resource_prefix}-ecs-security-group"
  vpc_id = aws_vpc.main.id

  # Allow all outbound data
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Security group for load balancer
resource "aws_security_group" "sg_lb" {
  name   = "${local.resource_prefix}-lb-security-group"
  vpc_id = aws_vpc.main.id

  # Allow all outbound data
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Security ingress rule for load balancer
# Allow all inbound data over HTTP (port 80)
resource "aws_vpc_security_group_ingress_rule" "allow_http_to_lb" {
  security_group_id = aws_security_group.sg_lb.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 80
  ip_protocol       = "tcp"
  to_port           = 80

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Security ingress rule for load balancer
# Allow all inbound data over HTTPS (port 443)
resource "aws_vpc_security_group_ingress_rule" "allow_https_to_lb" {
  security_group_id = aws_security_group.sg_lb.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 443
  ip_protocol       = "tcp"
  to_port           = 443

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Security ingress rule for ECS
# Allow all inbound data from the load balancer over port 8000
resource "aws_vpc_security_group_ingress_rule" "allow_app_traffic_from_lb" {
  security_group_id            = aws_security_group.sg_ecs.id
  referenced_security_group_id = aws_security_group.sg_lb.id
  from_port                    = 8000
  ip_protocol                  = "tcp"
  to_port                      = 8000

  tags = {
    Project = "${local.resource_prefix}"
  }
}
