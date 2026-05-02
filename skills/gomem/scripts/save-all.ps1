# GoMem save-all — index entire project into persistent memory
# Usage: .\save-all.ps1 [project-directory]
# If no directory given, indexes the current directory.

param(
    [string]$ProjectDir = ""
)

if (-not $ProjectDir) {
    $ProjectDir = Get-Location
}

# Resolve to absolute path
$ProjectDir = Resolve-Path $ProjectDir -ErrorAction Stop

# Find gomem binary
$GomemBin = Join-Path $PSScriptRoot "..\..\gomem.exe"
if (-not (Test-Path $GomemBin)) {
    $GomemBin = (Get-Command gomem -ErrorAction SilentlyContinue).Source
    if (-not $GomemBin) {
        Write-Error "gomem binary not found. Build it first: cd /path/to/gomem && go build -o gomem ./cmd/gomem"
        exit 1
    }
}

Write-Host "=== GoMem: Indexing $ProjectDir ==="
Write-Host ""

& $GomemBin save-all $ProjectDir
