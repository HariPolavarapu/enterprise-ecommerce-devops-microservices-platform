module "eks" {
  source = "../../modules/eks"
}

module "rds-postgresql" {
  source = "../../modules/rds-postgresql"
}
