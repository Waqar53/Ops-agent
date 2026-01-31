package infrastructure

import (
	"fmt"
	"strings"
)

// TerraformGenerator generates Terraform configurations
type TerraformGenerator struct{}

// NewTerraformGenerator creates a new Terraform generator
func NewTerraformGenerator() *TerraformGenerator {
	return &TerraformGenerator{}
}

// Generate generates Terraform configuration from infrastructure config
func (tg *TerraformGenerator) Generate(config *InfrastructureConfig) (string, error) {
	var tf strings.Builder

	// Provider configuration
	tf.WriteString(tg.generateProvider(config))
	tf.WriteString("\n\n")

	// VPC configuration
	if config.Network != nil && config.Network.VPC != nil {
		tf.WriteString(tg.generateVPC(config))
		tf.WriteString("\n\n")
	}

	// Compute configuration
	if config.Compute != nil {
		tf.WriteString(tg.generateCompute(config))
		tf.WriteString("\n\n")
	}

	// Database configuration
	if config.Database != nil {
		tf.WriteString(tg.generateDatabase(config))
		tf.WriteString("\n\n")
	}

	// Cache configuration
	if config.Cache != nil {
		tf.WriteString(tg.generateCache(config))
		tf.WriteString("\n\n")
	}

	// Storage configuration
	if config.Storage != nil {
		tf.WriteString(tg.generateStorage(config))
		tf.WriteString("\n\n")
	}

	// Load balancer configuration
	if config.Network != nil && config.Network.LoadBalancer != nil {
		tf.WriteString(tg.generateLoadBalancer(config))
		tf.WriteString("\n\n")
	}

	// Auto-scaling configuration
	if config.AutoScaling != nil && config.AutoScaling.Enabled {
		tf.WriteString(tg.generateAutoScaling(config))
		tf.WriteString("\n\n")
	}

	// Outputs
	tf.WriteString(tg.generateOutputs(config))

	return tf.String(), nil
}

func (tg *TerraformGenerator) generateProvider(config *InfrastructureConfig) string {
	switch config.Provider {
	case CloudAWS:
		return fmt.Sprintf(`terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "%s"
  
  default_tags {
    tags = {
      Project     = "%s"
      Environment = "%s"
      ManagedBy   = "OpsAgent"
    }
  }
}`, config.Region, config.Project, config.Environment)
	case CloudGCP:
		return fmt.Sprintf(`terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = "%s"
  region  = "%s"
}`, config.Project, config.Region)
	default:
		return ""
	}
}

func (tg *TerraformGenerator) generateVPC(config *InfrastructureConfig) string {
	vpc := config.Network.VPC

	return fmt.Sprintf(`# VPC Configuration
resource "aws_vpc" "main" {
  cidr_block           = "%s"
  enable_dns_hostnames = true
  enable_dns_support   = true
  
  tags = {
    Name = "%s-%s-vpc"
  }
}

# Public Subnets
%s

# Private Subnets
%s

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  
  tags = {
    Name = "%s-%s-igw"
  }
}

# NAT Gateways
%s`,
		vpc.CIDR,
		config.Project, config.Environment,
		tg.generateSubnets("public", vpc.PublicSubnets, config),
		tg.generateSubnets("private", vpc.PrivateSubnets, config),
		config.Project, config.Environment,
		tg.generateNATGateways(vpc.NATGateways, config))
}

func (tg *TerraformGenerator) generateSubnets(subnetType string, cidrs []string, config *InfrastructureConfig) string {
	var subnets strings.Builder

	for i, cidr := range cidrs {
		subnets.WriteString(fmt.Sprintf(`resource "aws_subnet" "%s_%d" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "%s"
  availability_zone = data.aws_availability_zones.available.names[%d]
  
  tags = {
    Name = "%s-%s-%s-subnet-%d"
  }
}

`, subnetType, i, cidr, i, config.Project, config.Environment, subnetType, i+1))
	}

	return subnets.String()
}

func (tg *TerraformGenerator) generateNATGateways(count int, config *InfrastructureConfig) string {
	if count == 0 {
		return ""
	}

	var nats strings.Builder

	for i := 0; i < count; i++ {
		nats.WriteString(fmt.Sprintf(`resource "aws_eip" "nat_%d" {
  domain = "vpc"
  
  tags = {
    Name = "%s-%s-nat-eip-%d"
  }
}

resource "aws_nat_gateway" "nat_%d" {
  allocation_id = aws_eip.nat_%d.id
  subnet_id     = aws_subnet.public_%d.id
  
  tags = {
    Name = "%s-%s-nat-%d"
  }
}

`, i, config.Project, config.Environment, i+1, i, i, i, config.Project, config.Environment, i+1))
	}

	return nats.String()
}

func (tg *TerraformGenerator) generateCompute(config *InfrastructureConfig) string {
	compute := config.Compute

	switch compute.Type {
	case "ecs":
		return tg.generateECS(config)
	case "eks":
		return tg.generateEKS(config)
	case "lambda":
		return tg.generateLambda(config)
	default:
		return tg.generateEC2(config)
	}
}

func (tg *TerraformGenerator) generateECS(config *InfrastructureConfig) string {
	compute := config.Compute

	return fmt.Sprintf(`# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "%s-%s-cluster"
  
  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# ECS Task Definition
resource "aws_ecs_task_definition" "app" {
  family                   = "%s-%s-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "%s"
  memory                   = "%s"
  
  container_definitions = jsonencode([{
    name  = "app"
    image = "${var.ecr_repository_url}:latest"
    
    portMappings = [{
      containerPort = 8080
      protocol      = "tcp"
    }]
    
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = "/ecs/%s-%s"
        "awslogs-region"        = "%s"
        "awslogs-stream-prefix" = "app"
      }
    }
  }])
}

# ECS Service
resource "aws_ecs_service" "app" {
  name            = "%s-%s-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = %d
  launch_type     = "FARGATE"
  
  network_configuration {
    subnets          = [aws_subnet.private_0.id, aws_subnet.private_1.id]
    security_groups  = [aws_security_group.ecs_tasks.id]
    assign_public_ip = false
  }
}`,
		config.Project, config.Environment,
		config.Project, config.Environment,
		compute.CPU, compute.Memory,
		config.Project, config.Environment, config.Region,
		config.Project, config.Environment,
		compute.MinInstances)
}

func (tg *TerraformGenerator) generateEKS(config *InfrastructureConfig) string {
	compute := config.Compute

	return fmt.Sprintf(`# EKS Cluster
resource "aws_eks_cluster" "main" {
  name     = "%s-%s-cluster"
  role_arn = aws_iam_role.eks_cluster.arn
  version  = "1.28"
  
  vpc_config {
    subnet_ids = [
      aws_subnet.private_0.id,
      aws_subnet.private_1.id
    ]
  }
}

# EKS Node Group
resource "aws_eks_node_group" "main" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "%s-%s-nodes"
  node_role_arn   = aws_iam_role.eks_nodes.arn
  subnet_ids      = [aws_subnet.private_0.id, aws_subnet.private_1.id]
  
  instance_types = ["%s"]
  
  scaling_config {
    desired_size = %d
    min_size     = %d
    max_size     = %d
  }
}`,
		config.Project, config.Environment,
		config.Project, config.Environment,
		compute.InstanceType,
		compute.MinInstances, compute.MinInstances, compute.MaxInstances)
}

func (tg *TerraformGenerator) generateEC2(config *InfrastructureConfig) string {
	compute := config.Compute

	return fmt.Sprintf(`# EC2 Launch Template
resource "aws_launch_template" "app" {
  name_prefix   = "%s-%s-"
  image_id      = data.aws_ami.amazon_linux_2.id
  instance_type = "%s"
  
  vpc_security_group_ids = [aws_security_group.app.id]
  
  user_data = base64encode(templatefile("user-data.sh", {
    environment = "%s"
  }))
}

# Auto Scaling Group
resource "aws_autoscaling_group" "app" {
  name                = "%s-%s-asg"
  min_size            = %d
  max_size            = %d
  desired_capacity    = %d
  vpc_zone_identifier = [aws_subnet.private_0.id, aws_subnet.private_1.id]
  
  launch_template {
    id      = aws_launch_template.app.id
    version = "$Latest"
  }
}`,
		config.Project, config.Environment,
		compute.InstanceType,
		config.Environment,
		config.Project, config.Environment,
		compute.MinInstances, compute.MaxInstances, compute.MinInstances)
}

func (tg *TerraformGenerator) generateLambda(config *InfrastructureConfig) string {
	return fmt.Sprintf(`# Lambda Function
resource "aws_lambda_function" "app" {
  function_name = "%s-%s-function"
  role          = aws_iam_role.lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs18.x"
  memory_size   = 512
  timeout       = 30
  
  filename         = "function.zip"
  source_code_hash = filebase64sha256("function.zip")
  
  environment {
    variables = {
      ENVIRONMENT = "%s"
    }
  }
}`,
		config.Project, config.Environment,
		config.Environment)
}

func (tg *TerraformGenerator) generateDatabase(config *InfrastructureConfig) string {
	db := config.Database

	return fmt.Sprintf(`# RDS Instance
resource "aws_db_instance" "main" {
  identifier     = "%s-%s-db"
  engine         = "%s"
  engine_version = "%s"
  instance_class = "%s"
  
  allocated_storage     = %d
  storage_encrypted     = %t
  multi_az              = %t
  backup_retention_period = %d
  
  db_name  = "%s"
  username = var.db_username
  password = var.db_password
  
  vpc_security_group_ids = [aws_security_group.rds.id]
  db_subnet_group_name   = aws_db_subnet_group.main.name
  
  skip_final_snapshot = false
  final_snapshot_identifier = "%s-%s-final-snapshot"
}

# DB Subnet Group
resource "aws_db_subnet_group" "main" {
  name       = "%s-%s-db-subnet"
  subnet_ids = [aws_subnet.private_0.id, aws_subnet.private_1.id]
}`,
		config.Project, config.Environment,
		db.Engine, db.Version, db.InstanceClass,
		db.Storage, db.Encryption, db.MultiAZ, db.BackupRetention,
		strings.ReplaceAll(config.Project, "-", "_"),
		config.Project, config.Environment,
		config.Project, config.Environment)
}

func (tg *TerraformGenerator) generateCache(config *InfrastructureConfig) string {
	cache := config.Cache

	return fmt.Sprintf(`# ElastiCache Cluster
resource "aws_elasticache_cluster" "main" {
  cluster_id           = "%s-%s-cache"
  engine               = "%s"
  node_type            = "%s"
  num_cache_nodes      = %d
  parameter_group_name = "default.redis7"
  port                 = 6379
  
  subnet_group_name    = aws_elasticache_subnet_group.main.name
  security_group_ids   = [aws_security_group.cache.id]
}

# Cache Subnet Group
resource "aws_elasticache_subnet_group" "main" {
  name       = "%s-%s-cache-subnet"
  subnet_ids = [aws_subnet.private_0.id, aws_subnet.private_1.id]
}`,
		config.Project, config.Environment,
		cache.Engine, cache.NodeType, cache.NumNodes,
		config.Project, config.Environment)
}

func (tg *TerraformGenerator) generateStorage(config *InfrastructureConfig) string {
	var buckets strings.Builder

	for _, bucket := range config.Storage.Buckets {
		bucketName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, bucket.Name)

		buckets.WriteString(fmt.Sprintf(`# S3 Bucket
resource "aws_s3_bucket" "%s" {
  bucket = "%s"
}

resource "aws_s3_bucket_versioning" "%s" {
  bucket = aws_s3_bucket.%s.id
  
  versioning_configuration {
    status = "%s"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "%s" {
  bucket = aws_s3_bucket.%s.id
  
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

`,
			bucket.Name, bucketName,
			bucket.Name, bucket.Name,
			map[bool]string{true: "Enabled", false: "Disabled"}[config.Storage.Versioning],
			bucket.Name, bucket.Name))
	}

	return buckets.String()
}

func (tg *TerraformGenerator) generateLoadBalancer(config *InfrastructureConfig) string {
	return fmt.Sprintf(`# Application Load Balancer
resource "aws_lb" "main" {
  name               = "%s-%s-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = [aws_subnet.public_0.id, aws_subnet.public_1.id]
}

# Target Group
resource "aws_lb_target_group" "app" {
  name     = "%s-%s-tg"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id
  
  health_check {
    path                = "/health"
    healthy_threshold   = 2
    unhealthy_threshold = 10
  }
}

# Listener
resource "aws_lb_listener" "app" {
  load_balancer_arn = aws_lb.main.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = var.ssl_certificate_arn
  
  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.app.arn
  }
}`,
		config.Project, config.Environment,
		config.Project, config.Environment)
}

func (tg *TerraformGenerator) generateAutoScaling(config *InfrastructureConfig) string {
	as := config.AutoScaling

	return fmt.Sprintf(`# Auto Scaling Target
resource "aws_appautoscaling_target" "app" {
  max_capacity       = %d
  min_capacity       = %d
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.app.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

# CPU-based Auto Scaling Policy
resource "aws_appautoscaling_policy" "cpu" {
  name               = "%s-%s-cpu-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.app.resource_id
  scalable_dimension = aws_appautoscaling_target.app.scalable_dimension
  service_namespace  = aws_appautoscaling_target.app.service_namespace
  
  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value = %.1f
  }
}

# Memory-based Auto Scaling Policy
resource "aws_appautoscaling_policy" "memory" {
  name               = "%s-%s-memory-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.app.resource_id
  scalable_dimension = aws_appautoscaling_target.app.scalable_dimension
  service_namespace  = aws_appautoscaling_target.app.service_namespace
  
  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
    target_value = %.1f
  }
}`,
		as.MaxCapacity, as.MinCapacity,
		config.Project, config.Environment,
		as.TargetCPU*100,
		config.Project, config.Environment,
		as.TargetMemory*100)
}

func (tg *TerraformGenerator) generateOutputs(config *InfrastructureConfig) string {
	var outputs strings.Builder

	outputs.WriteString("# Outputs\n")

	if config.Network != nil && config.Network.VPC != nil {
		outputs.WriteString(`output "vpc_id" {
  value = aws_vpc.main.id
}

`)
	}

	if config.Database != nil {
		outputs.WriteString(`output "database_endpoint" {
  value = aws_db_instance.main.endpoint
}

`)
	}

	if config.Cache != nil {
		outputs.WriteString(`output "cache_endpoint" {
  value = aws_elasticache_cluster.main.cache_nodes[0].address
}

`)
	}

	if config.Network != nil && config.Network.LoadBalancer != nil {
		outputs.WriteString(`output "load_balancer_dns" {
  value = aws_lb.main.dns_name
}

`)
	}

	return outputs.String()
}
