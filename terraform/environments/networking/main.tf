module "vpc" {
  source = "../../modules/vpc"
}

module "subnets" {
  source = "../../modules/subnets"
}
