param(
    [switch]$SkipInstall,
    [switch]$BackendOnly,
    [switch]$FrontendOnly,
    [switch]$Start,
    [switch]$Stop,
    [switch]$Status
)

$ErrorActionPreference = "Stop"

if ($BackendOnly -and $FrontendOnly) {
    throw "Use only one of -BackendOnly or -FrontendOnly."
}

if ($Start -and $Stop) {
    throw "Use only one of -Start or -Stop."
}

$repoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$backendPath = Join-Path $repoRoot "backend"
$frontendPath = Join-Path $repoRoot "frontend"
$runtimePath = Join-Path $repoRoot ".runtime"
$frontendPidFile = Join-Path $runtimePath "frontend.pid"
$backendPidFile = Join-Path $runtimePath "backend.pid"
$backendEnvPath = Join-Path $backendPath ".env"
$backendEnvExamplePath = Join-Path $backendPath ".env.example"

function Test-Command {
    param([Parameter(Mandatory = $true)][string]$Name)
    return [bool](Get-Command $Name -ErrorAction SilentlyContinue)
}

function Get-TrackedProcess {
    param([Parameter(Mandatory = $true)][string]$PidFile)

    if (-not (Test-Path $PidFile)) {
        return $null
    }

    $pidValue = Get-Content -Raw $PidFile -ErrorAction SilentlyContinue
    if (-not $pidValue) {
        return $null
    }

    $pidNumber = 0
    if (-not [int]::TryParse($pidValue.Trim(), [ref]$pidNumber)) {
        Remove-Item -Path $PidFile -ErrorAction SilentlyContinue
        return $null
    }
    $process = Get-Process -Id $pidNumber -ErrorAction SilentlyContinue
    if (-not $process) {
        Remove-Item -Path $PidFile -ErrorAction SilentlyContinue
    }
    return $process
}

function Save-TrackedProcess {
    param(
        [Parameter(Mandatory = $true)][string]$PidFile,
        [Parameter(Mandatory = $true)][int]$ProcessId
    )

    if (-not (Test-Path $runtimePath)) {
        New-Item -Path $runtimePath -ItemType Directory | Out-Null
    }

    Set-Content -Path $PidFile -Value $ProcessId
}

function Get-ListeningProcessId {
    param([Parameter(Mandatory = $true)][int]$Port)

    try {
        $connection = Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction Stop | Select-Object -First 1
        if ($connection) {
            return [int]$connection.OwningProcess
        }
    } catch {
        return $null
    }

    return $null
}

function Get-ChildProcessIds {
    param([Parameter(Mandatory = $true)][int]$ParentPid)

    try {
        $children = Get-CimInstance Win32_Process -Filter "ParentProcessId = $ParentPid" -ErrorAction Stop
    } catch {
        return @()
    }

    $ids = @()
    foreach ($child in $children) {
        $childId = [int]$child.ProcessId
        $ids += $childId
        $ids += Get-ChildProcessIds -ParentPid $childId
    }

    return $ids
}

function Stop-ProcessTree {
    param([Parameter(Mandatory = $true)][int]$RootPid)

    $allChildIds = Get-ChildProcessIds -ParentPid $RootPid | Select-Object -Unique
    foreach ($childId in ($allChildIds | Sort-Object -Descending)) {
        Stop-Process -Id $childId -Force -ErrorAction SilentlyContinue
    }

    Stop-Process -Id $RootPid -Force -ErrorAction SilentlyContinue
}

function Stop-ProcessListeningOnPort {
    param(
        [Parameter(Mandatory = $true)][int]$Port,
        [Parameter(Mandatory = $true)][string]$Name
    )

    $ownerPid = Get-ListeningProcessId -Port $Port
    if (-not $ownerPid) {
        return
    }

    Stop-ProcessTree -RootPid $ownerPid
    Write-Host "Stopped $Name listener on port $Port (PID $ownerPid)." -ForegroundColor Yellow
}

function Get-ServerAddrFromEnv {
    param([Parameter(Mandatory = $true)][string]$EnvPath)

    if (-not (Test-Path $EnvPath)) {
        return $null
    }

    $line = Get-Content $EnvPath | Where-Object { $_ -match '^SERVER_ADDR=' } | Select-Object -First 1
    if (-not $line) {
        return $null
    }

    return ($line -replace '^SERVER_ADDR=', '').Trim()
}

function Get-BackendPort {
    $serverAddr = Get-ServerAddrFromEnv -EnvPath $backendEnvPath
    if (-not $serverAddr) {
        $serverAddr = Get-ServerAddrFromEnv -EnvPath $backendEnvExamplePath
    }
    if (-not $serverAddr) {
        $serverAddr = ":8080"
    }

    if ($serverAddr -match ':(\d+)$') {
        return [int]$Matches[1]
    }

    return 8080
}

function Ensure-FrontendDependencies {
    if (-not (Test-Command "npm")) {
        throw "npm is not installed or not in PATH."
    }

    $vitePath = Join-Path $frontendPath "node_modules/.bin/vite.cmd"
    $needsInstall = -not (Test-Path $vitePath)
    $shouldInstall = (-not $SkipInstall) -or $needsInstall

    if ($shouldInstall) {
        if ($SkipInstall -and $needsInstall) {
            Write-Host "Frontend dependencies missing; running npm install (ignoring -SkipInstall)." -ForegroundColor Yellow
        } else {
            Write-Host "Installing frontend dependencies..." -ForegroundColor Cyan
        }
        Push-Location $frontendPath
        try {
            npm install
        } finally {
            Pop-Location
        }
    }
}

function Stop-TrackedProcess {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string]$PidFile
    )

    $process = Get-TrackedProcess -PidFile $PidFile
    if ($process) {
        Stop-ProcessTree -RootPid $process.Id
        Write-Host "Stopped $Name (PID $($process.Id))." -ForegroundColor Yellow
    } else {
        Write-Host "$Name is not running." -ForegroundColor DarkYellow
    }

    Remove-Item -Path $PidFile -ErrorAction SilentlyContinue
}

function Start-Frontend {
    Ensure-FrontendDependencies

    Write-Host "Starting frontend (Vite)..." -ForegroundColor Cyan
    $process = Start-Process -FilePath "cmd.exe" -WorkingDirectory $frontendPath -PassThru -ArgumentList "/k npm run dev"
    Save-TrackedProcess -PidFile $frontendPidFile -ProcessId $process.Id
}

function Start-Backend {
    if (-not (Test-Command "go")) {
        throw "Go is not installed or not in PATH."
    }

    $backendPort = Get-BackendPort
    $portOwnerPid = Get-ListeningProcessId -Port $backendPort
    if ($portOwnerPid) {
        $ownerProcess = Get-Process -Id $portOwnerPid -ErrorAction SilentlyContinue
        $ownerName = if ($ownerProcess) { $ownerProcess.ProcessName } else { "unknown" }
        throw "Cannot start backend: port $backendPort is already in use by PID $portOwnerPid ($ownerName). Stop that process first."
    }

    Write-Host "Starting backend (Go server)..." -ForegroundColor Cyan
    $process = Start-Process -FilePath "cmd.exe" -WorkingDirectory $backendPath -PassThru -ArgumentList "/k go run ./cmd/server"
    Save-TrackedProcess -PidFile $backendPidFile -ProcessId $process.Id
}

function Get-Targets {
    if ($BackendOnly) {
        return @("backend")
    }

    if ($FrontendOnly) {
        return @("frontend")
    }

    return @("frontend", "backend")
}

function Show-Status {
    param([Parameter(Mandatory = $true)][string[]]$Targets)

    $backendPort = Get-BackendPort
    $frontendPort = 5173

    foreach ($target in $Targets) {
        if ($target -eq "frontend") {
            $process = Get-TrackedProcess -PidFile $frontendPidFile
            $portOwnerPid = Get-ListeningProcessId -Port $frontendPort
            if ($portOwnerPid) {
                Write-Host "Frontend: running on :$frontendPort (PID $portOwnerPid)" -ForegroundColor Green
            } elseif ($process) {
                Write-Host "Frontend: process exists but port :$frontendPort is closed (PID $($process.Id))" -ForegroundColor DarkYellow
            } else {
                Write-Host "Frontend: stopped" -ForegroundColor DarkYellow
            }
        }

        if ($target -eq "backend") {
            $process = Get-TrackedProcess -PidFile $backendPidFile
            $portOwnerPid = Get-ListeningProcessId -Port $backendPort
            if ($portOwnerPid) {
                Write-Host "Backend: running on :$backendPort (PID $portOwnerPid)" -ForegroundColor Green
            } elseif ($process) {
                Write-Host "Backend: process exists but port :$backendPort is closed (PID $($process.Id))" -ForegroundColor DarkYellow
            } else {
                Write-Host "Backend: stopped" -ForegroundColor DarkYellow
            }
        }
    }
}

function Start-Targets {
    param([Parameter(Mandatory = $true)][string[]]$Targets)

    $backendPort = Get-BackendPort
    $frontendPort = 5173

    foreach ($target in $Targets) {
        if ($target -eq "frontend") {
            $existing = Get-TrackedProcess -PidFile $frontendPidFile
            $portOwnerPid = Get-ListeningProcessId -Port $frontendPort
            if ($portOwnerPid) {
                Write-Host "Frontend port :$frontendPort already in use by PID $portOwnerPid." -ForegroundColor DarkYellow
            } elseif ($existing) {
                Write-Host "Frontend had stale process (PID $($existing.Id)); restarting." -ForegroundColor DarkYellow
                Stop-TrackedProcess -Name "frontend" -PidFile $frontendPidFile
                Start-Frontend
            } else {
                Start-Frontend
            }
        }

        if ($target -eq "backend") {
            $existing = Get-TrackedProcess -PidFile $backendPidFile
            $portOwnerPid = Get-ListeningProcessId -Port $backendPort
            if ($portOwnerPid) {
                $ownerProcess = Get-Process -Id $portOwnerPid -ErrorAction SilentlyContinue
                $ownerName = if ($ownerProcess) { $ownerProcess.ProcessName } else { "unknown" }
                Write-Host "Backend port :$backendPort already in use by PID $portOwnerPid ($ownerName)." -ForegroundColor DarkYellow
            } elseif ($existing) {
                Write-Host "Backend had stale process (PID $($existing.Id)); restarting." -ForegroundColor DarkYellow
                Stop-TrackedProcess -Name "backend" -PidFile $backendPidFile
                Start-Backend
            } else {
                Start-Backend
            }
        }
    }

    Start-Sleep -Seconds 2

    if ($Targets -contains "frontend") {
        Start-Process "http://localhost:5173"
    }

    if ($Targets -contains "backend") {
        Start-Process "http://localhost:$backendPort/health"
    }
}

function Stop-Targets {
    param([Parameter(Mandatory = $true)][string[]]$Targets)

    $backendPort = Get-BackendPort

    foreach ($target in $Targets) {
        if ($target -eq "frontend") {
            Stop-TrackedProcess -Name "frontend" -PidFile $frontendPidFile
            Stop-ProcessListeningOnPort -Port 5173 -Name "frontend"
        }

        if ($target -eq "backend") {
            Stop-TrackedProcess -Name "backend" -PidFile $backendPidFile
            Stop-ProcessListeningOnPort -Port $backendPort -Name "backend"
        }
    }
}

$targets = Get-Targets

if ($Status) {
    Show-Status -Targets $targets
    exit 0
}

if ($Start) {
    Start-Targets -Targets $targets
    Write-Host "Start command completed." -ForegroundColor Green
    exit 0
}

if ($Stop) {
    Stop-Targets -Targets $targets
    Write-Host "Stop command completed." -ForegroundColor Green
    exit 0
}

$hasRunning = $false
if ($targets -contains "frontend") {
    if ((Get-TrackedProcess -PidFile $frontendPidFile) -or (Get-ListeningProcessId -Port 5173)) {
        $hasRunning = $true
    }
}
if ($targets -contains "backend") {
    if (Get-TrackedProcess -PidFile $backendPidFile) {
        $hasRunning = $true
    }
}

if ($hasRunning) {
    Stop-Targets -Targets $targets
    Write-Host "Toggle result: services stopped." -ForegroundColor Green
} else {
    Start-Targets -Targets $targets
    Write-Host "Toggle result: services started." -ForegroundColor Green
}
