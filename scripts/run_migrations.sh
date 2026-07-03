#!/bin/bash
# ============================================================
# Run all database migrations directly via psql
# ============================================================
# Karena banyak migration punya PL/pgSQL function dengan $$
# yang tidak compatible dengan goose parser, kita jalanin
# langsung via psql untuk menghindari error splitting.
# ============================================================

set -e

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-magnum}"
DB_PASSWORD="${DB_PASSWORD:-Pass@w0rd}"
DB_NAME="${DB_NAME:-hrms}"

MIGRATIONS_DIR="database/migrations"

echo "🚀 Running all migrations via psql..."
echo "   Host: $DB_HOST:$DB_PORT"
echo "   Database: $DB_NAME"
echo "   User: $DB_USER"
echo ""

# Export password for psql
export PGPASSWORD="$DB_PASSWORD"

# Run each migration file in order
for f in $(ls "$MIGRATIONS_DIR"/*.sql | sort); do
    filename=$(basename "$f")
    echo "⏳ Applying: $filename"
    
    # Extract the Up section (between -- +goose Up and -- +goose Down or end of file)
    # Use sed to grab everything between +goose Up and the next +goose Down (or EOF)
    sql=$(awk '/^-- \+goose Up$/,/^-- \+goose Down$/' "$f" | grep -v "^-- \+goose \(Up\|Down\)$")
    
    if [ -z "$sql" ]; then
        echo "   ⚠️  No Up section found in $filename, skipping..."
        continue
    fi
    
    echo "$sql" | psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -q -1 2>&1 || {
        echo "   ❌ Error applying $filename"
        exit 1
    }
    
    echo "   ✅ Applied: $filename"
done

# Record migrations in goose_db_version so goose knows they're applied
echo ""
echo "📝 Recording migrations in goose_db_version table..."

# Ensure goose_db_version table exists
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
CREATE TABLE IF NOT EXISTS goose_db_version (
    id SERIAL PRIMARY KEY,
    version_id BIGINT NOT NULL,
    is_applied BOOLEAN NOT NULL,
    tstamp TIMESTAMPTZ DEFAULT NOW()
);" -q 2>&1

# Get list of migration files and insert them
idx=0
for f in $(ls "$MIGRATIONS_DIR"/*.sql | sort); do
    filename=$(basename "$f")
    idx=$((idx + 1))
    
    # Check if already recorded
    exists=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM goose_db_version WHERE version_id = $idx;" | tr -d ' ')
    if [ "$exists" = "0" ]; then
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "INSERT INTO goose_db_version (version_id, is_applied) VALUES ($idx, TRUE);" -q 2>&1
        echo "   ✅ Recorded: $filename (version $idx)"
    else
        echo "   ⏭️  Already recorded: $filename (version $idx)"
    fi
done

echo ""
echo "✅ All migrations applied successfully!"
echo ""

unset PGPASSWORD
