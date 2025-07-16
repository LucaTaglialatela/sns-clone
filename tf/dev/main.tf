locals {
  resource_prefix = "intern-luca"
  iam_role        = "arn:aws:iam::741641693274:role/intern-devops-ecs"
  baseurl         = "luca.intern.aws.prd.demodesu.com"
}

resource "aws_ecr_repository" "ecr_repository_web_app" {
  name = "${local.resource_prefix}-web-app"

  image_scanning_configuration {
    scan_on_push = true
  }

  # To allow terraform to delete the repository without having to delete stored images beforehand
  force_delete = true

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Create an ECS (Elastic Container Service) cluster which is a logical grouping of tasks or services
resource "aws_ecs_cluster" "main" {
  name = "${local.resource_prefix}-ecs-cluster"

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Create a VPC (Virtual Private Cloud) which is a logically isolated section of the AWS Cloud
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Create a subnet which is a range of IP addresses within a VPC
resource "aws_subnet" "main" {
  vpc_id = aws_vpc.main.id
  for_each = {
    "ap-northeast-1a" = "10.0.1.0/24"
    "ap-northeast-1c" = "10.0.2.0/24"
  }
  cidr_block = each.value
  availability_zone = each.key

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Create an internet gateway which enables communication between the VPC and the internet
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Create a route table to direct all network traffic from our subnets to our main internet gateway
resource "aws_route_table" "main" {
  vpc_id = aws_vpc.main.id

  # Direct all traffic to the internet gateway
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Create a route table association to connect subnets to the main route table
resource "aws_route_table_association" "main" {
  for_each       = aws_subnet.main
  route_table_id = aws_route_table.main.id
  subnet_id      = each.value.id
}

# Create a cloudwatch log group which is a resource within Amazon
# CloudWatch Logs that represents a logical collection of log streams
resource "aws_cloudwatch_log_group" "main" {
  name = "${local.resource_prefix}-cloudwatch"

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Create the ECS task definition to run our application container
resource "aws_ecs_task_definition" "main" {
  family                   = "main"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = local.iam_role
  task_role_arn            = local.iam_role
  network_mode             = "awsvpc"
  container_definitions = jsonencode([
    {
      name      = "${local.resource_prefix}-web-app"
      image     = "${aws_ecr_repository.ecr_repository_web_app.repository_url}:${var.image_tag}"
      cpu       = 256
      memory    = 512
      essential = true
      portMappings = [
        {
          containerPort = 8000
        }
      ]
      environmentFiles = [
        {
          "value" : "arn:aws:s3:::intern-luca/.env",
          "type" : "s3"
        }
      ],
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.main.name
          "awslogs-region"        = "ap-northeast-1"
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  tags = {
    Project = "${local.resource_prefix}"
  }
}

# Create an ECS service to execute tasks specified by our ECS task definitions
resource "aws_ecs_service" "main" {
  name            = "${local.resource_prefix}-ecs-service"
  propagate_tags  = "SERVICE"
  launch_type     = "FARGATE"
  desired_count   = 2
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.main.arn

  network_configuration {
    subnets          = [for subnet in aws_subnet.main : subnet.id]
    security_groups  = [aws_security_group.sg_ecs.id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.main.arn
    container_name   = "${local.resource_prefix}-web-app"
    container_port   = 8000
  }

  depends_on = [aws_vpc.main]

  # Zero downtime deployment
  deployment_minimum_healthy_percent = 100
  deployment_maximum_percent         = 200

  tags = {
    Project = "${local.resource_prefix}"
  }
}
