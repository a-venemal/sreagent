#!/bin/bash
set -e

# ============================================================
# SREAgent Docker Entrypoint
# - 等待 MySQL 就绪
# - 数据库不存在时创建（仅 CREATE IF NOT EXISTS，绝不 DROP）
# - 启动服务（应用启动时内嵌 golang-migrate 自动完成建表和升级）
# ============================================================

DB_HOST="${SREAGENT_DATABASE_HOST:-127.0.0.1}"
DB_PORT="${SREAGENT_DATABASE_PORT:-3306}"
DB_USER="${SREAGENT_DATABASE_USERNAME:-sreagent}"
DB_PASS="${SREAGENT_DATABASE_PASSWORD:-sreagent}"
DB_NAME="${SREAGENT_DATABASE_DATABASE:-sreagent}"

echo "============================================"
echo "  SREAgent - Intelligent SRE Platform"
echo "============================================"

# --- 等待 MySQL 端口就绪 ---
echo "[entrypoint] Waiting for MySQL at ${DB_HOST}:${DB_PORT} ..."
MAX_RETRIES=60
RETRY=0
until timeout 1 bash -c "echo >/dev/tcp/${DB_HOST}/${DB_PORT}" 2>/dev/null; do
  RETRY=$((RETRY + 1))
  if [ $RETRY -ge $MAX_RETRIES ]; then
    echo "[entrypoint] ERROR: MySQL not ready after ${MAX_RETRIES} retries, giving up."
    exit 1
  fi
  echo "[entrypoint]   waiting... ($RETRY/$MAX_RETRIES)"
  sleep 2
done
echo "[entrypoint] MySQL is ready."

# --- 确保数据库存在（仅 CREATE IF NOT EXISTS）---
# 优先使用应用账号（需有 CREATE 权限），失败时尝试 root（可选）
MYSQL_CMD="mysql -h${DB_HOST} -P${DB_PORT} -u${DB_USER} -p${DB_PASS}"
CREATE_SQL="CREATE DATABASE IF NOT EXISTS \`${DB_NAME}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

if ${MYSQL_CMD} -e "${CREATE_SQL}" 2>/dev/null; then
  echo "[entrypoint] Database '${DB_NAME}' is ready."
else
  # 如果应用账号没有 CREATE 权限，尝试 root
  ROOT_PASS="${MYSQL_ROOT_PASSWORD:-}"
  if [ -n "${ROOT_PASS}" ]; then
    mysql -h"${DB_HOST}" -P"${DB_PORT}" -uroot -p"${ROOT_PASS}" \
      -e "${CREATE_SQL}" 2>/dev/null \
      && echo "[entrypoint] Database '${DB_NAME}' created via root." \
      || echo "[entrypoint] WARNING: Could not create database, it may already exist."
  else
    echo "[entrypoint] WARNING: Could not create database (no root password provided). Assuming it already exists."
  fi
fi

# --- 启动 SREAgent（数据库迁移在应用内自动执行）---
echo "[entrypoint] Starting SREAgent server (:8080)..."
exec ./sreagent --config configs/config.yaml "$@"
