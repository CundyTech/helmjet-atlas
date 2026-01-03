#!/bin/bash

# Seed MongoDB with mock data for Helmjet Atlas
# Usage: ./seed-mongodb.sh

export MONGO_URI="${MONGO_URI:-mongodb://localhost:27017}"
export MONGO_DB="${MONGO_DB:-helmjet-atlas}"

echo "🌱 Seeding MongoDB with mock data..."
echo "MongoDB URI: $MONGO_URI"
echo "Database: $MONGO_DB"

# Run the Go seed script
cd "$(dirname "$0")"
go run seed-mongodb.go

echo ""
echo "✅ Complete! Visit http://localhost:8000 to view the topology"
