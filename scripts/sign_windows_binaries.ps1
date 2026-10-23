# Подпись PE-файлов Authenticode через signtool (Windows SDK).
# Сертификат: PFX + пароль или расширьте скрипт под -Thumbprint и LocalMachine\My.
# Проверка: signtool verify /pa /v <file.exe>
param(
  [Parameter(Mandatory = $true)]
  [string[]]$Files,
  [string]$PfxPath = "",
  [System.Security.SecureString]$PfxPassword,
  [string]$TimestampUrl = "http://timestamp.digicert.com",
  [string]$SigntoolPath = ""
)

$ErrorActionPreference = "Stop"

function Find-SignTool {
  if ($SigntoolPath -ne "" -and (Test-Path -LiteralPath $SigntoolPath)) {
    return (Resolve-Path -LiteralPath $SigntoolPath).Path
  }
  $kits = @(
    Join-Path ${env:ProgramFiles(x86)} "Windows Kits\10\bin"
    Join-Path $env:ProgramFiles "Windows Kits\10\bin"
  )
  foreach ($kit in $kits) {
    if (!(Test-Path $kit)) { continue }
    $exe = Get-ChildItem -Path $kit -Recurse -Filter "signtool.exe" -ErrorAction SilentlyContinue |
      Where-Object { $_.FullName -match '\\x64\\signtool\.exe$' } |
      Sort-Object FullName -Descending |
      Select-Object -First 1
    if ($exe) { return $exe.FullName }
  }
  return $null
}

$signtool = Find-SignTool
if (-not $signtool) {
  throw "signtool.exe not found. Install Windows SDK (Desktop C++ workload or standalone SDK) or pass -SigntoolPath."
}

foreach ($rel in $Files) {
  if (!(Test-Path -LiteralPath $rel)) {
    throw "File not found: $rel"
  }
}

if ($PfxPath -eq "") {
  throw "Specify -PfxPath to your code-signing PFX."
}

if ($null -eq $PfxPassword) {
  $PfxPassword = Read-Host "PFX password" -AsSecureString
}

$BSTR = [IntPtr]::Zero
try {
  $BSTR = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($PfxPassword)
  $plain = [Runtime.InteropServices.Marshal]::PtrToStringAuto($BSTR)

  foreach ($rel in $Files) {
    $full = (Resolve-Path -LiteralPath $rel).Path
    Write-Host "Signing $full ..."
    $argList = @(
      "sign", "/fd", "SHA256",
      "/tr", $TimestampUrl, "/td", "SHA256",
      "/f", $PfxPath,
      "/p", $plain,
      $full
    )
    & $signtool @argList
    if ($LASTEXITCODE -ne 0) {
      throw "signtool failed with exit $LASTEXITCODE for $full"
    }
  }
}
finally {
  if ($BSTR -ne [IntPtr]::Zero) {
    [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($BSTR)
  }
}

Write-Host "Done. Verify: signtool verify /pa /v `"$($Files[0])`""
