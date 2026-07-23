resource "aws_elasticache_subnet_group" "main" {
  name       = "${var.env}-redis-subnet"
  subnet_ids = var.subnet_ids
}

resource "aws_elasticache_replication_group" "main" {
  replication_group_id          = "${var.env}-redis"
  description                   = "Redis for app"
  node_type                     = "cache.t3.micro"
  num_node_groups               = 1
  replicas_per_node_group      = 1
  automatic_failover_enabled   = true
  subnet_group_name            = aws_elasticache_subnet_group.main.name
  security_group_ids           = []
}
