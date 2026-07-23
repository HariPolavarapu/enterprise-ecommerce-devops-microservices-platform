# Disaster Recovery Plan

- Use multi-AZ RDS and ElastiCache replication for high availability.
- Store backups in S3 and automate restore with the provided scripts.
- Keep infrastructure definitions in Git and deploy through ArgoCD for rapid recovery.
- Validate failover quarterly and update runbooks after each exercise.
