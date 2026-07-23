# Operational Runbook

## Daily Operations
1. Verify Jenkins pipelines and ArgoCD sync status.
2. Check Prometheus and Grafana dashboards.
3. Review Fluent Bit and OpenSearch logs.
4. Validate RDS backups and Redis failover status.
5. Confirm Lambda cleanup jobs and backup scripts.

## Incident Response
1. Fail over to the standby region if required.
2. Check ALB health and EKS node status.
3. Restore from RDS backup using the restore script.
4. Notify stakeholders and update incident records.
