#!/usr/bin/env bash
# Bootstrap local Postgres for AI 协作看板（PG15+ public schema 权限）。
# 用法：./scripts/pg-bootstrap.sh
# 可选环境变量：PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE（管理端，默认 peer/sudo postgres）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP_USER="${APP_USER:-aicollab}"
APP_PASS="${APP_PASS:-aicollab}"
APP_DB="${APP_DB:-ai_collab}"
HOST="${APP_PGHOST:-127.0.0.1}"
PORT="${APP_PGPORT:-5432}"

run_as_super() {
  if [[ "${PGBOOTSTRAP_USE_SUDO:-1}" == "1" ]] && command -v sudo >/dev/null && id -u postgres >/dev/null 2>&1; then
    sudo -u postgres psql -v ON_ERROR_STOP=1 "$@"
  else
    psql -v ON_ERROR_STOP=1 "$@"
  fi
}

echo "==> ensure role ${APP_USER}"
run_as_super <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '${APP_USER}') THEN
    CREATE ROLE ${APP_USER} LOGIN PASSWORD '${APP_PASS}' CREATEDB;
  ELSE
    ALTER ROLE ${APP_USER} WITH LOGIN PASSWORD '${APP_PASS}' CREATEDB;
  END IF;
END
\$\$;
SQL

echo "==> ensure database ${APP_DB}"
EXISTS=$(run_as_super -Atc "SELECT 1 FROM pg_database WHERE datname='${APP_DB}'" || true)
if [[ "${EXISTS}" != "1" ]]; then
  run_as_super -c "CREATE DATABASE ${APP_DB} OWNER ${APP_USER}"
else
  run_as_super -c "ALTER DATABASE ${APP_DB} OWNER TO ${APP_USER}"
fi

echo "==> grant public schema (PG15+)"
run_as_super -d "${APP_DB}" <<SQL
GRANT ALL ON SCHEMA public TO ${APP_USER};
GRANT CREATE ON SCHEMA public TO ${APP_USER};
ALTER SCHEMA public OWNER TO ${APP_USER};
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ${APP_USER};
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ${APP_USER};
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO ${APP_USER};
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO ${APP_USER};
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
SQL

echo "==> probe as app user"
PGPASSWORD="${APP_PASS}" psql -h "${HOST}" -p "${PORT}" -U "${APP_USER}" -d "${APP_DB}" -v ON_ERROR_STOP=1 \
  -c "CREATE TABLE IF NOT EXISTS _perm_probe(id int); DROP TABLE _perm_probe; SELECT extversion FROM pg_extension WHERE extname='vector';"

echo "OK. Use:"
echo "  export DB_DRIVER=postgres"
echo "  export DATABASE_URL=postgres://${APP_USER}:${APP_PASS}@${HOST}:${PORT}/${APP_DB}?sslmode=disable"
echo "  export AUTO_MIGRATE=1"
echo "  # or: set -a && source ${ROOT}/.env.postgres && set +a"
