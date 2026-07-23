variable "region" {
  type        = string
  description = "AWS region"
  default     = "us-east-1"
}

variable "environment" {
  type        = string
  description = "Deployment environment"
  default     = "production"
}

variable "profile" {
  type        = string
  description = "AWS CLI profile"
  default     = "default"
}
