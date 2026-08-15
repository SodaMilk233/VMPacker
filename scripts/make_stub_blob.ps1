[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$ElfPath,

    [Parameter(Mandatory = $true)]
    [string]$RawPath,

    [Parameter(Mandatory = $true)]
    [string]$OutputPath,

    [Parameter(Mandatory = $true)]
    [string]$NmPath,

    [Parameter(Mandatory = $true)]
    [string]$ReadElfPath
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Invoke-CheckedTool {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Tool,

        [Parameter(Mandatory = $true)]
        [string[]]$Arguments
    )

    $output = & $Tool @Arguments 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "$Tool failed with exit code $LASTEXITCODE`n$($output -join [Environment]::NewLine)"
    }
    return $output
}

$elfFullPath = (Resolve-Path -LiteralPath $ElfPath).Path
$rawFullPath = (Resolve-Path -LiteralPath $RawPath).Path
$outputFullPath = [IO.Path]::GetFullPath($OutputPath)

$relocations = Invoke-CheckedTool -Tool $ReadElfPath -Arguments @('--relocations', $elfFullPath)
if (($relocations -join "`n") -match 'Relocation section') {
    throw "Linked VM stub still contains relocations:`n$($relocations -join [Environment]::NewLine)"
}

$undefined = @(Invoke-CheckedTool -Tool $NmPath -Arguments @('--undefined-only', $elfFullPath))
$undefinedLines = @($undefined | Where-Object { $_.Trim().Length -gt 0 })
if ($undefinedLines.Count -gt 0) {
    throw "Linked VM stub contains undefined symbols:`n$($undefinedLines -join [Environment]::NewLine)"
}

$symbols = @{}
$nmOutput = Invoke-CheckedTool -Tool $NmPath -Arguments @('--defined-only', '--numeric-sort', $elfFullPath)
foreach ($line in $nmOutput) {
    $parts = $line.Trim() -split '\s+'
    if ($parts.Count -ge 3) {
        $symbols[$parts[-1]] = [Convert]::ToUInt64($parts[0], 16)
    }
}

$requiredSymbols = @('vm_entry', 'vm_entry_token', '_token_table_va')
foreach ($name in $requiredSymbols) {
    if (-not $symbols.ContainsKey($name)) {
        throw "Required symbol '$name' was not found in $elfFullPath"
    }
}

$raw = [IO.File]::ReadAllBytes($rawFullPath)
$blob = [byte[]]::new(24 + $raw.Length)

[Buffer]::BlockCopy([BitConverter]::GetBytes([UInt64]$symbols['vm_entry']), 0, $blob, 0, 8)
[Buffer]::BlockCopy([BitConverter]::GetBytes([UInt64]$symbols['vm_entry_token']), 0, $blob, 8, 8)
[Buffer]::BlockCopy([BitConverter]::GetBytes([UInt64]$symbols['_token_table_va']), 0, $blob, 16, 8)
[Buffer]::BlockCopy($raw, 0, $blob, 24, $raw.Length)

$outputDirectory = [IO.Path]::GetDirectoryName($outputFullPath)
if ($outputDirectory) {
    [IO.Directory]::CreateDirectory($outputDirectory) | Out-Null
}
[IO.File]::WriteAllBytes($outputFullPath, $blob)

Write-Host ('[+] {0}: {1} bytes (vm_entry=0x{2:X}, vm_entry_token=0x{3:X}, _token_table_va=0x{4:X})' -f `
    $OutputPath,
    $blob.Length,
    $symbols['vm_entry'],
    $symbols['vm_entry_token'],
    $symbols['_token_table_va'])
