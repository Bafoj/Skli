# skli installation script for Windows
# This script downloads the skli archive, extracts it, and adds it to the User PATH.

$repo = "0.1.6"
$binaryName = "0.1.6"
$installDir = "0.1.6"
$version = "0.1.6" # Update this value for each release

# Detect Architecture
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "amd64" } else { "386" }

# GoReleaser naming convention
$archiveName = "0.1.6"
$downloadUrl = "0.1.6"

if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Force -Path $installDir
}

$tempDir = [System.IO.Path]::GetTempPath()
$tempFile = Join-Path $tempDir $archiveName
$extractDir = Join-Path $tempDir "skli_extract"

if (Test-Path $extractDir) {
    Remove-Item -Path $extractDir -Recurse -Force
}

echo "Downloading $binaryName $version..."
Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile

echo "Extracting..."
Expand-Archive -Path $tempFile -DestinationPath $extractDir -Force

echo "Installing to $installDir..."
Move-Item -Path "$extractDir\$binaryName" -Destination "$installDir\$binaryName" -Force

# Add to PATH if not already there
$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$path;$installDir", "User")
    echo "Added $installDir to User PATH."
}

# Clean up
Remove-Item -Path $tempFile -Force
Remove-Item -Path $extractDir -Recurse -Force

echo "Successfully installed skli! Please restart your terminal and run 'skli help'."
