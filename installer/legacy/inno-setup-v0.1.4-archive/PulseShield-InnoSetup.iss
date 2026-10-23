; Legacy Inno Setup (архив). Пути относительно этой папки: корень репозитория = ..\..\..\ 
; Текущая сборка: NSIS — scripts\installer\pulseshield.nsi

[Setup]
AppName=PulseShield
AppVersion=0.1.4
AppPublisher=DemedroL
DefaultDirName={pf}\PulseShield
DefaultGroupName=PulseShield
OutputBaseFilename=PulseShieldSetup-0.1.4-inno-legacy
Compression=lzma
SolidCompression=yes
ArchitecturesAllowed=x64
ArchitecturesInstallIn64BitMode=x64

[Files]
Source: "..\..\..\dist\pulseshield-win64\PulseShield.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\..\dist\pulseshield-win64\PulseShield.exe.manifest"; DestDir: "{app}"; Flags: ignoreversion; Check: FileExists(ExpandConstant('{src}\..\..\..\dist\pulseshield-win64\PulseShield.exe.manifest'))
Source: "..\..\..\dist\pulseshield-win64\opengl32.dll"; DestDir: "{app}"; Flags: ignoreversion; Check: FileExists(ExpandConstant('{src}\..\..\..\dist\pulseshield-win64\opengl32.dll'))
Source: "..\..\..\dist\pulseshield-win64\libgallium_wgl.dll"; DestDir: "{app}"; Flags: ignoreversion; Check: FileExists(ExpandConstant('{src}\..\..\..\dist\pulseshield-win64\libgallium_wgl.dll'))
Source: "..\..\..\dist\pulseshield-win64\README.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\..\dist\pulseshield-win64\ROUTING_GUIDE.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\..\dist\pulseshield-win64\QUICK_TEST.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\..\dist\pulseshield-win64\configs\*"; DestDir: "{app}\configs"; Flags: recursesubdirs ignoreversion
Source: "..\..\..\dist\pulseshield-win64\core\*"; DestDir: "{app}\core"; Flags: recursesubdirs ignoreversion
Source: "..\..\..\dist\pulseshield-win64\dpi\*"; DestDir: "{app}\dpi"; Flags: recursesubdirs ignoreversion
Source: "..\..\..\dist\pulseshield-win64\zapret\*"; DestDir: "{app}\zapret"; Flags: recursesubdirs ignoreversion; Check: DirExists(ExpandConstant('{src}\..\..\..\dist\pulseshield-win64\zapret'))

[Icons]
Name: "{group}\PulseShield"; Filename: "{app}\PulseShield.exe"
Name: "{commondesktop}\PulseShield"; Filename: "{app}\PulseShield.exe"; Tasks: desktopicon

[Tasks]
Name: "desktopicon"; Description: "Create a &desktop icon"; GroupDescription: "Additional icons:"; Flags: unchecked

[Run]
Filename: "{app}\PulseShield.exe"; Description: "Launch PulseShield"; Flags: nowait postinstall skipifsilent
