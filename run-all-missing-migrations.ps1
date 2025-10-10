# PowerShell script to run all missing migrations
# This will create all required tables

Write-Host "=== Running All Missing Migrations ===" -ForegroundColor Cyan
Write-Host ""

# Load environment variables from .env
if (Test-Path ".env") {
    Get-Content .env | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.+)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim()
            [Environment]::SetEnvironmentVariable($name, $value, "Process")
        }
    }
}

$DB_HOST = $env:DB_HOST
$DB_PORT = $env:DB_PORT
$DB_USER = $env:DB_USER
$DB_PASSWORD = $env:DB_PASSWORD
$DB_NAME = $env:DB_NAME

Write-Host "Database: $DB_NAME @ $DB_HOST:$DB_PORT" -ForegroundColor Yellow
Write-Host ""

# Set PGPASSWORD
$env:PGPASSWORD = $DB_PASSWORD

# Find psql
$psqlPath = $null
$possiblePaths = @(
    "C:\Program Files\PostgreSQL\16\bin\psql.exe",
    "C:\Program Files\PostgreSQL\15\bin\psql.exe",
    "C:\Program Files\PostgreSQL\14\bin\psql.exe",
    "C:\Program Files\PostgreSQL\13\bin\psql.exe"
)

foreach ($path in $possiblePaths) {
    if (Test-Path $path) {
        $psqlPath = $path
        break
    }
}

if (-not $psqlPath) {
    Write-Host "❌ psql not found!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Alternative: Run SQL manually" -ForegroundColor Yellow
    Write-Host "  File: create-missing-tables.sql"
    Write-Host ""
    exit 1
}

Write-Host "✅ Found psql: $psqlPath" -ForegroundColor Green
Write-Host ""

# Run the combined migration script
Write-Host "Creating missing tables..." -ForegroundColor Cyan
& $psqlPath -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "create-missing-tables.sql"

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ All tables created successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "  1. Restart server: .\bin\server.exe"
    Write-Host "  2. Test pairing code"
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "❌ Migration failed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Try manual SQL:" -ForegroundColor Yellow
    Write-Host "  1. Open pgAdmin"
    Write-Host "  2. Run: create-missing-tables.sql"
    Write-Host ""
}
