#!/usr/bin/env bash
set -euo pipefail
aws rds describe-db-instances --query 'DBInstances[*].DBInstanceIdentifier' --output text | xargs -n1 -I{} aws rds create-db-cluster-snapshot --db-cluster-identifier {} --db-cluster-snapshot-identifier {}-backup >/dev/null
