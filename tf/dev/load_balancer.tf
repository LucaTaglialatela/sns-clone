# Create a load balancer
resource "aws_lb" "main" {
  name                       = "${local.resource_prefix}-lb-tf"
  load_balancer_type         = "application"
  subnets                    = [for subnet in aws_subnet.main : subnet.id]
  security_groups            = [aws_security_group.sg_lb.id]
  enable_deletion_protection = false

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Create a load balancer target group targeting port 8000
resource "aws_lb_target_group" "main" {
  name        = "${local.resource_prefix}-lb-tg"
  port        = 8000
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    port                = 8000
    healthy_threshold   = 6
    unhealthy_threshold = 2
    timeout             = 2
    interval            = 5
    matcher             = "200"
  }

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Retrieve the existing Route53 hosted zone
data "aws_route53_zone" "parent" {
  name = "intern.aws.prd.demodesu.com."
}

# Create an HTTPS listener for the Load Balancer using the validated ACM certificate
resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.main.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-Res-2021-06"
  certificate_arn   = aws_acm_certificate_validation.main.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main.arn
  }

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Create an HTTP listener that redirects all traffic to HTTPS
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Create an A-record to point the domain name to the load balancer
resource "aws_route53_record" "main" {
  zone_id = data.aws_route53_zone.parent.zone_id
  name    = local.baseurl
  type    = "A"

  alias {
    zone_id                = aws_lb.main.zone_id
    name                   = aws_lb.main.dns_name
    evaluate_target_health = true
  }
}
