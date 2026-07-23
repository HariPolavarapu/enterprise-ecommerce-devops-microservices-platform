variable "vpc_id" {
  type    = string
  default = ""
}

variable "public_cidr" {
  type    = string
  default = "10.0.1.0/24"
}

variable "private_cidr" {
  type    = string
  default = "10.0.2.0/24"
}

variable "availability_zone" {
  type    = string
  default = "us-east-1a"
}
