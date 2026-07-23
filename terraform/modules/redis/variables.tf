variable "name" {
  type    = string
  default = "enterprise"
}

variable "environment" {
  type    = string
  default = "prod"
}

variable "region" {
  type    = string
  default = "us-east-1"
}

variable "subnet_ids" {
  type = list(string)
}
