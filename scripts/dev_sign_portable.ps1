# Dev-only: self-signed Authenticode for portable folder (testing). Requires Windows SDK (signtool.exe).
# Install SDK: https://developer.microsoft.com/windows/downloads/windows-sdk/ then re-run this script.
param(
  [string]$PortableRoot = "",
  [string]$PfxPath = "",
  [string]$PfxPlainPassword = "PulseShieldDevTest-0.3.0"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
if ($PortableRoot -eq "") {
  $PortableRoot = Join-Path $root "dist\pulseshield-win64"
}
if (!(Test-Path -LiteralPath $PortableRoot)) {
  throw "Portable folder not found: $PortableRoot"
}

$signScript = Join-Path $PSScriptRoot "sign_windows_binaries.ps1"
if (!(Test-Path $signScript)) {
  throw "Missing $signScript"
}

if ($PfxPath -eq "") {
  $PfxPath = Join-Path $env:TEMP "pulseshield-dev-test.pfx"
}
if (!(Test-Path -LiteralPath $PfxPath)) {
  Write-Host "Creating dev code-signing cert and PFX at $PfxPath ..."
  Get-ChildItem Cert:\CurrentUser\My -ErrorAction SilentlyContinue |
    Where-Object { $_.Subject -eq "CN=PulseShield Dev Test" } |
    Remove-Item -Force
  $cert = New-SelfSignedCertificate -Type CodeSigningCert -Subject "CN=PulseShield Dev Test" `
    -NotAfter (Get-Date).AddYears(3) -CertStoreLocation "Cert:\CurrentUser\My"
  $sec = ConvertTo-SecureString -String $PfxPlainPassword -AsPlainText -Force
  Export-PfxCertificate -Cert $cert -FilePath $PfxPath -Password $sec -Force
}

$exes = @(Get-ChildItem -Path $PortableRoot -Recurse -Filter "*.exe" | ForEach-Object { $_.FullName })
if ($exes.Count -eq 0) {
  throw "No .exe under $PortableRoot"
}

$pwd = ConvertTo-SecureString -String $PfxPlainPassword -AsPlainText -Force
& $signScript -Files $exes -PfxPath $PfxPath -PfxPassword $pwd
Write-Host "Signed $($exes.Count) file(s). (Dev cert: not trusted by Windows by default; for tests only.)"
