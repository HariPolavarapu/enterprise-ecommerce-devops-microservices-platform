#!/usr/bin/env bash
set -euo pipefail
DB_ID=${1:-appdb}
echo "Restoring $DB_ID"
