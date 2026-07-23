module "eks" {
  source = "../../modules/eks"
}

module "ecr" {
  source = "../../modules/ecr"
}
