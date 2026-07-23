resource "aws_db_subnet_group" "main" {
  name       = "${var.env}-db-subnet"
  subnet_ids = var.subnet_ids
}

resource "aws_db_instance" "main" {
  identifier              = "${var.env}-postgres"
  engine                  = "postgres"
  engine_version          = "15.5"
  instance_class          = "db.t3.micro"
  allocated_storage       = 20
  db_name                 = "appdb"
  username                = "appuser"
  password                = "ChangeMe123!"
  skip_final_snapshot     = true
  multi_az                = true
  db_subnet_group_name    = aws_db_subnet_group.main.name
  publicly_accessible     = false
  backup_retention_period = 7
}
