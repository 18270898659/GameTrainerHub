# Graceful Close Helper for Game Trainer Hub
param(
    [string]$ProcessName = "GameTrainerHub"
)

$csharp = @"
using System;
using System.Runtime.InteropServices;
public class WinCloseHelper {
    public delegate bool EnumWindowProc(IntPtr hWnd, IntPtr lParam);
    [DllImport("user32.dll")] public static extern bool EnumWindows(EnumWindowProc lpfn, IntPtr lParam);
    [DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(IntPtr hWnd, out uint lpdwProcessId);
    [DllImport("user32.dll")] public static extern bool PostMessage(IntPtr hWnd, uint Msg, IntPtr wParam, IntPtr lParam);
    public const uint WM_CLOSE = 0x0010;
    public const uint WM_QUIT = 0x0012;

    public static void ClosePid(uint targetPid) {
        EnumWindows((hWnd, lParam) => {
            uint pid;
            GetWindowThreadProcessId(hWnd, out pid);
            if (pid == targetPid) {
                PostMessage(hWnd, WM_CLOSE, IntPtr.Zero, IntPtr.Zero);
                PostMessage(hWnd, WM_QUIT, IntPtr.Zero, IntPtr.Zero);
            }
            return true;
        }, IntPtr.Zero);
    }
}
"@

try {
    Add-Type -TypeDefinition $csharp -Language CSharp -ErrorAction SilentlyContinue
    $procs = Get-Process -Name $ProcessName -ErrorAction SilentlyContinue
    foreach ($p in $procs) {
        try {
            [WinCloseHelper]::ClosePid([uint32]$p.Id)
            $p.CloseMainWindow() | Out-Null
        } catch {}
    }
} catch {
    Start-Process -FilePath "taskkill.exe" -ArgumentList "/IM $ProcessName.exe" -WindowStyle Hidden -Wait -ErrorAction SilentlyContinue
}
