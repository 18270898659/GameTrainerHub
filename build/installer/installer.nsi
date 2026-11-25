; ==============================================================================
; Game Trainer Hub - NSIS Installation Script
; ==============================================================================

Unicode True
SetCompressor /SOLID lzma

; ------------------------------------------------------------------------------
; Definitions
; ------------------------------------------------------------------------------
!define PRODUCT_NAME "Game Trainer Hub"
!define PRODUCT_VERSION "1.0.1"
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
!define MUI_PAGE_CUSTOMFUNCTION_LEAVE DirectoryPageLeave
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES

!define MUI_FINISHPAGE_RUN "$INSTDIR\${APP_EXE}"
!define MUI_FINISHPAGE_RUN_TEXT "运行 ${PRODUCT_NAME}"
!insertmacro MUI_PAGE_FINISH

; ------------------------------------------------------------------------------
; Uninstaller Pages
; ------------------------------------------------------------------------------
!insertmacro MUI_UNPAGE_WELCOME
!define MUI_PAGE_CUSTOMFUNCTION_LEAVE un.ConfirmPageLeave
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

; ------------------------------------------------------------------------------
; Language
; ------------------------------------------------------------------------------
!insertmacro MUI_LANGUAGE "SimpChinese"

; ------------------------------------------------------------------------------
; Functions - Process Detection & Path Normalization (Native Win32, Fail-Safe)
; ------------------------------------------------------------------------------
!macro DEFINE_DETECT_TARGET_PROCESS UN
Function ${UN}NormalizePath
  ; Input: $0 (Path to normalize)
  ; Output: $0 (Normalized Canonical Long Path without \\?\ or trailing slashes)
  Push $1
  Push $2

  ; 1. Strip \\?\ prefix
  StrCpy $1 $0 4
  ${If} $1 == "\\?\"
    StrCpy $0 $0 "" 4
  ${EndIf}

  ; 2. Convert to Full Path (resolves .. and .)
  System::Call "kernel32::GetFullPathNameW(w r0, i 1024, w .r1, i 0) i .r2"
  ${If} $2 > 0
    StrCpy $0 $1
  ${EndIf}

  ; 3. Convert to Long Path (resolves short 8.3 names like PROGRA~1)
  System::Call "kernel32::GetLongPathNameW(w r0, w .r1, i 1024) i .r2"
  ${If} $2 > 0
    StrCpy $0 $1
  ${EndIf}

  ; 4. Strip trailing slashes
  loop_trim_${UN}:
    StrCpy $1 $0 1 -1
    ${If} $1 == "\"
    ${OrIf} $1 == "/"
      StrCpy $0 $0 -1
      Goto loop_trim_${UN}
    ${EndIf}

  Pop $2
  Pop $1
FunctionEnd

; Inputs:
; $R0 = TargetDir ($INSTDIR)
; $R8 = AltDir (Registry InstallLocation, can be empty)
; $R5 = AppPath (Registry App Paths, can be empty)
; Outputs:
; $0 = Status:
;      0: STATE_NOT_RUNNING (No target process is running)
;      1: STATE_RUNNING (Target process is currently running)
;      2: STATE_DETECTION_ERROR (Failed to access process, snapshot failed, or error)
; $1 = Details (Matched Process Path and PID, or Error Description)
Function ${UN}DetectTargetProcess
  Push $2
  Push $3
  Push $4
  Push $5
  Push $6
  Push $7
  Push $8
  Push $9
  Push $R1
  Push $R2
  Push $R3
  Push $R4

  StrCpy $0 0
  StrCpy $1 ""

  ; 1. Prepare normalized target dir and target exe
  StrCpy $0 $R0
  Call ${UN}NormalizePath
  StrCpy $R1 $0                       ; NormTargetDir
  StrCpy $R2 "$R1\${APP_EXE}"         ; NormTargetExe

  ; Norm AltDir if non-empty
  StrCpy $R3 ""
  ${If} $R8 != ""
    StrCpy $0 $R8
    Call ${UN}NormalizePath
    StrCpy $R3 $0
  ${EndIf}

  ; Norm AppPath if non-empty
  StrCpy $R4 ""
  ${If} $R5 != ""
    StrCpy $0 $R5
    Call ${UN}NormalizePath
    StrCpy $R4 $0
  ${EndIf}

  ; 2. Create snapshot of all running processes (TH32CS_SNAPPROCESS = 0x2)
  System::Call "kernel32::CreateToolhelp32Snapshot(i 0x2, i 0) i .r1"
  ${If} $1 == -1
    System::Call "kernel32::GetLastError() i .r2"
    StrCpy $0 2
    StrCpy $1 "无法创建系统进程快照 (CreateToolhelp32Snapshot failed, ErrorCode: $2)"
    Goto exit_func_${UN}
  ${EndIf}

  ; PROCESSENTRY32W struct: 556 bytes
  System::Alloc 556
  Pop $2
  System::Call "*$2(i 556)"

  System::Call "kernel32::Process32FirstW(i r1, i r2) i .r3"

  loop_scan_${UN}:
    ${If} $3 == 0
      Goto scan_done_${UN}
    ${EndIf}

    ; Read th32ProcessID at offset 8, szExeFile at offset 36
    System::Call "*$2(i, i, i .r5, i, i, i, i, i, i, &w260 .r6)"

    ${If} $6 == "${APP_EXE}"
      ; Open process with PROCESS_QUERY_LIMITED_INFORMATION (0x1000)
      System::Call "kernel32::OpenProcess(i 0x1000, i 0, i r5) i .r7"
      ${If} $7 == 0
        ; Permission denied or access restricted -> Fail-Safe DETECTION_ERROR
        System::Call "kernel32::GetLastError() i .r8"
        System::Free $2
        System::Call "kernel32::CloseHandle(i r1)"
        StrCpy $0 2
        StrCpy $1 "检测到 ${APP_EXE} (PID: $5)，但无法访问其进程信息 (权限受限, ErrorCode: $8)"
        Goto exit_func_${UN}
      ${EndIf}

      ; Query full image name across bitness (Vista+)
      System::Call "*(i 1024) i .r4"
      System::Call "kernel32::QueryFullProcessImageNameW(i r7, i 0, w .r8, i r4) i .r0"
      System::Call "kernel32::CloseHandle(i r7)"
      System::Free $4

      ${If} $0 == 0
        ; Query path failed -> Fail-Safe DETECTION_ERROR
        System::Call "kernel32::GetLastError() i .r4"
        System::Free $2
        System::Call "kernel32::CloseHandle(i r1)"
        StrCpy $0 2
        StrCpy $1 "检测到 ${APP_EXE} (PID: $5)，但无法读取其可执行文件路径 (ErrorCode: $4)"
        Goto exit_func_${UN}
      ${EndIf}

      ; Normalize process path
      StrCpy $0 $8
      Call ${UN}NormalizePath
      StrCpy $8 $0 ; $8 = NormProcPath

      ; Path Match Checks:
      ; Match A: Exact match with target executable ($INSTDIR\GameTrainerHub.exe)
      ${If} $8 == $R2
        System::Free $2
        System::Call "kernel32::CloseHandle(i r1)"
        StrCpy $0 1
        StrCpy $1 "$8 (PID: $5)"
        Goto exit_func_${UN}
      ${EndIf}

      ; Match B: Process executable is located inside TargetDir ($INSTDIR)
      StrLen $9 $R1
      StrCpy $7 $8 $9
      ${If} $7 == $R1
        System::Free $2
        System::Call "kernel32::CloseHandle(i r1)"
        StrCpy $0 1
        StrCpy $1 "$8 (PID: $5)"
        Goto exit_func_${UN}
      ${EndIf}

      ; Match C: Match with previously installed directory (AltDir)
      ${If} $R3 != ""
        ${If} $8 == "$R3\${APP_EXE}"
          System::Free $2
          System::Call "kernel32::CloseHandle(i r1)"
          StrCpy $0 1
          StrCpy $1 "$8 (PID: $5)"
          Goto exit_func_${UN}
        ${EndIf}
        StrLen $9 $R3
        StrCpy $7 $8 $9
        ${If} $7 == $R3
          System::Free $2
          System::Call "kernel32::CloseHandle(i r1)"
          StrCpy $0 1
          StrCpy $1 "$8 (PID: $5)"
          Goto exit_func_${UN}
        ${EndIf}
      ${EndIf}

      ; Match D: Match with registered AppPath
      ${If} $R4 != ""
        ${If} $8 == $R4
          System::Free $2
          System::Call "kernel32::CloseHandle(i r1)"
          StrCpy $0 1
          StrCpy $1 "$8 (PID: $5)"
          Goto exit_func_${UN}
        ${EndIf}
      ${EndIf}

    ${EndIf}

    System::Call "kernel32::Process32NextW(i r1, i r2) i .r3"
    Goto loop_scan_${UN}

  scan_done_${UN}:
  System::Free $2
  System::Call "kernel32::CloseHandle(i r1)"
  StrCpy $0 0
  StrCpy $1 "无运行中的目标进程"

  exit_func_${UN}:
  Pop $R4
  Pop $R3
  Pop $R2
  Pop $R1
  Pop $9
  Pop $8
  Pop $7
  Pop $6
  Pop $5
  Pop $4
  Pop $3
  Pop $2
FunctionEnd
!macroend

!insertmacro DEFINE_DETECT_TARGET_PROCESS ""
!insertmacro DEFINE_DETECT_TARGET_PROCESS "un."

; ------------------------------------------------------------------------------
; Functions - Installation Confirmation & Process Check (Triggered on 'Install' click)
; ------------------------------------------------------------------------------
Function DirectoryPageLeave
  ; ----------------------------------------------------------------------------
  ; Step 1: Detect whether GameTrainerHub is already installed (Registry Check)
  ; ----------------------------------------------------------------------------
  SetRegView 64
  StrCpy $R9 0  ; $R9 = 1 if existing installation detected
  ReadRegStr $R8 ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "InstallLocation"
  ReadRegStr $R7 ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "DisplayVersion"
  ReadRegStr $R6 ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "UninstallString"
  ReadRegStr $R5 HKLM "${PRODUCT_DIR_REGKEY}" ""

  ${If} $R8 != ""
  ${OrIf} $R7 != ""
  ${OrIf} $R6 != ""
  ${OrIf} $R5 != ""
    StrCpy $R9 1
    DetailPrint "检测到已安装记录: 版本=$R7, 路径=$R8, 当前目标路径=$INSTDIR"
  ${Else}
    DetailPrint "未检测到系统安装记录，执行全新安装。"
  ${EndIf}

  ; ----------------------------------------------------------------------------
  ; Step 2: Detect whether target GameTrainerHub process is currently running
  ; ----------------------------------------------------------------------------
  check_loop:
    StrCpy $R0 "$INSTDIR"
    Call DetectTargetProcess

    ; --------------------------------------------------------------------------
    ; Safe Failure Case 1: DETECTION_ERROR (State 2) -> MUST NOT PROCEED!
    ; --------------------------------------------------------------------------
    ${If} $0 == 2
      DetailPrint "进程检测异常: $1"
      System::Call 'user32::MessageBoxW(i $HWNDPARENT, w "检测目标程序运行状态时遇到异常，无法确认 GameTrainerHub 是否正在运行。$\r$\n$\r$\n异常原因:$\r$\n$1$\r$\n$\r$\n为保护已有文件及避免覆盖损坏，安装已被阻止。$\r$\n$\r$\n【重试】：再次尝试检测程序状态。$\r$\n【取消】：退出安装向导。", w "进程状态检测异常", i 0x35) i .r0'
      ${If} $0 == 4 ; IDRETRY
        Goto check_loop
      ${Else}
        DetailPrint "用户取消安装：由于进程检测异常，安装已终止。"
        Quit
      ${EndIf}
    ${EndIf}

    ; --------------------------------------------------------------------------
    ; Safe Failure Case 2: NOT_RUNNING (State 0) -> Safe to proceed!
    ; --------------------------------------------------------------------------
    ${If} $0 == 0
      Goto proceed_with_install
    ${EndIf}

    ; --------------------------------------------------------------------------
    ; Safe Failure Case 3: RUNNING (State 1) -> Prompt user with Title & Content
    ; Title: 程序正在运行
    ; Content: 检测到 GameTrainerHub 正在运行，是否关闭程序并继续安装？
    ; Buttons: 是 (IDYES = 6), 否 (IDNO = 7)
    ; --------------------------------------------------------------------------
    DetailPrint "检测到运行中的目标进程: $1"
    System::Call 'user32::MessageBoxW(i $HWNDPARENT, w "检测到 GameTrainerHub 正在运行，是否关闭程序并继续安装？", w "程序正在运行", i 0x24) i .r0'

    ${If} $0 != 6
      ; User clicked "NO" (7) -> Immediately terminate installer
      DetailPrint "用户取消安装：程序正在运行，已终止安装向导。"
      Quit
    ${EndIf}

    ; --------------------------------------------------------------------------
    ; Step 4: User clicked "YES" (6) -> Graceful Shutdown & Wait
    ; --------------------------------------------------------------------------
    ; 4.1 Send graceful exit request to target instances
    FindWindow $0 "wailsWindow" ""
    ${If} $0 != 0
      System::Call 'user32::PostMessageW(i r0, i 0x0010, i 0, i 0)'
      System::Call 'user32::PostMessageW(i r0, i 0x0012, i 0, i 0)'
    ${EndIf}
    FindWindow $0 "GameTrainerHub_TrayWindow" ""
    ${If} $0 != 0
      System::Call 'user32::PostMessageW(i r0, i 0x0010, i 0, i 0)'
      System::Call 'user32::PostMessageW(i r0, i 0x0012, i 0, i 0)'
    ${EndIf}

    ; Trigger graceful shutdown via helper
    ${DisableX64FSRedirection}
    nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$PLUGINSDIR\graceful_close.ps1" -Action close -TargetDir "$INSTDIR" -AltDir "$R8" -AppPath "$R5"'
    ${EnableX64FSRedirection}

    ; 4.2 Wait for all target processes to completely exit (up to 6 seconds)
    StrCpy $R1 0
    wait_exit_loop:
      Sleep 500
      IntOp $R1 $R1 + 1

      StrCpy $R0 "$INSTDIR"
      Call DetectTargetProcess

      ${If} $0 == 0
        ; All target processes exited cleanly!
        DetailPrint "目标程序已完全退出，准备进入安装流程。"
        Goto proceed_with_install
      ${EndIf}

      ${If} $R1 < 12
        Goto wait_exit_loop
      ${EndIf}

    ; 4.3 Timeout reached: Normal shutdown failed -> Prompt user Retry or Cancel
    System::Call 'user32::MessageBoxW(i $HWNDPARENT, w "GameTrainerHub 未能在规定时间内正常退出。$\r$\n$\r$\n是否重新尝试等待程序退出？$\r$\n$\r$\n【重试】：再次尝试关闭并等待退出。$\r$\n【取消】：退出安装程序，不执行任何安装操作。", w "程序关闭超时", i 0x35) i .r0'

    ${If} $0 == 4
      ; User clicked RETRY (4)
      Goto check_loop
    ${Else}
      ; User clicked CANCEL (2) or closed dialog -> Terminate
      Quit
    ${EndIf}

  proceed_with_install:
FunctionEnd

Function CheckRunningProcess
  ${If} ${Silent}
    ; In silent mode (/S): graceful close with path validation
    FindWindow $0 "wailsWindow" ""
    ${If} $0 != 0
      System::Call 'user32::PostMessageW(i r0, i 0x0010, i 0, i 0)'
      System::Call 'user32::PostMessageW(i r0, i 0x0012, i 0, i 0)'
    ${EndIf}
    FindWindow $0 "GameTrainerHub_TrayWindow" ""
    ${If} $0 != 0
      System::Call 'user32::PostMessageW(i r0, i 0x0010, i 0, i 0)'
      System::Call 'user32::PostMessageW(i r0, i 0x0012, i 0, i 0)'
    ${EndIf}
    ${DisableX64FSRedirection}
    nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$PLUGINSDIR\graceful_close.ps1" -Action close -TargetDir "$INSTDIR" -AltDir "$R8" -AppPath "$R5"'
    ${EnableX64FSRedirection}
    Sleep 1500

    StrCpy $R0 "$INSTDIR"
    Call DetectTargetProcess
    ${If} $0 != 0
      ${DisableX64FSRedirection}
      nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$PLUGINSDIR\graceful_close.ps1" -Action force_kill -TargetDir "$INSTDIR" -AltDir "$R8" -AppPath "$R5"'
      ${EnableX64FSRedirection}
      Sleep 500
    ${EndIf}
  ${EndIf}
FunctionEnd

Function un.ConfirmPageLeave
  un_check_loop:
    StrCpy $R0 "$INSTDIR"
    StrCpy $R8 ""
    StrCpy $R5 ""
    Call un.DetectTargetProcess

    ${If} $0 == 2
      System::Call 'user32::MessageBoxW(i $HWNDPARENT, w "检测目标程序运行状态时遇到异常，无法确认 GameTrainerHub 是否正在运行。$\r$\n$\r$\n异常原因:$\r$\n$1$\r$\n$\r$\n为保护已有文件，卸载已被阻止。$\r$\n$\r$\n【重试】：再次尝试检测程序状态。$\r$\n【取消】：退出卸载程序。", w "进程状态检测异常", i 0x35) i .r0'
      ${If} $0 == 4
        Goto un_check_loop
      ${Else}
        Quit
      ${EndIf}
    ${EndIf}

    ${If} $0 == 0
      Goto un_proceed_with_uninstall
    ${EndIf}

    System::Call 'user32::MessageBoxW(i $HWNDPARENT, w "检测到 GameTrainerHub 正在运行，是否关闭程序并继续卸载？", w "程序正在运行", i 0x24) i .r0'

    ${If} $0 != 6
      Quit
    ${EndIf}

    FindWindow $0 "wailsWindow" ""
    ${If} $0 != 0
      System::Call 'user32::PostMessageW(i r0, i 0x0010, i 0, i 0)'
      System::Call 'user32::PostMessageW(i r0, i 0x0012, i 0, i 0)'
    ${EndIf}
    FindWindow $0 "GameTrainerHub_TrayWindow" ""
    ${If} $0 != 0
      System::Call 'user32::PostMessageW(i r0, i 0x0010, i 0, i 0)'
      System::Call 'user32::PostMessageW(i r0, i 0x0012, i 0, i 0)'
    ${EndIf}
    ${DisableX64FSRedirection}
    nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$PLUGINSDIR\graceful_close.ps1" -Action close -TargetDir "$INSTDIR"'
    ${EnableX64FSRedirection}

    StrCpy $R1 0
    un_wait_exit_loop:
      Sleep 500
      IntOp $R1 $R1 + 1

      StrCpy $R0 "$INSTDIR"
      StrCpy $R8 ""
      StrCpy $R5 ""
      Call un.DetectTargetProcess

      ${If} $0 == 0
        Goto un_proceed_with_uninstall
      ${EndIf}

      ${If} $R1 < 12
        Goto un_wait_exit_loop
      ${EndIf}

    System::Call 'user32::MessageBoxW(i $HWNDPARENT, w "GameTrainerHub 未能在规定时间内正常退出。$\r$\n$\r$\n是否重新尝试等待程序退出？$\r$\n$\r$\n【重试】：再次尝试关闭并等待退出。$\r$\n【取消】：退出卸载程序，不执行任何更改。", w "程序关闭超时", i 0x35) i .r0'

    ${If} $0 == 4
      Goto un_check_loop
    ${Else}
      Quit
    ${EndIf}

  un_proceed_with_uninstall:
FunctionEnd

Function un.CheckRunningProcess
  ${If} ${Silent}
    FindWindow $0 "wailsWindow" ""
    ${If} $0 != 0
      System::Call 'user32::PostMessageW(i r0, i 0x0012, i 0, i 0)'
    ${EndIf}
    FindWindow $0 "GameTrainerHub_TrayWindow" ""
    ${If} $0 != 0
      System::Call 'user32::PostMessageW(i r0, i 0x0012, i 0, i 0)'
    ${EndIf}
    ${DisableX64FSRedirection}
    nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$PLUGINSDIR\graceful_close.ps1" -Action close -TargetDir "$INSTDIR"'
    ${EnableX64FSRedirection}
    Sleep 1500

    StrCpy $R0 "$INSTDIR"
    StrCpy $R8 ""
    StrCpy $R5 ""
    Call un.DetectTargetProcess
    ${If} $0 != 0
      ${DisableX64FSRedirection}
      nsExec::Exec 'powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$PLUGINSDIR\graceful_close.ps1" -Action force_kill -TargetDir "$INSTDIR"'
      ${EnableX64FSRedirection}
      Sleep 500
    ${EndIf}
  ${EndIf}
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

  ; In silent mode (/S): ensure running processes are closed automatically without GUI
  ${If} ${Silent}
    Call CheckRunningProcess
  ${EndIf}

  ; 3. Check WebView2 Runtime
  Call CheckWebView2Runtime
FunctionEnd

Function un.onInit
  SetRegView 64
  SetShellVarContext all

  ; Initialize temporary plugins dir and extract helper
  InitPluginsDir
  File "/oname=$PLUGINSDIR\graceful_close.ps1" "graceful_close.ps1"

  ; In silent uninstall (/S): ensure running processes are closed automatically without GUI
  ${If} ${Silent}
    Call un.CheckRunningProcess
  ${EndIf}
FunctionEnd

; ------------------------------------------------------------------------------
; Installer Sections
; ------------------------------------------------------------------------------
Section "主程序核心文件 (必选)" SecCore
  SectionIn RO

  ; In silent mode (/S): ensure running process is closed before file extraction
  ${If} ${Silent}
    Call CheckRunningProcess
  ${EndIf}

  ; Pre-install double check: ensure target process did not restart
  StrCpy $R0 "$INSTDIR"
  Call DetectTargetProcess

  ${If} $0 == 1
    MessageBox MB_OK|MB_ICONSTOP "检测到 GameTrainerHub 进程仍在运行或已被重新启动，无法安全覆盖程序文件。$\r$\n$\r$\n请先关闭正在运行的程序，然后重新运行安装程序。"
    Abort
  ${EndIf}

  ${If} $0 == 2
    MessageBox MB_OK|MB_ICONSTOP "进程状态二次校验异常，无法确认主程序是否已退出。$\r$\n$\r$\n异常原因:$\r$\n$1$\r$\n$\r$\n为保护已有文件，安装已中止。"
    Abort
  ${EndIf}

  SetOutPath "$INSTDIR"
  SetOverwrite on

  ; Ensure target exe is not locked before overwriting
  ${If} ${FileExists} "$INSTDIR\${APP_EXE}"
    ClearErrors
    Delete "$INSTDIR\${APP_EXE}"
    ${If} ${Errors}
      MessageBox MB_OK|MB_ICONSTOP "无法写入 $INSTDIR\${APP_EXE}，文件可能被其他进程锁定。$\r$\n$\r$\n请先关闭正在运行的 GameTrainerHub，然后重新运行安装程序。"
      Abort
    ${EndIf}
  ${EndIf}
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
  ; In silent uninstall (/S): ensure running process is closed before deleting files
  ${If} ${Silent}
    Call un.CheckRunningProcess
  ${EndIf}

  ; Pre-uninstall double check: ensure target process did not restart
  ${DisableX64FSRedirection}
  nsExec::ExecToStack 'powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$PLUGINSDIR\graceful_close.ps1" -Action check -TargetDir "$INSTDIR"'
  Pop $0
  Pop $1
  ${EnableX64FSRedirection}

  ${If} $0 == 0
    MessageBox MB_OK|MB_ICONSTOP "检测到 GameTrainerHub 进程正在运行，无法安全卸载。$\r$\n$\r$\n请先关闭正在运行的程序，然后重新运行卸载程序。"
    Abort
  ${EndIf}

  ; 1. Delete installed files
  ${If} ${FileExists} "$INSTDIR\${APP_EXE}"
    ClearErrors
    Delete "$INSTDIR\${APP_EXE}"
    ${If} ${Errors}
      MessageBox MB_OK|MB_ICONSTOP "无法删除 $INSTDIR\${APP_EXE}，文件可能被其他进程锁定。$\r$\n$\r$\n请先关闭正在运行的 GameTrainerHub，然后重新运行卸载程序。"
      Abort
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
