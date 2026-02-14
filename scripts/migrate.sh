#!/usr/bin/env bash
set -euo pipefail

CONTAINER="${1:-deploy-mysql-1}"
SQL_PATH="${2:-../infra/mysql/init.sql}"

echo "Applying schema from: ${SQL_PATH} to container: ${CONTAINER}"
docker exec -i "${CONTAINER}" mysql -u root -proot < "${SQL_PATH}"
echo "done"
