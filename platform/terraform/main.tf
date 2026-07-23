terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.region
}

module "networking" {
  source = "./modules/networking"
  env    = var.env
  region = var.region
}

module "eks" {
  source              = "./modules/eks"
  env                 = var.env
  region              = var.region
  vpc_id              = module.networking.vpc_id
  private_subnet_ids  = module.networking.private_subnet_ids
  public_subnet_ids   = module.networking.public_subnet_ids
}

module "rds" {
  source     = "./modules/rds"
  env        = var.env
  region     = var.region
  subnet_ids = module.networking.private_subnet_ids
}

module "redis" {
  source     = "./modules/redis"
  env        = var.env
  region     = var.region
  subnet_ids = module.networking.private_subnet_ids
}

module "observability" {
  source = "./modules/observability"
  env    = var.env
  region = var.region
}
