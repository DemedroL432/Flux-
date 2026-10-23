# Обёртка над корневым build.ps1: установщик NSIS (см. scripts/installer/pulseshield.nsi и RELEASING.md).
# Раньше здесь вызывался Inno Setup и vpnboost.iss — больше не используется.
param(
  [string]$Version = "0.7.0",
  [string]$GitHubOwner = "DemedroL",
  [string]$GitHubRepo = "pulseshield"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$buildPs1 = Join-Path $root "build.ps1"
if (!(Test-Path $buildPs1)) {
  throw "Не найден $buildPs1"
}
& $buildPs1 -Version $Version -GitHubOwner $GitHubOwner -GitHubRepo $GitHubRepo -BuildInstaller

