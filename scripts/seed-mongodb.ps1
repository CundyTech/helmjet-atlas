# Seed MongoDB with mock data for Helmjet Atlas
# Usage: .\seed-mongodb.ps1

param(
    [string]$MongoUri = "mongodb://localhost:27017",
    [string]$MongoDb = "helmjet-atlas"
)

$env:MONGO_URI = $MongoUri
$env:MONGO_DB = $MongoDb

Write-Host "🌱 Seeding MongoDB with mock data..." -ForegroundColor Green
Write-Host "MongoDB URI: $MongoUri"
Write-Host "Database: $MongoDb"
Write-Host ""

# Run the Go seed script
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
Push-Location $scriptPath

go run seed-mongodb.go

Pop-Location

Write-Host ""
Write-Host "✅ Complete! Your database now has mock data." -ForegroundColor Green
Write-Host "Visit http://localhost:8000 to view the topology" -ForegroundColor Cyan
