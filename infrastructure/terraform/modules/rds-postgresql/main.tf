resource "aws_db_subnet_group" "this" {
  name       = "${var.name}-${var.environment}-subnet-group"
  subnet_ids = var.subnet_ids

  tags = {
    Name = "${var.name}-${var.environment}-db-subnet-group"
  }
}

resource "aws_db_instance" "this" {
  identifier              = "${var.name}-${var.environment}-postgres"
  engine                  = "postgres"
  engine_version          = var.engine_version
  instance_class          = var.instance_class
  allocated_storage       = var.allocated_storage
  storage_type            = var.storage_type
  db_name                 = var.db_name
  username                = var.username
  password                = var.password
  db_subnet_group_name    = aws_db_subnet_group.this.name
  vpc_security_group_ids  = var.vpc_security_group_ids
  backup_retention_period = var.backup_retention_period
  publicly_accessible     = var.publicly_accessible
  deletion_protection     = var.deletion_protection
  skip_final_snapshot     = true

  tags = {
    Name = "${var.name}-${var.environment}-postgres"
  }
}
