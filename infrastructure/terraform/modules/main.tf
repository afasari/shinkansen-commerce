# Terraform module for EKS cluster setup
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# VPC Module
module "vpc" {
  source = "./modules/vpc"

  name               = "${var.project_name}-vpc"
  cidr               = var.vpc_cidr
  availability_zones = var.availability_zones

  enable_nat_gateway   = true
  single_nat_gateway   = var.environment == "dev"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = var.common_tags
}

# EKS Cluster Module
module "eks" {
  source = "./modules/eks"

  cluster_name    = "${var.project_name}-cluster"
  cluster_version = var.kubernetes_version

  vpc_id          = module.vpc.vpc_id
  subnet_ids      = module.vpc.private_subnet_ids

  cluster_endpoint_public_access  = true
  cluster_endpoint_private_access = true

  cluster_enabled_log_types = var.cluster_log_types

  cluster_addons = {
    aws-ebs-csi-driver = {
      most_recent = true
    }
    coredns = {
      most_recent = true
    }
    kube-proxy = {
      most_recent = true
    }
    vpc-cni = {
      most_recent = false
      version = "v1.15.1-eksbuild.1"
    }
  }

  cluster_tags = var.common_tags
}

# IAM roles for service accounts
module "irsa" {
  source = "./modules/irsa"

  cluster_name = module.eks.cluster_name
  cluster_id   = module.eks.cluster_id

  oidc_provider_arn = module.eks.oidc_provider_arn
}

# RDS PostgreSQL for each service
module "rds_gateway" {
  source = "./modules/rds"

  identifier     = "${var.project_name}-gateway"
  instance_class = var.database_instance_class
  allocated_storage = 20
  engine         = "postgres"
  engine_version = "15.4"
  db_name        = "shinkansen_gateway"
  db_username    = var.database_admin_user

  vpc_id            = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids

  multi_az               = var.environment == "prod"
  backup_retention_period = var.environment == "prod" ? 30 : 7

  tags = var.common_tags
}

module "rds_product" {
  source = "./modules/rds"

  identifier     = "${var.project_name}-product"
  instance_class = var.database_instance_class
  allocated_storage = 100 # More storage for product service
  engine         = "postgres"
  engine_version = "15.4"
  db_name        = "shinkansen_product"
  db_username    = var.database_admin_user

  vpc_id            = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids

  multi_az               = var.environment == "prod"
  backup_retention_period = var.environment == "prod" ? 30 : 7

  tags = var.common_tags
}

module "rds_order" {
  source = "./modules/rds"

  identifier     = "${var.project_name}-order"
  instance_class = var.database_instance_class
  allocated_storage = 50
  engine         = "postgres"
  engine_version = "15.4"
  db_name        = "shinkansen_order"
  db_username    = var.database_admin_user

  vpc_id            = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids

  multi_az               = var.environment == "prod"
  backup_retention_period = var.environment == "prod" ? 30 : 7

  tags = var.common_tags
}

# ElastiCache Redis
module "elasticache_redis" {
  source = "./modules/elasticache"

  cluster_id      = "${var.project_name}-redis"
  node_type       = var.redis_node_type
  num_cache_nodes = var.environment == "prod" ? 3 : 1
  engine_version  = "7.1"

  vpc_id            = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids

  automatic_failover_enabled = var.environment == "prod"
  multi_az_enabled          = var.environment == "prod"

  tags = var.common_tags
}

# MSK (Kafka)
module "msk_kafka" {
  source = "./modules/msk"

  cluster_name = "${var.project_name}-kafka"
  kafka_version = "3.5.x"

  vpc_id            = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids

  broker_node_type    = var.kafka_broker_type
  number_of_broker_nodes = var.environment == "prod" ? 3 : 1

  tags = var.common_tags
}

# Application Load Balancers
module "alb_public" {
  source = "./modules/alb"

  name       = "${var.project_name}-public"
  internal   = false
  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.public_subnet_ids

  enable_deletion_protection = false
  enable_http2               = true

  tags = var.common_tags
}

module "alb_internal" {
  source = "./modules/alb"

  name       = "${var.project_name}-internal"
  internal   = true
  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnet_ids

  enable_deletion_protection = false
  enable_http2               = true

  tags = var.common_tags
}

# Route53 hosted zone
module "route53" {
  source = "./modules/route53"

  domain_name = var.domain_name
  vpc_id      = module.vpc.vpc_id

  tags = var.common_tags
}

# ACM certificates
module "acm" {
  source = "./modules/acm"

  domain_name = var.domain_name
  subject_alternative_names = [
    "*.${var.domain_name}",
    "api.${var.domain_name}",
    "*.api.${var.domain_name}",
  ]

  tags = var.common_tags
}

# S3 buckets
module "s3_assets" {
  source = "./modules/s3"

  bucket_name = "${var.project_name}-assets-${var.environment}"
  versioning  = true

  lifecycle_rules = [
    {
      id      = "delete-old-versions"
      enabled = true
      noncurrent_version_expiration_days = 90
    }
  ]

  tags = var.common_tags
}

module "s3_backups" {
  source = "./modules/s3"

  bucket_name = "${var.project_name}-backups-${var.environment}"
  versioning  = true

  lifecycle_rules = [
    {
      id      = "delete-old-backups"
      enabled = true
      expiration_days = 90
    }
  ]

  tags = var.common_tags
}

# CloudWatch Alarms and Dashboards
module "cloudwatch" {
  source = "./modules/cloudwatch"

  project_name = var.project_name
  environment  = var.environment

  alarm_actions = var.alarm_sns_topic_arn

  tags = var.common_tags
}

# Security groups
module "security_groups" {
  source = "./modules/security_groups"

  vpc_id      = module.vpc.vpc_id
  vpc_cidr    = var.vpc_cidr

  tags = var.common_tags
}

# Outputs
output "vpc_id" {
  value = module.vpc.vpc_id
}

output "eks_cluster_id" {
  value = module.eks.cluster_id
}

output "eks_cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

output "eks_cluster_security_group_id" {
  value = module.eks.cluster_security_group_id
}

output "alb_public_dns_name" {
  value = module.alb_public.dns_name
}

output "alb_internal_dns_name" {
  value = module.alb_internal.dns_name
}

output "rds_gateway_endpoint" {
  value = module.rds_gateway.endpoint
}

output "rds_product_endpoint" {
  value = module.rds_product.endpoint
}

output "rds_order_endpoint" {
  value = module.rds_order.endpoint
}

output "elasticache_redis_endpoint" {
  value = module.elasticache_redis.endpoint
}

output "msk_kafka_bootstrap_brokers" {
  value = module.msk_kafka.bootstrap_brokers
}
