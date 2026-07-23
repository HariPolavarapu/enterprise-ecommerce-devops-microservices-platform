variable "name" {
  type    = string
  default = "enterprise"
}

variable "environment" {
  type    = string
  default = "prod"
}

variable "engine_version" {
  type    = string
  default = "15.4"
}

variable "instance_class" {
  type    = string
  default = "db.t3.micro"
}

variable "allocated_storage" {
  type    = number
  default = 20
}

variable "storage_type" {
  type    = string
  default = "gp3"
}

variable "db_name" {
  type    = string
  default = "enterprise"
}

variable "username" {
  type    = string
  default = "postgres"
}

variable "password" {
  type      = string
  sensitive = true
  default   = "ChangeMe123!"
}

variable "subnet_ids" {
  type    = list(string)
  default = []
}

variable "vpc_security_group_ids" {
  type    = list(string)
  default = []
}

variable "backup_retention_period" {
  type    = number
  default = 7
}

variable "publicly_accessible" {
  type    = bool
  default = false
}

variable "deletion_protection" {
  type    = bool
  default = true
}
