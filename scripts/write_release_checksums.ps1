# Записывает SHA256 для артефактов релиза (ZIP portable и NSIS Setup) в dist/.
# Формат строк: <hex> *<имя_файла> (как в coreutils sha256sum).
param(
  [Parameter(Mandatory = $false)]
  [string]$DistRoot = "dist",
  # Метка для имени файла списка, напр. v0.2.0 (как git tag). Если пусто — SHA256SUMS.txt
  [string]$ReleaseLabel = ""
)

$ErrorActionPreference = "Stop"
$dist = Resolve-Path $DistRoot

$artifacts = Get-ChildItem -LiteralPath $dist -File -ErrorAction SilentlyContinue | Where-Object {
  $n = $_.Name
  if ($ReleaseLabel -ne "" -and $n -notlike "*-$ReleaseLabel-*") { return $false }
  ($n -match '^(Flux|PulseShield)-.+-win64\.zip$') -or ($n -match '^(Flux|PulseShield)-.+-Setup\.exe$')
} | Sort-Object Name

if ($artifacts.Count -eq 0) {
  Write-Warning "No Flux-*-win64.zip or Flux-*-Setup.exe in $dist - checksums skipped."
  exit 0
}

$lines = foreach ($f in $artifacts) {
  $h = Get-FileHash -Algorithm SHA256 -LiteralPath $f.FullName
  "$($h.Hash.ToLower()) *$($f.Name)"
}

if ($ReleaseLabel -ne "") {
  $outName = "Flux-$ReleaseLabel-SHA256SUMS.txt"
} else {
  $outName = "SHA256SUMS.txt"
}
$outPath = Join-Path $dist $outName
$lines | Set-Content -LiteralPath $outPath -Encoding utf8
Write-Host "Written $outPath - $($artifacts.Count) artifact(s)"
