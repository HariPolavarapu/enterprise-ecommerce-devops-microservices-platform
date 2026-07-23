terraform {
  backend "s3" {
    bucket         = "enterprise-ecommerce-tfstate"
    key            = "platform/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "enterprise-ecommerce-tfstate-lock"
  }
}
