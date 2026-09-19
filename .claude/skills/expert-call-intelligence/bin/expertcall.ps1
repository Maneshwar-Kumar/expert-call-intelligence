param([Parameter(ValueFromRemainingArguments=$true)][string[]]$Arguments)
$dir = Split-Path -Parent $MyInvocation.MyCommand.Path
$arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
if ($IsWindows -or $env:OS -eq "Windows_NT") {
    $bin = if ($arch -eq "arm64") { "expertcall-windows-arm64.exe" } else { "expertcall-windows-amd64.exe" }
} elseif ($IsMacOS) {
    $bin = if ($arch -eq "arm64") { "expertcall-darwin-arm64" } else { "expertcall-darwin-amd64" }
} else {
    $bin = if ($arch -eq "arm64") { "expertcall-linux-arm64" } else { "expertcall-linux-amd64" }
}
& (Join-Path $dir $bin) @Arguments
exit $LASTEXITCODE
