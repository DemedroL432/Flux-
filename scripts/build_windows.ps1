param(
  [string]$Version = "0.7.4",
  [string]$GitHubOwner = "DemedroL432",
  [string]$GitHubRepo = "Flux-",
  [string]$CoreBinaryPath = ".\\third_party\\core\\sing-box.exe",
  [string]$DpiBinaryPath = ".\\third_party\\dpi\\goodbyedpi.exe",
  [string]$UiMode = "wails",
  [switch]$BuildInstaller
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$dist = Join-Path $root "dist"
$outDir = Join-Path $dist ("Flux-v" + $Version + "-win64")
if (Test-Path -LiteralPath $outDir) {
  $outDir = Join-Path $dist ("Flux-v" + $Version + "-win64-" + (Get-Date -Format "yyyyMMdd-HHmmss"))
}
$coreDir = Join-Path $outDir "core"
$dpiDir = Join-Path $outDir "dpi"
$zapretDir = Join-Path $outDir "zapret"

New-Item -ItemType Directory -Path $coreDir -Force | Out-Null
New-Item -ItemType Directory -Path $dpiDir -Force | Out-Null
New-Item -ItemType Directory -Path $zapretDir -Force | Out-Null

Push-Location $root
try {
  if ($UiMode -eq "wails") {
    $wails = Join-Path $env:USERPROFILE "go\\bin\\wails.exe"
    if (!(Test-Path $wails)) {
      throw "Wails CLI not found at $wails. Install with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    }
    $uiDir = Join-Path $root "ui\\PulseShieldDesktop"
    if (!(Test-Path $uiDir)) {
      throw "Wails UI project not found at $uiDir"
    }
    Push-Location $uiDir
    try {
      $brandLogo = Join-Path $root "resources\branding\pulseshield-logo.png"
      $appicon = Join-Path $uiDir "build\appicon.png"
      if (Test-Path $brandLogo) {
        Copy-Item $brandLogo $appicon -Force
      } elseif (!(Test-Path $appicon)) {
        Write-Warning "Branding asset missing: $brandLogo and $appicon - tray/Wails icon may fail to embed."
      }
      # Wails не всегда пересобирает icon.ico при смене appicon.png — остаётся бирюзовая иконка по умолчанию.
      $iconIco = Join-Path $uiDir "build\windows\icon.ico"
      if (Test-Path $iconIco) {
        Remove-Item -LiteralPath $iconIco -Force
      }
      $ldflags = "-X main.Version=$Version -X main.GitHubOwner=$GitHubOwner -X main.GitHubRepo=$GitHubRepo"
      & $wails build -platform windows/amd64 -ldflags $ldflags
      if ($LASTEXITCODE -ne 0) { throw "wails build failed" }
      $uiExe = Join-Path $uiDir "build\\bin\\Flux.exe"
      if (!(Test-Path $uiExe)) {
        $uiExe = Join-Path $uiDir "build\\bin\\PulseShieldDesktop.exe"
      }
      if (!(Test-Path $uiExe)) { throw "Wails exe not found: $uiExe" }
      Copy-Item $uiExe (Join-Path $outDir "Flux.exe") -Force
      Copy-Item $uiExe (Join-Path $outDir "PulseShield.exe") -Force
    }
    finally {
      Pop-Location
    }
  } else {
    $env:CGO_ENABLED = "0"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"

    go mod tidy
    if ($LASTEXITCODE -ne 0) { throw "go mod tidy failed" }
    go build -ldflags "-H=windowsgui -s -w -X main.version=$Version" -o (Join-Path $outDir "Flux.exe") ./cmd/pulseshield
    if ($LASTEXITCODE -ne 0) { throw "go build failed" }
    Copy-Item (Join-Path $outDir "Flux.exe") (Join-Path $outDir "PulseShield.exe") -Force
  }

  # Walk requires Common-Controls v6 manifest for stable tooltips.
  $manifest = ".\\resources\\PulseShield.exe.manifest"
  if (Test-Path $manifest) {
    Copy-Item $manifest (Join-Path $outDir "Flux.exe.manifest") -Force
    Copy-Item $manifest (Join-Path $outDir "PulseShield.exe.manifest") -Force
  }

  if (!(Test-Path $CoreBinaryPath)) {
    throw "Core binary not found: $CoreBinaryPath. Place sing-box.exe there before packaging."
  }
  Copy-Item $CoreBinaryPath (Join-Path $coreDir "sing-box.exe") -Force

  $wintunDll = ".\\third_party\\core\\wintun.dll"
  if (Test-Path $wintunDll) {
    Copy-Item $wintunDll (Join-Path $coreDir "wintun.dll") -Force
    Copy-Item $wintunDll (Join-Path $outDir "wintun.dll") -Force
  } else {
    Write-Warning "third_party\\core\\wintun.dll not found - place Wintun DLL next to sing-box for TUN (https://www.wintun.net)."
  }

  if (Test-Path $DpiBinaryPath) {
    Copy-Item $DpiBinaryPath (Join-Path $dpiDir "goodbyedpi.exe") -Force
  }

  Copy-Item ".\\configs" (Join-Path $outDir "configs") -Recurse -Force
  Copy-Item ".\\README.md" (Join-Path $outDir "README.md") -Force
  Copy-Item ".\\ROUTING_GUIDE.md" (Join-Path $outDir "ROUTING_GUIDE.md") -Force
  Copy-Item ".\\QUICK_TEST.md" (Join-Path $outDir "QUICK_TEST.md") -Force

  # Bundle zapret runtime (winws2 presets/lua/lists/filters) if available
  $zapretSrc = ".\\third_party\\zapret"
  if (Test-Path $zapretSrc) {
    $zapretNested = Join-Path $zapretSrc "zapret"
    if (Test-Path $zapretNested) {
      Copy-Item (Join-Path $zapretNested "*") $zapretDir -Recurse -Force
    } else {
      Copy-Item (Join-Path $zapretSrc "*") $zapretDir -Recurse -Force
    }
  } else {
    Write-Warning "third_party\\zapret not found - portable will have no DPI lists. Clone Zapret into repo before packaging."
  }
  $listsProbe = Join-Path $zapretDir "lists"
  if (-not (Test-Path $listsProbe)) {
    Write-Warning "Portable bundle: zapret lists folder missing under $zapretDir (DPI tab may be empty)."
  } elseif (Test-Path $zapretSrc) {
    Write-Host "Portable bundle: zapret + lists copied to $zapretDir"
  }

  # Optional: bundle Mesa software OpenGL to avoid GPU/OpenGL driver issues
  $mesaOgl = ".\\third_party\\mesa\\opengl32.dll"
  $mesaWgl = ".\\third_party\\mesa\\libgallium_wgl.dll"
  if (Test-Path $mesaOgl) { Copy-Item $mesaOgl (Join-Path $outDir "opengl32.dll") -Force }
  if (Test-Path $mesaWgl) { Copy-Item $mesaWgl (Join-Path $outDir "libgallium_wgl.dll") -Force }
}
finally {
  Pop-Location
}

Write-Host "Portable build ready: $outDir"

$zipOut = Join-Path $dist ("Flux-v" + $Version + "-win64.zip")
Write-Host "[Flux] Creating portable ZIP: $zipOut" -ForegroundColor Cyan
if (Test-Path $zipOut) { Remove-Item $zipOut -Force }
Compress-Archive -Path (Join-Path $outDir "*") -DestinationPath $zipOut -Force
if (Test-Path $zipOut) {
    $zipSize = [math]::Round((Get-Item $zipOut).Length / 1MB, 2)
    Write-Host "[Flux] Portable ZIP created: $zipOut ($zipSize MB)" -ForegroundColor Green
} else {
    Write-Error "Failed to create portable ZIP archive: $zipOut"
}

if ($BuildInstaller) {
  $nsi = Join-Path $PSScriptRoot "installer\\pulseshield.nsi"
  if (!(Test-Path $nsi)) { throw "NSIS script not found: $nsi" }
  $makensisExe = $null
  $mc = Get-Command makensis.exe -ErrorAction SilentlyContinue
  if ($mc) { $makensisExe = $mc.Source }
  if (-not $makensisExe) {
    $pf86 = [Environment]::GetFolderPath("ProgramFilesX86")
    $fb = Join-Path $pf86 "NSIS\makensis.exe"
    if (Test-Path $fb) { $makensisExe = $fb }
  }
  if (-not $makensisExe) {
    Write-Warning "makensis.exe not found (PATH or Program Files (x86)\NSIS). Install NSIS: https://nsis.sourceforge.io/"
  } else {
    $releaseTag = "v$Version"
    $setupOut = Join-Path $dist "Flux-$releaseTag-Setup.exe"
    if (Test-Path $setupOut) { Remove-Item -Force $setupOut }
    $srcAbs = (Resolve-Path $outDir).Path
    $outAbs = (Join-Path (Resolve-Path $dist).Path "Flux-$releaseTag-Setup.exe")
    & $makensisExe "/DSOURCE_DIR=$srcAbs" "/DVERSION=$Version" "/DOUTFILE=$outAbs" $nsi
    if ($LASTEXITCODE -ne 0) { throw "makensis failed with exit $LASTEXITCODE" }
    Write-Host "Installer: $outAbs"
    $sumScript = Join-Path $PSScriptRoot "write_release_checksums.ps1"
    if (Test-Path $sumScript) {
      & $sumScript -DistRoot $dist -ReleaseLabel $releaseTag
    }
  }
}
