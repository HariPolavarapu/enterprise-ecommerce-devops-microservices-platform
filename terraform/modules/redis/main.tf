resource "aws_elasticache_subnet_group" "this" {
  name       = "${var.name}-redis-subnet"
  subnet_ids = var.subnet_ids
}

resource "aws_elasticache_replication_group" "this" {
  replication_group_id       = "${var.name}-redis"
  description                = "Redis cluster for application"
  node_type                  = "cache.t3.micro"
  num_node_groups            = 1
  replicas_per_node_group    = 1
  automatic_failover_enabled = true
  subnet_group_name          = aws_elasticache_subnet_group.this.name
}
