# Корневая точка входа для сборки Windows (обёртка над scripts/build_windows.ps1).
# Пример: .\build.ps1 -Version 0.7.1 -BuildInstaller
param(
  [string]$Version = "0.7.5",
  [string]$GitHubOwner = "DemedroL432",
  [string]$GitHubRepo = "Flux-",
  [string]$CoreBinaryPath = ".\third_party\core\sing-box.exe",
  [string]$DpiBinaryPath = ".\third_party\dpi\goodbyedpi.exe",
  [string]$UiMode = "wails",
  [switch]$BuildInstaller
)

$ErrorActionPreference = "Stop"
$scriptPath = Join-Path $PSScriptRoot "scripts\build_windows.ps1"
if (-not (Test-Path $scriptPath)) {
  throw "Не найден $scriptPath"
}

$invokeArgs = @{
  Version         = $Version
  GitHubOwner     = $GitHubOwner
  GitHubRepo      = $GitHubRepo
  CoreBinaryPath  = $CoreBinaryPath
  DpiBinaryPath   = $DpiBinaryPath
  UiMode          = $UiMode
}
if ($BuildInstaller) {
  $invokeArgs.BuildInstaller = $true
}

& $scriptPath @invokeArgs
