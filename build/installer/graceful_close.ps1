# Graceful Close & Path-Verified Process Helper for Game Trainer Hub
# Compatible with Windows 10 and Windows 11 (including Smart App Control & ConstrainedLanguage)
param(
    [string]$ProcessName = "GameTrainerHub",
    [string]$TargetDir = "",
    [string]$AltDir = "",
    [string]$AppPath = "",
    [string]$Action = "close" # "check", "close", "force_kill"
)

# Fail-Safe Return Codes:
# 0: Process confirmed RUNNING
# 1: Process confirmed NOT_RUNNING
# 2: DETECTION_ERROR (Permission denied, path unavailable, or unhandled exception)

function Normalize-PathSafe([string]$p) {
    if ([string]::IsNullOrWhiteSpace($p)) { return "" }
    try {
        $clean = $p.Trim().Trim('"', "'")
        if ([string]::IsNullOrWhiteSpace($clean)) { return "" }
        if ($clean.StartsWith('\\?\')) { $clean = $clean.Substring(4) }
        $clean = [System.IO.Path]::GetFullPath($clean)
        return $clean.TrimEnd('\', '/').ToLowerInvariant()
    }
    catch {
        return $p.Trim().Trim('"', "'").TrimEnd('\', '/').ToLowerInvariant()
    }
}

$normTargetDir = Normalize-PathSafe $TargetDir
$normAltDir = Normalize-PathSafe $AltDir
$normAppPath = Normalize-PathSafe $AppPath

function Get-ProcessPathSafe($p) {
    # Method 1: MainModule.FileName
    try {
        if ($p.MainModule -and $p.MainModule.FileName) {
            return $p.MainModule.FileName
        }
    }
    catch {}

    # Method 2: Path property
    try {
        if ($p.Path) { return $p.Path }
    }
    catch {}

    # Method 3: CIM / WMI Win32_Process
    try {
        $cim = Get-CimInstance Win32_Process -Filter "ProcessId = $($p.Id)" -ErrorAction SilentlyContinue
        if ($cim -and $cim.ExecutablePath) {
            return $cim.ExecutablePath
        }
    }
    catch {}

    return ""
}

# Returns:
#  1 = Matching target process
#  0 = Non-matching process
# -1 = Path unavailable / permission restricted (Detection Error)
function Is-MatchingProcess($p) {
    if ([string]::IsNullOrEmpty($normTargetDir) -and [string]::IsNullOrEmpty($normAltDir) -and [string]::IsNullOrEmpty($normAppPath)) {
        return 1
    }

    $rawPath = Get-ProcessPathSafe $p
    if ([string]::IsNullOrWhiteSpace($rawPath)) {
        return -1
    }

    $normProcPath = Normalize-PathSafe $rawPath
    $procDir = [System.IO.Path]::GetDirectoryName($normProcPath).ToLowerInvariant()

    if (![string]::IsNullOrEmpty($normAppPath)) {
        if ($normProcPath -eq $normAppPath) {
            return 1
        }
    }

    if (![string]::IsNullOrEmpty($normTargetDir)) {
        if ($procDir -eq $normTargetDir -or $normProcPath.StartsWith($normTargetDir + "\") -or $normProcPath -eq ($normTargetDir + "\gametrainerhub.exe")) {
            return 1
        }
    }

    if (![string]::IsNullOrEmpty($normAltDir)) {
        if ($procDir -eq $normAltDir -or $normProcPath.StartsWith($normAltDir + "\") -or $normProcPath -eq ($normAltDir + "\gametrainerhub.exe")) {
            return 1
        }
    }

    return 0
}

try {
    $procs = @(Get-Process -Name $ProcessName -ErrorAction SilentlyContinue)
    if ($procs.Count -eq 0) {
        if ($Action -eq "check") { exit 1 } # NOT_RUNNING
        exit 0
    }

    $matchingProcs = @()
    $hasDetectionError = $false

    foreach ($p in $procs) {
        $match = Is-MatchingProcess $p
        if ($match -eq -1) {
            $hasDetectionError = $true
        }
        elseif ($match -eq 1) {
            $matchingProcs += $p
        }
    }

    if ($Action -eq "check") {
        if ($hasDetectionError) {
            exit 2 # DETECTION_ERROR
        }
        if ($matchingProcs.Count -gt 0) {
            exit 0 # RUNNING
        }
        exit 1 # NOT_RUNNING
    }
    elseif ($Action -eq "close") {
        foreach ($p in $matchingProcs) {
            try { $p.CloseMainWindow() | Out-Null } catch {}
        }
        exit 0
    }
    elseif ($Action -eq "force_kill") {
        foreach ($p in $matchingProcs) {
            try { Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue } catch {}
        }
        exit 0
    }
}
catch {
    exit 2 # DETECTION_ERROR
}
