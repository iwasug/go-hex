# Load environment variables from .env file
if (Test-Path ".env") {
    Get-Content .env | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.+)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim()
            [Environment]::SetEnvironmentVariable($name, $value, "Process")
            Write-Host "Loaded: $name"
        }
    }
}

# Get database connection info
$DB_HOST = $env:DB_HOST
$DB_PORT = $env:DB_PORT
$DB_USER = $env:DB_USER
$DB_PASSWORD = $env:DB_PASSWORD
$DB_NAME = $env:DB_NAME

Write-Host ""
Write-Host "Database Info:"
Write-Host "  Host: $DB_HOST"
Write-Host "  Port: $DB_PORT"
Write-Host "  User: $DB_USER"
Write-Host "  Database: $DB_NAME"
Write-Host ""

# Set PGPASSWORD environment variable for psql
$env:PGPASSWORD = $DB_PASSWORD

# Find psql executable
$psqlPath = $null
$possiblePaths = @(
    "C:\Program Files\PostgreSQL\16\bin\psql.exe",
    "C:\Program Files\PostgreSQL\15\bin\psql.exe",
    "C:\Program Files\PostgreSQL\14\bin\psql.exe",
    "C:\Program Files\PostgreSQL\13\bin\psql.exe",
    "C:\PostgreSQL\16\bin\psql.exe",
    "C:\PostgreSQL\15\bin\psql.exe"
)

foreach ($path in $possiblePaths) {
    if (Test-Path $path) {
        $psqlPath = $path
        Write-Host "Found psql at: $psqlPath"
        break
    }
}

if (-not $psqlPath) {
    Write-Host "ERROR: psql not found. Please install PostgreSQL or add psql to PATH."
    Write-Host ""
    Write-Host "Alternative: Run SQL manually in your database tool:"
    Write-Host "  File: migrations\000009_add_whatsapp_lid_mappings.up.sql"
    exit 1
}

# Run the migration
$migrationFile = "migrations\000009_add_whatsapp_lid_mappings.up.sql"

if (Test-Path $migrationFile) {
    Write-Host ""
    Write-Host "Running migration: $migrationFile"
    Write-Host ""
    
    & $psqlPath -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $migrationFile
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "✅ Migration completed successfully!"
        Write-Host ""
        Write-Host "Next steps:"
        Write-Host "1. Restart server"
        Write-Host "2. Test pairing code again"
    } else {
        Write-Host ""
        Write-Host "❌ Migration failed with exit code: $LASTEXITCODE"
        Write-Host ""
        Write-Host "Try running SQL manually:"
        Write-Host "  1. Open pgAdmin or database tool"
        Write-Host "  2. Run content from: $migrationFile"
    }
} else {
    Write-Host "ERROR: Migration file not found: $migrationFile"
    exit 1
}
