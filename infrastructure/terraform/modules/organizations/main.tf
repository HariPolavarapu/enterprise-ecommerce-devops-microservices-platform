resource "aws_organizations_account" "this" {
  name  = var.name
  email = var.email
}
