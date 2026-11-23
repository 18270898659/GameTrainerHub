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
Function CheckRunningProcess
  ${If} ${Silent}
    nsExec::Exec 'cmd /c taskkill /F /T /IM "${APP_EXE}"'
    Sleep 1000
    Return
  ${EndIf}
  loop:
    nsExec::ExecToStack 'cmd /c tasklist /NH /FI "IMAGENAME eq ${APP_EXE}" | findstr /I "${APP_EXE}"'
    Pop $0
    Pop $1
    ${If} $0 == 0
      MessageBox MB_YESNO|MB_ICONQUESTION "检测到 ${PRODUCT_NAME} 正在运行。$\r$\n$\r$\n原有程序已打开，是否关闭？$\r$\n$\r$\n点击【是】将关闭程序并继续安装。$\r$\n点击【否】将退出安装程序。" IDYES kill_and_continue IDNO abort_install
      kill_and_continue:
        nsExec::Exec 'cmd /c taskkill /F /T /IM "${APP_EXE}"'
        Sleep 1000
        Goto loop
      abort_install:
        Abort
    ${EndIf}
FunctionEnd

Function un.CheckRunningProcess
  ${If} ${Silent}
    nsExec::Exec 'cmd /c taskkill /F /T /IM "${APP_EXE}"'
    Sleep 1000
    Return
  ${EndIf}
  un_loop:
    nsExec::ExecToStack 'cmd /c tasklist /NH /FI "IMAGENAME eq ${APP_EXE}" | findstr /I "${APP_EXE}"'
    Pop $0
    Pop $1
    ${If} $0 == 0
      MessageBox MB_YESNO|MB_ICONQUESTION "检测到 ${PRODUCT_NAME} 正在运行。$\r$\n$\r$\n原有程序已打开，是否关闭？$\r$\n$\r$\n点击【是】将关闭程序并继续卸载。$\r$\n点击【否】将退出卸载程序。" IDYES un_kill_and_continue IDNO un_abort_uninstall
      un_kill_and_continue:
        nsExec::Exec 'cmd /c taskkill /F /T /IM "${APP_EXE}"'
        Sleep 1000
        Goto un_loop
      un_abort_uninstall:
        Abort
    ${EndIf}
FunctionEnd

; ------------------------------------------------------------------------------
; Functions - WebView2 Runtime Check
; ------------------------------------------------------------------------------
Function CheckWebView2Runtime
  ${If} ${Silent}
    Return
  ${EndIf}
  ; Check 64-bit machine registry
  ReadRegStr $0 HKLM "SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-94D0-44AC-A0C1-0452F34307F1}" "pv"
  ${If} $0 != ""
    Return
  ${EndIf}
  ; Check standard machine registry
  ReadRegStr $0 HKLM "SOFTWARE\Microsoft\EdgeUpdate\Clients\{F3017226-94D0-44AC-A0C1-0452F34307F1}" "pv"
  ${If} $0 != ""
    Return
  ${EndIf}
  ; Check current user registry
  ReadRegStr $0 HKCU "SOFTWARE\Microsoft\EdgeUpdate\Clients\{F3017226-94D0-44AC-A0C1-0452F34307F1}" "pv"
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

  ; 2. Check running process
  Call CheckRunningProcess

  ; 3. Check WebView2 Runtime
  Call CheckWebView2Runtime
FunctionEnd

Function un.onInit
  SetRegView 64
  SetShellVarContext all

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
  Delete "$INSTDIR\${APP_EXE}"
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
