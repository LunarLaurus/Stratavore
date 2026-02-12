#!/bin/bash
set -e

# Docker container name
PG_CONTAINER="stratavore-postgres"

# Stratavore PostgreSQL user (from environment in container)
PG_USER=$(docker exec "$PG_CONTAINER" env | grep POSTGRES_USER | cut -d= -f2)

if [ -z "$PG_USER" ]; then
    echo "Error: Could not detect POSTGRES_USER in container $PG_CONTAINER"
    exit 1
fi

# Prompt for new password
read -s -p "Enter new password for PostgreSQL user '$PG_USER': " NEW_PASS
echo ""
read -s -p "Confirm new password: " CONFIRM_PASS
echo ""

if [ "$NEW_PASS"!= "$CONFIRM_PASS" ]; then
    echo "Error: Passwords do not match"
    exit 1
fi

# Update password explicitly on the correct database
docker exec "$PG_CONTAINER" psql -U "$PG_USER" -d stratavore_state -c "ALTER USER $PG_USER WITH PASSWORD '$NEW_PASS';"

echo "[OK] Password updated successfully for user '$PG_USER'"