; ==============================================================================
; Game Trainer Hub - NSIS Installation Script
; ==============================================================================

Unicode True
SetCompressor /SOLID lzma

; ------------------------------------------------------------------------------
; Definitions
; ------------------------------------------------------------------------------
!define PRODUCT_NAME "Game Trainer Hub"
!define PRODUCT_VERSION "1.0.0"
!define PRODUCT_PUBLISHER "GameTrainerHub"
!define PRODUCT_WEB_SITE "https://github.com/FLiNGTrainer"
!define PRODUCT_DIR_REGKEY "Software\Microsoft\Windows\CurrentVersion\App Paths\GameTrainerHub.exe"
!define PRODUCT_UNINST_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\GameTrainerHub"
!define PRODUCT_UNINST_ROOT_KEY "HKLM"
!define APP_EXE "GameTrainerHub.exe"

; ------------------------------------------------------------------------------
; Includes
; ------------------------------------------------------------------------------
!include "MUI2.nsh"
!include "x64.nsh"
!include "LogicLib.nsh"
!include "FileFunc.nsh"

; ------------------------------------------------------------------------------
; General Settings
; ------------------------------------------------------------------------------
Name "${PRODUCT_NAME} ${PRODUCT_VERSION}"
OutFile "..\..\bin\GameTrainerHub-Setup-x64.exe"
InstallDir "$PROGRAMFILES64\GameTrainerHub"
InstallDirRegKey ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "InstallLocation"
RequestExecutionLevel admin
ShowInstDetails show
ShowUnInstDetails show

; ------------------------------------------------------------------------------
; UI & Branding
; ------------------------------------------------------------------------------
!define MUI_ABORTWARNING
!define MUI_ICON "..\windows\icon.ico"
!define MUI_UNICON "..\windows\icon.ico"

; ------------------------------------------------------------------------------
; Installer Pages
; ------------------------------------------------------------------------------
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "NOTICE.txt"
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES

!define MUI_FINISHPAGE_RUN "$INSTDIR\${APP_EXE}"
!define MUI_FINISHPAGE_RUN_TEXT "运行 ${PRODUCT_NAME}"
!insertmacro MUI_PAGE_FINISH

; ------------------------------------------------------------------------------
; Uninstaller Pages
; ------------------------------------------------------------------------------
!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

; ------------------------------------------------------------------------------
; Language
; ------------------------------------------------------------------------------
!insertmacro MUI_LANGUAGE "SimpChinese"

; ------------------------------------------------------------------------------
; Functions - Running Process Check
; ------------------------------------------------------------------------------
; ------------------------------------------------------------------------------
; Functions - Running Process Check & Graceful Shutdown
; ------------------------------------------------------------------------------
Function CheckRunningProcess
  ${If} ${Silent}
    ; In silent mode: attempt graceful close first, fallback to force kill
    FindWindow $0 "wailsWindow" ""
    ${If} $0 != 0
      System::Call 'user32::PostMessage(i $0, i 0x0012, i 0, i 0)'
    ${EndIf}
    FindWindow $0 "GameTrainerHub_TrayWindow" ""
    ${If} $0 != 0
      System::Call 'user32::PostMessage(i $0, i 0x0012, i 0, i 0)'
    ${EndIf}
    ${DisableX64FSRedirection}
    nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$PLUGINSDIR\graceful_close.ps1"'
    nsExec::Exec 'cmd.exe /c taskkill.exe /IM "${APP_EXE}"'
    ${EnableX64FSRedirection}
    Sleep 1500

    ; Check if still running
    ${DisableX64FSRedirection}
    nsExec::ExecToStack 'cmd.exe /c tasklist /NH /FI "IMAGENAME eq ${APP_EXE}" | findstr /I "${APP_EXE}"'
    Pop $0
    Pop $1
    ${EnableX64FSRedirection}
    ${If} $0 == 0
      ${DisableX64FSRedirection}
      nsExec::Exec 'cmd.exe /c taskkill.exe /F /T /IM "${APP_EXE}"'
      nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "Get-Process GameTrainerHub -ErrorAction SilentlyContinue | Stop-Process -Force"'
      ${EnableX64FSRedirection}
      Sleep 1000
    ${EndIf}
    Return
  ${EndIf}

  check_loop:
    StrCpy $R0 0  ; $R0 = 1 if process is running

    ; --- Method 1: 64-bit tasklist check (bypassing WOW64 redirection) ---
    ${DisableX64FSRedirection}
    nsExec::ExecToStack 'cmd.exe /c tasklist /NH /FI "IMAGENAME eq ${APP_EXE}" | findstr /I "${APP_EXE}"'
    Pop $0
    Pop $1
    ${EnableX64FSRedirection}
    ${If} $0 == 0
      StrCpy $R0 1
    ${EndIf}

    ; --- Method 2: Fallback via 64-bit PowerShell Get-Process ---
    ${If} $R0 == 0
      ${DisableX64FSRedirection}
      nsExec::ExecToStack 'powershell.exe -NoProfile -Command "if (Get-Process GameTrainerHub -ErrorAction SilentlyContinue) { exit 0 } else { exit 1 }"'
      Pop $0
      Pop $1
      ${EnableX64FSRedirection}
      ${If} $0 == 0
        StrCpy $R0 1
      ${EndIf}
    ${EndIf}

    ; If running, prompt user for confirmation
    ${If} $R0 == 1
      MessageBox MB_OKCANCEL|MB_ICONEXCLAMATION "检测到 ${PRODUCT_NAME} 当前正在运行。$\r$\n$\r$\n为了确保安装或升级顺利完成，需要先关闭正在运行的程序。$\r$\n$\r$\n请确认是否关闭程序并继续安装？$\r$\n$\r$\n【确定】：关闭正在运行的程序，并继续安装。$\r$\n【取消】：退出安装程序，不继续安装。" IDOK on_confirm IDCANCEL on_cancel

      on_confirm:
        ; Phase 1: Request graceful shutdown (WM_QUIT & graceful helper)
        FindWindow $0 "wailsWindow" ""
        ${If} $0 != 0
          System::Call 'user32::PostMessage(i $0, i 0x0012, i 0, i 0)'
        ${EndIf}
        FindWindow $0 "GameTrainerHub_TrayWindow" ""
        ${If} $0 != 0
          System::Call 'user32::PostMessage(i $0, i 0x0012, i 0, i 0)'
        ${EndIf}
        ${DisableX64FSRedirection}
        nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$PLUGINSDIR\graceful_close.ps1"'
        nsExec::Exec 'cmd.exe /c taskkill.exe /IM "${APP_EXE}"'
        ${EnableX64FSRedirection}

        ; Phase 2: Wait for process to exit cleanly (up to 6 seconds)
        StrCpy $R1 0
        wait_exit_loop:
          Sleep 500
          IntOp $R1 $R1 + 1

          ; Check if still running
          StrCpy $R0 0
          ${DisableX64FSRedirection}
          nsExec::ExecToStack 'cmd.exe /c tasklist /NH /FI "IMAGENAME eq ${APP_EXE}" | findstr /I "${APP_EXE}"'
          Pop $0
          Pop $1
          ${EnableX64FSRedirection}
          ${If} $0 == 0
            StrCpy $R0 1
          ${EndIf}

          ${If} $R0 == 0
            ; Process exited cleanly!
            Goto process_exited
          ${EndIf}

          ${If} $R1 < 12
            Goto wait_exit_loop
          ${EndIf}

        ; Phase 3: Timeout - Shutdown failed, prompt user with explicit options
        MessageBox MB_YESNOCANCEL|MB_ICONEXCLAMATION "${PRODUCT_NAME} 未能在规定时间内正常退出。$\r$\n$\r$\n是否尝试强制结束该程序？$\r$\n$\r$\n【是】：强制结束程序并继续安装（警告：未保存的数据可能会丢失）。$\r$\n【否】：重新尝试等待正常退出。$\r$\n【取消】：退出安装程序，不执行任何更改。" IDYES do_force_kill IDNO on_confirm
        Goto on_cancel

        do_force_kill:
          ${DisableX64FSRedirection}
          nsExec::Exec 'cmd.exe /c taskkill.exe /F /T /IM "${APP_EXE}"'
          nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "Get-Process GameTrainerHub -ErrorAction SilentlyContinue | Stop-Process -Force"'
          ${EnableX64FSRedirection}
          Sleep 1500
          Goto check_loop

      on_cancel:
        Abort
    ${EndIf}

  process_exited:
FunctionEnd

Function un.CheckRunningProcess
  ${If} ${Silent}
    FindWindow $0 "wailsWindow" ""
    ${If} $0 != 0
      System::Call 'user32::PostMessage(i $0, i 0x0012, i 0, i 0)'
    ${EndIf}
    FindWindow $0 "GameTrainerHub_TrayWindow" ""
    ${If} $0 != 0
      System::Call 'user32::PostMessage(i $0, i 0x0012, i 0, i 0)'
    ${EndIf}
    ${DisableX64FSRedirection}
    nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$PLUGINSDIR\graceful_close.ps1"'
    nsExec::Exec 'cmd.exe /c taskkill.exe /IM "${APP_EXE}"'
    ${EnableX64FSRedirection}
    Sleep 1500

    ${DisableX64FSRedirection}
    nsExec::ExecToStack 'cmd.exe /c tasklist /NH /FI "IMAGENAME eq ${APP_EXE}" | findstr /I "${APP_EXE}"'
    Pop $0
    Pop $1
    ${EnableX64FSRedirection}
    ${If} $0 == 0
      ${DisableX64FSRedirection}
      nsExec::Exec 'cmd.exe /c taskkill.exe /F /T /IM "${APP_EXE}"'
      nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "Get-Process GameTrainerHub -ErrorAction SilentlyContinue | Stop-Process -Force"'
      ${EnableX64FSRedirection}
      Sleep 1000
    ${EndIf}
    Return
  ${EndIf}

  un_check_loop:
    StrCpy $R0 0

    ${DisableX64FSRedirection}
    nsExec::ExecToStack 'cmd.exe /c tasklist /NH /FI "IMAGENAME eq ${APP_EXE}" | findstr /I "${APP_EXE}"'
    Pop $0
    Pop $1
    ${EnableX64FSRedirection}
    ${If} $0 == 0
      StrCpy $R0 1
    ${EndIf}

    ${If} $R0 == 0
      ${DisableX64FSRedirection}
      nsExec::ExecToStack 'powershell.exe -NoProfile -Command "if (Get-Process GameTrainerHub -ErrorAction SilentlyContinue) { exit 0 } else { exit 1 }"'
      Pop $0
      Pop $1
      ${EnableX64FSRedirection}
      ${If} $0 == 0
        StrCpy $R0 1
      ${EndIf}
    ${EndIf}

    ${If} $R0 == 1
      MessageBox MB_OKCANCEL|MB_ICONEXCLAMATION "检测到 ${PRODUCT_NAME} 当前正在运行。$\r$\n$\r$\n为了确保卸载顺利完成，需要先关闭正在运行的程序。$\r$\n$\r$\n请确认是否关闭程序并继续卸载？$\r$\n$\r$\n【确定】：关闭正在运行的程序，并继续卸载。$\r$\n【取消】：退出卸载程序，不继续卸载。" IDOK un_on_confirm IDCANCEL un_on_cancel

      un_on_confirm:
        FindWindow $0 "wailsWindow" ""
        ${If} $0 != 0
          System::Call 'user32::PostMessage(i $0, i 0x0012, i 0, i 0)'
        ${EndIf}
        FindWindow $0 "GameTrainerHub_TrayWindow" ""
        ${If} $0 != 0
          System::Call 'user32::PostMessage(i $0, i 0x0012, i 0, i 0)'
        ${EndIf}
        ${DisableX64FSRedirection}
        nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$PLUGINSDIR\graceful_close.ps1"'
        nsExec::Exec 'cmd.exe /c taskkill.exe /IM "${APP_EXE}"'
        ${EnableX64FSRedirection}

        StrCpy $R1 0
        un_wait_exit_loop:
          Sleep 500
          IntOp $R1 $R1 + 1

          StrCpy $R0 0
          ${DisableX64FSRedirection}
          nsExec::ExecToStack 'cmd.exe /c tasklist /NH /FI "IMAGENAME eq ${APP_EXE}" | findstr /I "${APP_EXE}"'
          Pop $0
          Pop $1
          ${EnableX64FSRedirection}
          ${If} $0 == 0
            StrCpy $R0 1
          ${EndIf}

          ${If} $R0 == 0
            Goto un_process_exited
          ${EndIf}

          ${If} $R1 < 12
            Goto un_wait_exit_loop
          ${EndIf}

        MessageBox MB_YESNOCANCEL|MB_ICONEXCLAMATION "${PRODUCT_NAME} 未能在规定时间内正常退出。$\r$\n$\r$\n是否尝试强制结束该程序？$\r$\n$\r$\n【是】：强制结束程序并继续卸载（警告：未保存的数据可能会丢失）。$\r$\n【否】：重新尝试等待正常退出。$\r$\n【取消】：退出卸载程序，不执行任何更改。" IDYES un_do_force_kill IDNO un_on_confirm
        Goto un_on_cancel

        un_do_force_kill:
          ${DisableX64FSRedirection}
          nsExec::Exec 'cmd.exe /c taskkill.exe /F /T /IM "${APP_EXE}"'
          nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "Get-Process GameTrainerHub -ErrorAction SilentlyContinue | Stop-Process -Force"'
          ${EnableX64FSRedirection}
          Sleep 1500
          Goto un_check_loop

      un_on_cancel:
        Abort
    ${EndIf}

  un_process_exited:
FunctionEnd

; ------------------------------------------------------------------------------
; Functions - WebView2 Runtime Check
; ------------------------------------------------------------------------------
Function CheckWebView2Runtime
  ${If} ${Silent}
    Return
  ${EndIf}
  ; Check 64-bit machine registry (official Evergreen GUID: {F3017226-FE2A-4295-8BDF-00C3A9A7E4C5})
  ReadRegStr $0 HKLM "SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" "pv"
  ${If} $0 != ""
    Return
  ${EndIf}
  ; Check standard machine registry
  ReadRegStr $0 HKLM "SOFTWARE\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" "pv"
  ${If} $0 != ""
    Return
  ${EndIf}
  ; Check current user registry
  ReadRegStr $0 HKCU "SOFTWARE\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" "pv"
  ${If} $0 != ""
    Return
  ${EndIf}
  ; Check alternative GUID
  ReadRegStr $0 HKLM "SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-94D0-44AC-A0C1-0452F34307F1}" "pv"
  ${If} $0 != ""
    Return
  ${EndIf}

  ; If not found, notify user
  MessageBox MB_YESNO|MB_ICONEXCLAMATION "未检测到系统安装 Microsoft Edge WebView2 Runtime。$\r$\n${PRODUCT_NAME} 依赖此组件以呈现客户端界面。$\r$\n$\r$\n是否立即前往微软官方网站下载 WebView2 常青版引导程序 (Evergreen Bootstrapper)？$\r$\n（选择【否】将继续安装，但首次运行可能提示缺少组件）" IDYES open_wv2 IDNO continue_wv2
  open_wv2:
    ExecShell "open" "https://go.microsoft.com/fwlink/p/?LinkId=2124703"
  continue_wv2:
FunctionEnd

; ------------------------------------------------------------------------------
; Initialization Callbacks
; ------------------------------------------------------------------------------
Function .onInit
  SetRegView 64
  SetShellVarContext all

  ; 1. Check 64-bit Windows
  ${IfNot} ${RunningX64}
    MessageBox MB_OK|MB_ICONSTOP "${PRODUCT_NAME} 仅支持 64 位 Windows 操作系统。"
    Abort
  ${EndIf}

  ; Initialize temporary plugins dir and extract helper
  InitPluginsDir
  File "/oname=$PLUGINSDIR\graceful_close.ps1" "graceful_close.ps1"

  ; 2. Check running process
  Call CheckRunningProcess

  ; 3. Check WebView2 Runtime
  Call CheckWebView2Runtime
FunctionEnd

Function un.onInit
  SetRegView 64
  SetShellVarContext all

  ; Initialize temporary plugins dir and extract helper
  InitPluginsDir
  File "/oname=$PLUGINSDIR\graceful_close.ps1" "graceful_close.ps1"

  ; Check running process during uninstall
  Call un.CheckRunningProcess
FunctionEnd

; ------------------------------------------------------------------------------
; Installer Sections
; ------------------------------------------------------------------------------
Section "主程序核心文件 (必选)" SecCore
  SectionIn RO
  ; Double-check running process before extracting files
  Call CheckRunningProcess

  SetOutPath "$INSTDIR"
  SetOverwrite on

  ; Ensure target exe is not locked before overwriting
  ${If} ${FileExists} "$INSTDIR\${APP_EXE}"
    ClearErrors
    Delete "$INSTDIR\${APP_EXE}"
    ${If} ${Errors}
      ; If delete failed due to file lock, force kill and wait
      ${DisableX64FSRedirection}
      nsExec::Exec 'cmd.exe /c taskkill.exe /F /T /IM "${APP_EXE}"'
      nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "Get-Process GameTrainerHub -ErrorAction SilentlyContinue | Stop-Process -Force"'
      ${EnableX64FSRedirection}
      Sleep 1500
      ClearErrors
      Delete "$INSTDIR\${APP_EXE}"
    ${EndIf}
  ${EndIf}

  ; Install application binary and assets
  File "..\bin\GameTrainerHub.exe"
  File "NOTICE.txt"
  File "version.json"
  File "..\windows\icon.ico"

  ; Create uninstaller
  WriteUninstaller "$INSTDIR\Uninstall.exe"

  ; Register App Paths
  WriteRegStr HKLM "${PRODUCT_DIR_REGKEY}" "" "$INSTDIR\${APP_EXE}"
  WriteRegStr HKLM "${PRODUCT_DIR_REGKEY}" "Path" "$INSTDIR"

  ; Register Control Panel / Add-Remove Programs
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "DisplayName" "${PRODUCT_NAME}"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "UninstallString" '"$INSTDIR\Uninstall.exe"'
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "QuietUninstallString" '"$INSTDIR\Uninstall.exe" /S'
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "DisplayIcon" '"$INSTDIR\${APP_EXE}",0'
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "DisplayVersion" "${PRODUCT_VERSION}"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "URLInfoAbout" "${PRODUCT_WEB_SITE}"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "Publisher" "${PRODUCT_PUBLISHER}"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "InstallLocation" "$INSTDIR"
  WriteRegDWORD ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "NoModify" 1
  WriteRegDWORD ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "NoRepair" 1

  ; Calculate and write EstimatedSize (in KB)
  ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
  IntFmt $0 "0x%08X" $0
  WriteRegDWORD ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "EstimatedSize" "$0"

  ; Create Start Menu Shortcuts
  CreateDirectory "$SMPROGRAMS\${PRODUCT_NAME}"
  CreateShortcut "$SMPROGRAMS\${PRODUCT_NAME}\${PRODUCT_NAME}.lnk" "$INSTDIR\${APP_EXE}" "" "$INSTDIR\${APP_EXE}" 0
  CreateShortcut "$SMPROGRAMS\${PRODUCT_NAME}\卸载 ${PRODUCT_NAME}.lnk" "$INSTDIR\Uninstall.exe" "" "$INSTDIR\Uninstall.exe" 0

  ; Create Desktop Shortcut
  CreateShortcut "$DESKTOP\${PRODUCT_NAME}.lnk" "$INSTDIR\${APP_EXE}" "" "$INSTDIR\${APP_EXE}" 0
SectionEnd

Section "创建桌面快捷方式" SecDesktop
  CreateShortcut "$DESKTOP\${PRODUCT_NAME}.lnk" "$INSTDIR\${APP_EXE}" "" "$INSTDIR\${APP_EXE}" 0
SectionEnd

; ------------------------------------------------------------------------------
; Section Descriptions
; ------------------------------------------------------------------------------
!insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
  !insertmacro MUI_DESCRIPTION_TEXT ${SecCore} "包含 Game Trainer Hub 主程序、核心运行时配置及必要文件。"
  !insertmacro MUI_DESCRIPTION_TEXT ${SecDesktop} "在 Windows 桌面上创建快速启动图标。"
!insertmacro MUI_FUNCTION_DESCRIPTION_END

; ------------------------------------------------------------------------------
; Uninstaller Section
; ------------------------------------------------------------------------------
Section "Uninstall"
  ; Double-check running process before deleting files
  Call un.CheckRunningProcess

  ; 1. Delete installed files
  ${If} ${FileExists} "$INSTDIR\${APP_EXE}"
    ClearErrors
    Delete "$INSTDIR\${APP_EXE}"
    ${If} ${Errors}
      ${DisableX64FSRedirection}
      nsExec::Exec 'cmd.exe /c taskkill.exe /F /T /IM "${APP_EXE}"'
      nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "Get-Process GameTrainerHub -ErrorAction SilentlyContinue | Stop-Process -Force"'
      ${EnableX64FSRedirection}
      Sleep 1500
      ClearErrors
      Delete "$INSTDIR\${APP_EXE}"
    ${EndIf}
  ${EndIf}
  Delete "$INSTDIR\NOTICE.txt"
  Delete "$INSTDIR\version.json"
  Delete "$INSTDIR\icon.ico"
  Delete "$INSTDIR\Uninstall.exe"

  ; 2. Delete Shortcuts
  Delete "$DESKTOP\${PRODUCT_NAME}.lnk"
  Delete "$SMPROGRAMS\${PRODUCT_NAME}\${PRODUCT_NAME}.lnk"
  Delete "$SMPROGRAMS\${PRODUCT_NAME}\卸载 ${PRODUCT_NAME}.lnk"
  RMDir "$SMPROGRAMS\${PRODUCT_NAME}"

  SetShellVarContext current
  Delete "$DESKTOP\${PRODUCT_NAME}.lnk"
  Delete "$SMPROGRAMS\${PRODUCT_NAME}\${PRODUCT_NAME}.lnk"
  Delete "$SMPROGRAMS\${PRODUCT_NAME}\卸载 ${PRODUCT_NAME}.lnk"
  RMDir "$SMPROGRAMS\${PRODUCT_NAME}"
  SetShellVarContext all

  ; 3. Delete Registry entries
  DeleteRegKey HKLM "${PRODUCT_DIR_REGKEY}"
  DeleteRegKey ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}"

  ; 4. Inquire about User Personal Data (%APPDATA%\GameTrainerHub)
  ; Default is KEEP user data. Only delete if explicitly confirmed twice.
  ${IfNot} ${Silent}
    MessageBox MB_YESNO|MB_ICONQUESTION "是否同时删除用户个人数据（包括下载的修改器、数据库及配置）？$\r$\n$\r$\n【否】（推荐）：保留用户数据以备日后再次安装。$\r$\n【是】：彻底删除用户数据。" IDNO keep_user_data
    MessageBox MB_YESNO|MB_ICONEXCLAMATION|MB_DEFBUTTON2 "【二次确认】确定要彻底清除以下目录中的所有个人数据吗？$\r$\n$APPDATA\GameTrainerHub$\r$\n$\r$\n此操作不可撤销！" IDYES purge_user_data IDNO keep_user_data

    purge_user_data:
      RMDir /r "$APPDATA\GameTrainerHub"
      Goto finish_user_data

    keep_user_data:
      ; Do nothing, user data is preserved
    finish_user_data:
  ${EndIf}

  ; 5. Remove Installation Directory
  RMDir "$INSTDIR"
SectionEnd
