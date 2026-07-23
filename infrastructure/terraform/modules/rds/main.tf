resource "aws_db_subnet_group" "this" {
  name       = "${var.name}-subnet-group"
  subnet_ids = var.subnet_ids
}

resource "aws_db_instance" "this" {
  identifier              = "${var.name}-postgres"
  engine                  = "postgres"
  engine_version          = "15.5"
  instance_class          = "db.t3.micro"
  allocated_storage       = 20
  db_name                 = "appdb"
  username                = "appuser"
  password                = "ChangeMe123!"
  db_subnet_group_name    = aws_db_subnet_group.this.name
  skip_final_snapshot     = true
  multi_az                = true
  backup_retention_period = 7
  publicly_accessible     = false
}
