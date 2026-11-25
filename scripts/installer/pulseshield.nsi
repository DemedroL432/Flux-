; Full-tree installer for portable folder (PulseShield.exe + core/dpi/zapret/configs).
; Build after scripts/build_windows.ps1:
;   makensis /DSOURCE_DIR="C:\path\to\dist\pulseshield-win64" /DVERSION=1.0.0 /DOUTFILE="C:\path\PulseShield-v1.0.0-Setup.exe" pulseshield.nsi

!ifndef SOURCE_DIR
  !error "Pass /DSOURCE_DIR=full path to portable folder"
!endif
!ifndef VERSION
  !define VERSION "0.0.0"
!endif
!ifndef OUTFILE
  !define OUTFILE "PulseShield-Setup.exe"
!endif

!define PRODUCT_NAME "Flux"
!define UNINST_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\Flux"

Unicode true
RequestExecutionLevel admin
Name "${PRODUCT_NAME} ${VERSION}"
OutFile "${OUTFILE}"
InstallDir "$PROGRAMFILES64\Flux"
InstallDirRegKey HKLM "${UNINST_KEY}" "InstallLocation"

!include "MUI2.nsh"

!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"

Section "Main"
  nsExec::ExecToStack 'taskkill /F /IM Flux.exe /IM PulseShield.exe /IM PulseShieldDesktop.exe /T'
  Sleep 800
  Delete "$DESKTOP\PulseShield.lnk"
  Delete "$COMMONDESKTOP\PulseShield.lnk"
  RMDir /r "$SMPROGRAMS\PulseShield"
  RMDir /r "$PROGRAMFILES64\PulseShield"
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\PulseShield"

  SetOutPath $INSTDIR
  File /r "${SOURCE_DIR}\*.*"
  WriteRegStr HKLM "${UNINST_KEY}" "DisplayName" "${PRODUCT_NAME}"
  WriteRegStr HKLM "${UNINST_KEY}" "DisplayVersion" "${VERSION}"
  WriteRegStr HKLM "${UNINST_KEY}" "Publisher" "Flux"
  WriteRegStr HKLM "${UNINST_KEY}" "InstallLocation" "$INSTDIR"
  WriteRegStr HKLM "${UNINST_KEY}" "UninstallString" "$\"$INSTDIR\Uninstall.exe$\""
  WriteUninstaller "$INSTDIR\Uninstall.exe"
  CreateDirectory "$SMPROGRAMS\Flux"
  CreateShortcut "$SMPROGRAMS\Flux\Flux.lnk" "$INSTDIR\Flux.exe"
  CreateShortcut "$DESKTOP\Flux.lnk" "$INSTDIR\Flux.exe"

  IfSilent 0 +2
    Exec '"$INSTDIR\Flux.exe"'
SectionEnd

Section "Uninstall"
  nsExec::ExecToStack 'taskkill /F /IM Flux.exe /IM PulseShield.exe /IM PulseShieldDesktop.exe /T'
  nsExec::ExecToStack 'schtasks /Delete /TN "FluxAutostart" /F'
  Delete "$DESKTOP\Flux.lnk"
  RMDir /r "$SMPROGRAMS\Flux"
  RMDir /r "$INSTDIR"
  DeleteRegKey HKLM "${UNINST_KEY}"
SectionEnd
