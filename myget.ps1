$ErrorActionPreference = "Stop"

# Read in the version variables set by the pre-build script
$Version = ${env:GitVersion.SemVer}
$NuGetVersion = ${env:GitVersion.NuGetVersion}
echo "Version: $Version"
echo "NuGetVersion: $NuGetVersion"

# Download the Windows 64-bit binary
$url = "https://download.getcarina.com/carina/v$Version/Windows/x86_64/carina.exe"
echo "Downloading $url"
iwr $url -outfile script\chocolatey\carina.exe

# Package the binary into a Chocolatey installer
choco pack script\chocolatey\carina.nuspec --Version $NuGetVersion

# Publish the package to https://chocolatey.org/packages/carina
choco push carina.$NuGetVersion.nupkg --api-key $env:CHOCOLATEY_APIKEY