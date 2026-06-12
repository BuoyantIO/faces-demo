# ---------------------------------------------------------------------------
# Faces Demo Windows Manager
# Created with love by Buoyant Inc. (https://buoyant.io) for the Faces Demo project.
# Brought to you by Linkerd users like you!
# ---------------------------------------------------------------------------
#
#
#Requires -RunAsAdministrator
<#
.SYNOPSIS
    Faces Demo Windows Manager - install, build, start, stop, status, logs.

.DESCRIPTION
    Single entry point for all Faces Demo lifecycle operations on Windows.
    Manages all four components from the faces-demo repo as Windows Scheduled
    Tasks that run as SYSTEM and auto-restart on failure:

        smiley  - HTTP backend returning an emoji          (default port 8080)
        color   - gRPC backend returning a hex colour      (default port 8081)
        face    - HTTP backend calling smiley + color      (default port 8082)
        gui     - HTTP front-end served to the browser     (default port 8083)

    Each component gets:
        * A Windows Scheduled Task (SYSTEM, restarts every minute on failure)
        * A config file at  <AppRoot>\<component>\env\<component>.env   (default: C:\faces-demo)
        * Log files at      <LogRoot>\<component>\                         (default: C:\temp\faces-demo)
        * Windows Firewall rules for its HTTP and metrics ports

.PARAMETER Command
    Action to perform. One of:
        Install    - install prereqs, build, deploy, and start components
        Build      - (re)build binaries from source only
        Configure  - create dirs, env files, tasks, firewall rules (no build, no start)
        Repair     - add any missing keys to existing env files (safe, never overwrites)
        Start      - start one or all component tasks
        Stop       - stop one or all component tasks
        Restart    - stop then start
        Status     - show task state and recent log lines
        Logs       - tail log output for a component
        Uninstall  - remove tasks, firewall rules, and optionally files
        Help       - show this help text

.PARAMETER Component
    Which component to target: smiley | color | face | gui | all
    Default: all

.PARAMETER Port
    Override the listen port for the targeted component.

.PARAMETER LogLines
    Number of log lines to show with Logs or Status. Default: 50

.PARAMETER RepoUrl
    Git URL of the faces-demo repo.
    Default: https://github.com/BuoyantIO/faces-demo.git

.PARAMETER SourceRoot
    Directory where the repo is cloned. Default: C:\src

.PARAMETER AppRoot
    Parent directory for installed components. Default: C:\faces-demo

.PARAMETER LogRoot
    Parent directory for log files. Default: C:\temp\faces-demo

.PARAMETER SkipPrereqInstall
    Skip automatic download/install of Git and Go.

.PARAMETER SkipBuild
    Skip the Go build step during Install.

.PARAMETER ResetConfig
    Delete existing env files before writing defaults. Use this to recover from a
    broken or stale configuration. Combine with Configure or Install.
    WARNING: permanently discards any customisations in the env files.

.PARAMETER Force
    For Uninstall: also delete runtime and log directories.

.PARAMETER UpdateRepo
    Pass this switch to make Build or Install do a git fetch+merge before
    building. Without it the existing local clone is used as-is, which is
    faster and works without network access.

.EXAMPLE
    # Full install of all components
    .\Faces-Demo-Windows-Manager.ps1 -Command Install

.EXAMPLE
    # Install only the smiley component on port 80
    .\Faces-Demo-Windows-Manager.ps1 -Command Install -Component smiley -Port 80

.EXAMPLE
    # Check status of everything
    .\Faces-Demo-Windows-Manager.ps1 -Command Status

.EXAMPLE
    # Tail last 100 lines of gui logs
    .\Faces-Demo-Windows-Manager.ps1 -Command Logs -Component gui -LogLines 100

.EXAMPLE
    # Restart just the face service after editing its config
    .\Faces-Demo-Windows-Manager.ps1 -Command Restart -Component face

.EXAMPLE
    # Remove everything including files
    .\Faces-Demo-Windows-Manager.ps1 -Command Uninstall -Force

.EXAMPLE
    # Build, then configure without starting - edit configs, then start individually
    .\Faces-Demo-Windows-Manager.ps1 -Command Build
    .\Faces-Demo-Windows-Manager.ps1 -Command Configure
    # edit <AppRoot>\smiley\env\smiley.env ...
    .\Faces-Demo-Windows-Manager.ps1 -Command Start -Component smiley

.NOTES
    Run from an elevated (Run as Administrator) PowerShell 5.1+ session.
    Set-ExecutionPolicy Bypass -Scope Process -Force
    .\Faces-Demo-Windows-Manager.ps1 -Command Install
#>

param(
    [Parameter(Position = 0)]
    [ValidateSet("Install","Build","Configure","Repair","Start","Stop","Restart","Status","Logs","Uninstall","Help")]
    [string]$Command = "Help",

    [ValidateSet("smiley","color","face","gui","all")]
    [string]$Component = "all",

    [int]$Port = 0,
    [int]$LogLines = 50,

    [string]$RepoUrl    = "https://github.com/BuoyantIO/faces-demo.git",
    [string]$SourceRoot = "C:\src",
    [string]$AppRoot    = "C:\faces-demo",
    [string]$LogRoot    = "C:\temp\faces-demo",

    [switch]$SkipPrereqInstall,
    [switch]$SkipBuild,
    [switch]$Force,

    # Wipe existing env files and regenerate defaults. Use after a broken install.
    # Combine with -Command Configure or -Command Install.
    [switch]$ResetConfig,

    # Pass -UpdateRepo to make Build/Install do a git fetch before building.
    # Omit it (the default) to build from whatever is already cloned locally.
    [switch]$UpdateRepo
)

$ErrorActionPreference = "Stop"

# ---------------------------------------------------------------------------
# COMPONENT DEFINITIONS
# ---------------------------------------------------------------------------
$ComponentDefs = @{
    smiley = @{
        BuildPath    = ".\cmd\generic\smiley"
        Binary       = "smiley.exe"
        DefaultPort  = 8080
        MetricsPort  = 9090
        TaskName     = "FacesDemoSmiley"
        Protocol     = "HTTP"
        EnvDefaults  = [ordered]@{
            # Emoji to serve. Accepts: name (Grinning|Sleeping|Cursing|Kaboom|
            # HeartEyes|Neutral|RollingEyes|Screaming|Vomiting), Unicode
            # codepoint (U+1F603), HTML entity (&#x1F603;), or literal emoji.
            SMILEY            = "Grinning"
            PORT              = 8080
            ENABLE_PROMETHEUS = "true"
            # Prometheus metrics port - must be unique per component.
            # Requires rebuilt binaries (script patches source automatically).
            METRICS_PORT      = 9090
            # Fraction of requests (0-100) to return as errors.
            ERROR_FRACTION    = 0
            # Fraction of requests (0-100) to latch into error state.
            LATCH_FRACTION    = 0
            # Maximum request rate (req/s). 0 = unlimited.
            MAX_RATE          = 0
            # Comma-separated artificial delay buckets in ms.
            DELAY_BUCKETS     = ""
            USER_HEADER_NAME  = "X-Faces-User"
            DEBUG_ENABLED     = "false"
        }
        FirewallRules = @(
            @{ Name = "Faces Demo smiley HTTP";    Port = 8080 }
            @{ Name = "Faces Demo smiley Metrics"; Port = 9090 }
        )
    }
    color = @{
        BuildPath    = ".\cmd\generic\color"
        Binary       = "color.exe"
        DefaultPort  = 8081
        MetricsPort  = 9091
        TaskName     = "FacesDemoColor"
        Protocol     = "gRPC"
        EnvDefaults  = [ordered]@{
            # Color name: blue|darkblue|green|yellow|red|purple|grey|black|white
            # or a hex code like #66CCEE.
            COLOR             = "blue"
            PORT              = 8081
            ENABLE_PROMETHEUS = "true"
            # Prometheus metrics port - must be unique per component.
            # Requires rebuilt binaries (script patches source automatically).
            METRICS_PORT      = 9091
            ERROR_FRACTION    = 0
            LATCH_FRACTION    = 0
            MAX_RATE          = 0
            DELAY_BUCKETS     = ""
            USER_HEADER_NAME  = "X-Faces-User"
            DEBUG_ENABLED     = "false"
        }
        FirewallRules = @(
            @{ Name = "Faces Demo color HTTP";    Port = 8081 }
            @{ Name = "Faces Demo color Metrics"; Port = 9091 }
        )
    }
    face = @{
        BuildPath    = ".\cmd\generic\face"
        Binary       = "face.exe"
        DefaultPort  = 8082
        MetricsPort  = 9092
        TaskName     = "FacesDemoFace"
        Protocol     = "HTTP"
        EnvDefaults  = [ordered]@{
            # Hostname (or host:port) of the smiley backend.
            SMILEY_SERVICE    = "localhost:8080"
            # Hostname (or host:port) of the color backend (gRPC).
            COLOR_SERVICE     = "localhost:8081"
            PORT              = 8082
            ENABLE_PROMETHEUS = "true"
            # Prometheus metrics port - must be unique per component.
            # Requires rebuilt binaries (script patches source automatically).
            METRICS_PORT      = 9092
            ERROR_FRACTION    = 0
            LATCH_FRACTION    = 0
            MAX_RATE          = 0
            DELAY_BUCKETS     = ""
            USER_HEADER_NAME  = "X-Faces-User"
            DEBUG_ENABLED     = "false"
        }
        FirewallRules = @(
            @{ Name = "Faces Demo face HTTP";    Port = 8082 }
            @{ Name = "Faces Demo face Metrics"; Port = 9092 }
        )
    }
    gui = @{
        BuildPath    = ".\cmd\generic\gui"
        Binary       = "gui.exe"
        DefaultPort  = 8083
        MetricsPort  = 9093
        TaskName     = "FacesDemoGUI"
        Protocol     = "HTTP"
        EnvDefaults  = [ordered]@{
            # Path to the assets/html directory (web UI files).
            DATA_PATH         = "APPROOT_PLACEHOLDER\gui\data"
            # Hostname (or host:port) of the face backend.
            FACE_SERVICE      = "localhost:8082"
            # Background color of the grid.
            COLOR             = "white"
            PORT              = 8083
            NUM_ROWS          = 4
            NUM_COLS          = 4
            EDGE_SIZE         = 1
            HIDE_KEY          = "false"
            SHOW_PODS         = "false"
            # Whether the GUI starts in active (polling) mode. true = polling on load.
            START_ACTIVE      = "true"
            ENABLE_PROMETHEUS = "true"
            # Prometheus metrics port - must be unique per component
            METRICS_PORT      = 9093
            ERROR_FRACTION    = 0
            LATCH_FRACTION    = 0
            MAX_RATE          = 0
            DELAY_BUCKETS     = ""
            USER_HEADER_NAME  = "X-Faces-User"
            DEBUG_ENABLED     = "false"
        }
        FirewallRules = @(
            @{ Name = "Faces Demo gui HTTP";    Port = 8083 }
            @{ Name = "Faces Demo gui Metrics"; Port = 9093 }
        )
    }
}

# ---------------------------------------------------------------------------
# HELPERS
# ---------------------------------------------------------------------------
function Write-Step { param([string]$Msg) Write-Host "==> $Msg" -ForegroundColor Cyan }
function Write-Ok   { param([string]$Msg) Write-Host "    OK   $Msg" -ForegroundColor Green }
function Write-Warn { param([string]$Msg) Write-Host "    WARN $Msg" -ForegroundColor Yellow }
function Write-Err  { param([string]$Msg) Write-Host "    ERR  $Msg" -ForegroundColor Red }

function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $ts      = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $entry   = "[$ts][$Level] $Message"
    $logDir  = Join-Path $LogRoot "manager"
    $logFile = Join-Path $logDir "manage-faces-demo.log"
    try {
        if (-not (Test-Path $logDir)) { New-Item -ItemType Directory -Path $logDir -Force | Out-Null }
        Add-Content -Path $logFile -Value $entry -Encoding UTF8
    } catch {}
    Write-Host $entry
}

function Ensure-Dir {
    param([string]$Path)
    if (-not (Test-Path $Path)) { New-Item -ItemType Directory -Path $Path -Force | Out-Null }
}

function Test-Cmd {
    param([string]$Cmd)
    return [bool](Get-Command $Cmd -ErrorAction SilentlyContinue)
}

function Refresh-Path {
    $mp      = [Environment]::GetEnvironmentVariable("Path","Machine")
    $up      = [Environment]::GetEnvironmentVariable("Path","User")
    $env:Path = "$mp;$up;C:\Program Files\Git\cmd;C:\Program Files\Go\bin"
}

function Get-Arch {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { return "arm64" }
    return "amd64"
}

function Resolve-Components {
    param([string]$Name)
    if ($Name -eq "all") { return @("smiley","color","face","gui") }
    return @($Name)
}

# ---------------------------------------------------------------------------
# PREREQUISITES
# ---------------------------------------------------------------------------
function Install-GitDirect {
    if (Test-Cmd git) {
        Write-Ok "Git already installed: $(git --version 2>&1)"
        return
    }
    if ($SkipPrereqInstall) { throw "git.exe not found and -SkipPrereqInstall is set." }

    $arch = Get-Arch
    Write-Step "Downloading Git for Windows ($arch) from GitHub Releases"
    Write-Log "Downloading Git ($arch)"

    $release = Invoke-RestMethod `
        -Uri "https://api.github.com/repos/git-for-windows/git/releases/latest" `
        -Headers @{ "User-Agent" = "faces-demo-manager" }

    if ($arch -eq "arm64") {
        $asset = $release.assets |
            Where-Object { $_.name -match "Git-.*-arm64\.exe$" -and $_.name -notmatch "PortableGit|MinGit" } |
            Select-Object -First 1
    } else {
        $asset = $release.assets |
            Where-Object { $_.name -match "Git-.*-64-bit\.exe$" -and $_.name -notmatch "PortableGit|MinGit" } |
            Select-Object -First 1
    }
    if (-not $asset) { throw "Could not find Git installer asset in latest release." }

    $installer = Join-Path $env:TEMP $asset.name
    Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $installer -UseBasicParsing

    Write-Step "Installing Git silently"
    $p = Start-Process -FilePath $installer `
        -ArgumentList @("/VERYSILENT","/NORESTART","/NOCANCEL","/SP-","/SUPPRESSMSGBOXES","/CLOSEAPPLICATIONS") `
        -Wait -PassThru
    if ($p.ExitCode -ne 0) { throw "Git installer failed with exit code $($p.ExitCode)" }

    Refresh-Path
    if (-not (Test-Cmd git)) { throw "Git installed but git.exe not in PATH. Open a new elevated session and retry." }
    Write-Ok "Git installed: $(git --version 2>&1)"
    Write-Log "Git installed"
}

function Install-GoDirect {
    if (Test-Cmd go) {
        Write-Ok "Go already installed: $(go version 2>&1)"
        return
    }
    if ($SkipPrereqInstall) { throw "go.exe not found and -SkipPrereqInstall is set." }

    $goArch  = Get-Arch
    Write-Step "Fetching latest stable Go version from go.dev"
    $verResp = Invoke-WebRequest -Uri "https://go.dev/VERSION?m=text" -UseBasicParsing
    $goVer   = ($verResp.Content -split "`n" | Select-Object -First 1).Trim()
    if ($goVer -notmatch "^go\d+\.\d+(\.\d+)?$") { throw "Unexpected Go version string: $goVer" }

    $msiName = "$goVer.windows-$goArch.msi"
    $msiUrl  = "https://go.dev/dl/$msiName"
    $msiPath = Join-Path $env:TEMP $msiName

    Write-Step "Downloading $msiName"
    Invoke-WebRequest -Uri $msiUrl -OutFile $msiPath -UseBasicParsing

    Write-Step "Installing Go silently"
    $p = Start-Process -FilePath "msiexec.exe" `
        -ArgumentList @("/i", "`"$msiPath`"", "/qn", "/norestart") `
        -Wait -PassThru
    if ($p.ExitCode -ne 0) { throw "Go installer failed with exit code $($p.ExitCode)" }

    Refresh-Path
    if (-not (Test-Cmd go)) { throw "Go installed but go.exe not in PATH. Open a new elevated session and retry." }
    Write-Ok "Go installed: $(go version 2>&1)"
    Write-Log "Go installed"
}

# ---------------------------------------------------------------------------
# REPO / BUILD
# ---------------------------------------------------------------------------
function Sync-Repo {
    $repoDir = Join-Path $SourceRoot "faces-demo"
    Ensure-Dir $SourceRoot

    # Prevent git from ever hanging waiting for a credential prompt.
    $env:GIT_TERMINAL_PROMPT = "0"
    $env:GCM_INTERACTIVE     = "never"

    if (-not (Test-Path (Join-Path $repoDir ".git"))) {
        Write-Step "Cloning faces-demo repository"
        Write-Log "Cloning $RepoUrl to $repoDir"
        git clone --quiet $RepoUrl $repoDir
        if ($LASTEXITCODE -ne 0) { throw "git clone failed (exit $LASTEXITCODE). Check network or -RepoUrl." }
        Write-Ok "Repository cloned to $repoDir"
        Write-Log "Repository cloned"
    } else {
        if ($UpdateRepo) {
            Write-Step "Updating repository (-UpdateRepo specified)"
            Write-Log "Fetching latest in $repoDir"

            # Short TCP timeout via config so we fail fast instead of waiting 5 min
            git -C $repoDir -c http.lowSpeedLimit=1000 -c http.lowSpeedTime=15 `
                fetch --quiet --no-tags origin 2>&1 | Out-Null

            if ($LASTEXITCODE -ne 0) {
                Write-Warn "git fetch failed - building from existing local source"
                Write-Log  "git fetch failed, using local source" "WARN"
            } else {
                git -C $repoDir merge --ff-only --quiet origin/main 2>&1 | Out-Null
                if ($LASTEXITCODE -ne 0) {
                    git -C $repoDir merge --ff-only --quiet origin/master 2>&1 | Out-Null
                }
                if ($LASTEXITCODE -eq 0) {
                    Write-Ok "Repository updated"
                    Write-Log "Repository updated"
                } else {
                    Write-Warn "Could not fast-forward - building from existing local source"
                    Write-Log  "git merge ff failed, using local source" "WARN"
                }
            }
        } else {
            Write-Ok "Repository already cloned at $repoDir (use -UpdateRepo to pull latest)"
            Write-Log "Using existing local repository"
        }
    }
    return $repoDir
}

function Patch-PrometheusSource {
    # The upstream prometheus.go hardcodes port 9090 and calls log.Fatalf on
    # any bind error, which kills the whole process. This patch rewrites it to:
    #   - read METRICS_PORT env var (default 9090)
    #   - log a warning and continue if the port is already in use
    param([string]$RepoDir)

    $target = Join-Path $RepoDir "pkg\faces\prometheus.go"
    if (-not (Test-Path $target)) {
        Write-Warn "prometheus.go not found at $target - skipping patch"
        return
    }

    # Check if already patched
    $current = Get-Content $target -Raw
    if ($current -match "METRICS_PORT") {
        Write-Ok "prometheus.go already patched"
        return
    }

    Write-Step "Patching pkg/faces/prometheus.go to support METRICS_PORT env var"

    $L = [System.Collections.Generic.List[string]]::new()
    $L.Add('// SPDX-FileCopyrightText: 2025 Buoyant Inc.')
    $L.Add('// SPDX-License-Identifier: Apache-2.0')
    $L.Add('//')
    $L.Add('// Patched by Faces-Demo-Windows-Manager.ps1 to read METRICS_PORT from environment')
    $L.Add('// and degrade gracefully on port conflict instead of calling log.Fatalf.')
    $L.Add('')
    $L.Add('package faces')
    $L.Add('')
    $L.Add('import (')
    $L.Add('	"fmt"')
    $L.Add('	"log/slog"')
    $L.Add('	"net/http"')
    $L.Add('	"os"')
    $L.Add('')
    $L.Add('	"github.com/prometheus/client_golang/prometheus/promhttp"')
    $L.Add(')')
    $L.Add('')
    $L.Add('// Start the HTTP server for Prometheus metrics.')
    $L.Add('// Reads METRICS_PORT from the environment (default: 9090).')
    $L.Add('// Logs a warning and continues if the port cannot be bound.')
    $L.Add('func StartPrometheusServer() {')
    $L.Add('	port := os.Getenv("METRICS_PORT")')
    $L.Add('	if port == "" {')
    $L.Add('		port = "9090"')
    $L.Add('	}')
    $L.Add('	addr := fmt.Sprintf(":%s", port)')
    $L.Add('')
    $L.Add('	promServer := &http.Server{')
    $L.Add('		Handler: promhttp.Handler(),')
    $L.Add('		Addr:    addr,')
    $L.Add('	}')
    $L.Add('')
    $L.Add('	go func() {')
    $L.Add('		if err := promServer.ListenAndServe(); err != nil {')
    $L.Add('			slog.Warn(fmt.Sprintf("Metrics server on %s unavailable: %v", addr, err))')
    $L.Add('		}')
    $L.Add('	}()')
    $L.Add('}')

    [System.IO.File]::WriteAllLines($target, $L, [System.Text.Encoding]::UTF8)
    Write-Ok  "prometheus.go patched"
    Write-Log "prometheus.go patched in $RepoDir"
}

function Build-Components {
    param([string[]]$Names)

    $repoDir = Join-Path $SourceRoot "faces-demo"
    if (-not (Test-Path $repoDir)) { $repoDir = Sync-Repo }

    Patch-PrometheusSource -RepoDir $repoDir

    Ensure-Dir (Join-Path $AppRoot "bin")
    $env:CGO_ENABLED = "0"

    foreach ($name in $Names) {
        $def    = $ComponentDefs[$name]
        $outExe = Join-Path $AppRoot "bin\$($def.Binary)"

        Write-Step "Building $name -> $outExe"
        Write-Log  "Building $name"

        Push-Location $repoDir
        try {
            go build -o $outExe $def.BuildPath
            if (-not (Test-Path $outExe)) { throw "Binary not created: $outExe" }
            Write-Ok "$name built"
            Write-Log "$name built: $outExe"
        } finally {
            Pop-Location
        }
    }
}

# ---------------------------------------------------------------------------
# CONFIG / ENV FILE
# ---------------------------------------------------------------------------
function Get-EnvFilePath {
    param([string]$Name)
    return Join-Path $AppRoot "$Name\env\$Name.env"
}

function Write-DefaultEnvFile {
    param([string]$Name)

    $def     = $ComponentDefs[$Name]
    $envFile = Get-EnvFilePath $Name
    $envDir  = Split-Path $envFile

    Ensure-Dir $envDir

    if (Test-Path $envFile) {
        if ($ResetConfig) {
            Remove-Item $envFile -Force
            Write-Warn "Config reset (-ResetConfig): $envFile"
            Write-Log  "Config reset: $envFile" "WARN"
        } else {
            Write-Warn "Config already exists (preserving edits): $envFile"
            Write-Log  "Config exists, skipping write: $envFile" "WARN"
            return
        }
    }

    $lines = [System.Collections.Generic.List[string]]::new()
    $lines.Add("# Faces Demo - $Name component configuration")
    $lines.Add("# Generated: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')")
    $lines.Add("# Edit this file then run:")
    $lines.Add("#   .\Faces-Demo-Windows-Manager.ps1 -Command Restart -Component $Name")
    $lines.Add("")

    foreach ($kv in $def.EnvDefaults.GetEnumerator()) {
        if ($kv.Key -in @("PORT","DEBUG_ENABLED","ENABLE_PROMETHEUS")) { $lines.Add("") }
        # Resolve APPROOT_PLACEHOLDER so DATA_PATH reflects the actual $AppRoot
        $val = "$($kv.Value)" -replace "APPROOT_PLACEHOLDER", $AppRoot
        $lines.Add("$($kv.Key)=$val")
    }

    [System.IO.File]::WriteAllLines($envFile, $lines, [System.Text.Encoding]::UTF8)
    Write-Ok  "Config written: $envFile"
    Write-Log "Config written: $envFile"
}

function Read-EnvFile {
    param([string]$Path)
    $vars = @{}
    if (-not (Test-Path $Path)) { return $vars }
    Get-Content $Path | ForEach-Object {
        $line = $_.Trim()
        if ($line -eq "" -or $line.StartsWith("#")) { return }
        $parts = $line -split "=", 2
        if ($parts.Count -eq 2) { $vars[$parts[0].Trim()] = $parts[1].Trim() }
    }
    return $vars
}

function Apply-PortOverride {
    param([string]$Name, [int]$NewPort)
    if ($NewPort -le 0) { return }
    $envFile = Get-EnvFilePath $Name
    if (-not (Test-Path $envFile)) { Write-Warn "Config not found for port override: $envFile"; return }
    $content = Get-Content $envFile
    $updated = $content -replace "^PORT=.*", "PORT=$NewPort"
    $updated | Set-Content $envFile -Encoding UTF8
    Write-Ok  "Port updated to $NewPort in $envFile"
    Write-Log "Port set to $NewPort in $envFile"
}

function Repair-EnvFile {
    # Add any keys from EnvDefaults that are missing in the on-disk env file.
    # Existing values are never changed - this is purely additive.
    param([string]$Name)

    $def     = $ComponentDefs[$Name]
    $envFile = Get-EnvFilePath $Name

    if (-not (Test-Path $envFile)) {
        Write-Warn "Repair-EnvFile: $envFile not found, skipping"
        return
    }

    $existing = Read-EnvFile $envFile
    $added    = [System.Collections.Generic.List[string]]::new()

    foreach ($kv in $def.EnvDefaults.GetEnumerator()) {
        if (-not $existing.ContainsKey($kv.Key)) {
            $added.Add("$($kv.Key)=$($kv.Value)")
        }
    }

    if ($added.Count -gt 0) {
        $header = "# Added by Repair-EnvFile on $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
        Add-Content -Path $envFile -Value "" -Encoding UTF8
        Add-Content -Path $envFile -Value $header -Encoding UTF8
        foreach ($line in $added) {
            Add-Content -Path $envFile -Value $line -Encoding UTF8
            Write-Ok "  Added to $Name config: $line"
        }
        Write-Log "Repair-EnvFile added $($added.Count) missing keys to $envFile"
    } else {
        Write-Ok "$Name config is already up to date"
    }
}

# ---------------------------------------------------------------------------
# LAUNCHER SCRIPT
# Each component gets a small run-<name>.ps1 that loads its .env and starts
# the binary. Built line-by-line (no here-strings) to avoid encoding issues.
# ---------------------------------------------------------------------------
function Write-LauncherScript {
    param([string]$Name)

    $def         = $ComponentDefs[$Name]
    $compDir     = Join-Path $AppRoot $Name
    $exePath     = Join-Path $AppRoot "bin\$($def.Binary)"
    $envFile     = Get-EnvFilePath $Name
    $logDir      = Join-Path $LogRoot $Name
    $launcherPs  = Join-Path $compDir "run-$Name.ps1"
    $stdoutLog   = Join-Path $logDir "$Name.stdout.log"
    $stderrLog   = Join-Path $logDir "$Name.stderr.log"
    $defaultPort = $def.DefaultPort

    Ensure-Dir $compDir

    # Copy GUI web assets so DATA_PATH works out of the box
    if ($Name -eq "gui") {
        $repoAssets = Join-Path $SourceRoot "faces-demo\assets\html"
        $destData   = Join-Path $compDir "data"
        if (Test-Path $repoAssets) {
            Write-Step "Copying GUI web assets to $destData"
            Ensure-Dir $destData
            Copy-Item -Path "$repoAssets\*" -Destination $destData -Recurse -Force
            Write-Log "GUI assets copied to $destData"
        } else {
            Write-Warn "Repo assets not found at $repoAssets - update DATA_PATH in config manually"
        }
    }

    # Build the launcher as a list of plain strings - no here-string, no escaping issues
    $L = [System.Collections.Generic.List[string]]::new()
    $L.Add('# Auto-generated launcher for Faces Demo - ' + $Name)
    $L.Add('# Do NOT edit this file directly.')
    $L.Add('# Edit the config file and restart:')
    $L.Add('#   ' + $envFile)
    $L.Add('$ErrorActionPreference = "Stop"')
    $L.Add('')
    $L.Add('$FD_EnvFile = "' + $envFile + '"')
    $L.Add('$FD_ExePath = "' + $exePath + '"')
    $L.Add('$FD_LogDir  = "' + $logDir + '"')
    $L.Add('$FD_Stdout  = "' + $stdoutLog + '"')
    $L.Add('$FD_Stderr  = "' + $stderrLog + '"')
    $L.Add('')
    $L.Add('if (-not (Test-Path $FD_LogDir)) {')
    $L.Add('    New-Item -ItemType Directory -Path $FD_LogDir -Force | Out-Null')
    $L.Add('}')
    $L.Add('')
    $L.Add('if (-not (Test-Path $FD_EnvFile)) {')
    $L.Add('    throw "Config file missing: $FD_EnvFile"')
    $L.Add('}')
    $L.Add('')
    $L.Add('Get-Content $FD_EnvFile | ForEach-Object {')
    $L.Add('    $fd_line = $_.Trim()')
    $L.Add('    if ($fd_line -eq "" -or $fd_line.StartsWith("#")) { return }')
    $L.Add('    $fd_parts = $fd_line -split "=", 2')
    $L.Add('    if ($fd_parts.Count -eq 2) {')
    $L.Add('        [Environment]::SetEnvironmentVariable(')
    $L.Add('            $fd_parts[0].Trim(), $fd_parts[1].Trim(), "Process")')
    $L.Add('    }')
    $L.Add('}')
    $L.Add('')
    $L.Add('$fd_port = [Environment]::GetEnvironmentVariable("PORT", "Process")')
    $L.Add('if ([string]::IsNullOrWhiteSpace($fd_port)) { $fd_port = "' + $defaultPort + '" }')
    $L.Add('')
    $L.Add('Set-Location (Split-Path $FD_ExePath)')
    $L.Add('')
    $L.Add('& $FD_ExePath -port $fd_port 1>> $FD_Stdout 2>> $FD_Stderr')

    [System.IO.File]::WriteAllLines($launcherPs, $L, [System.Text.Encoding]::UTF8)
    Write-Ok  "Launcher written: $launcherPs"
    Write-Log "Launcher written: $launcherPs"
    return $launcherPs
}

# ---------------------------------------------------------------------------
# SCHEDULED TASK
# ---------------------------------------------------------------------------
function Register-ComponentTask {
    param([string]$Name, [string]$LauncherPath)

    $taskName = $ComponentDefs[$Name].TaskName

    $existing = Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue
    if ($existing) {
        Stop-ScheduledTask       -TaskName $taskName -ErrorAction SilentlyContinue
        Unregister-ScheduledTask -TaskName $taskName -Confirm:$false -ErrorAction SilentlyContinue
        Write-Log "Removed existing task: $taskName"
    }

    $action = New-ScheduledTaskAction `
        -Execute          "powershell.exe" `
        -Argument         "-NoProfile -ExecutionPolicy Bypass -File `"$LauncherPath`"" `
        -WorkingDirectory (Split-Path $LauncherPath)

    $trigger = New-ScheduledTaskTrigger -AtStartup

    $principal = New-ScheduledTaskPrincipal `
        -UserId    "SYSTEM" `
        -LogonType ServiceAccount `
        -RunLevel  Highest

    $settings = New-ScheduledTaskSettingsSet `
        -AllowStartIfOnBatteries `
        -DontStopIfGoingOnBatteries `
        -ExecutionTimeLimit (New-TimeSpan -Days 0) `
        -RestartCount       999 `
        -RestartInterval    (New-TimeSpan -Minutes 1)

    Register-ScheduledTask `
        -TaskName    $taskName `
        -Action      $action `
        -Trigger     $trigger `
        -Principal   $principal `
        -Settings    $settings `
        -Description "Faces Demo $Name component. Config: $(Get-EnvFilePath $Name)" `
        -Force | Out-Null

    Write-Ok  "Scheduled task registered: $taskName"
    Write-Log "Task registered: $taskName"
}

# ---------------------------------------------------------------------------
# FIREWALL
# ---------------------------------------------------------------------------
function Add-FirewallRules {
    param([string]$Name)
    foreach ($rule in $ComponentDefs[$Name].FirewallRules) {
        if (-not (Get-NetFirewallRule -DisplayName $rule.Name -ErrorAction SilentlyContinue)) {
            New-NetFirewallRule -DisplayName $rule.Name -Direction Inbound `
                -Action Allow -Protocol TCP -LocalPort $rule.Port | Out-Null
            Write-Ok  "Firewall rule added: $($rule.Name) (TCP $($rule.Port))"
            Write-Log "Firewall rule added: $($rule.Name) port $($rule.Port)"
        } else {
            Write-Ok  "Firewall rule already present: $($rule.Name)"
        }
    }
}

function Remove-FirewallRules {
    param([string]$Name)
    foreach ($rule in $ComponentDefs[$Name].FirewallRules) {
        if (Get-NetFirewallRule -DisplayName $rule.Name -ErrorAction SilentlyContinue) {
            Remove-NetFirewallRule -DisplayName $rule.Name
            Write-Ok  "Firewall rule removed: $($rule.Name)"
            Write-Log "Firewall rule removed: $($rule.Name)"
        }
    }
}

# ---------------------------------------------------------------------------
# START / STOP / STATUS
# ---------------------------------------------------------------------------
function Start-Component {
    param([string]$Name)
    $taskName = $ComponentDefs[$Name].TaskName
    $task     = Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue
    if (-not $task) {
        Write-Err "$Name task not found. Run -Command Install first."
        Write-Log "$Name task not found" "ERROR"
        return
    }
    Write-Step "Starting $Name"
    Start-ScheduledTask -TaskName $taskName
    Start-Sleep -Seconds 2
    $state = (Get-ScheduledTask -TaskName $taskName).State
    Write-Ok  "$Name task state: $state"
    Write-Log "$Name started - state: $state"
}

function Stop-Component {
    param([string]$Name)
    $taskName = $ComponentDefs[$Name].TaskName
    $task     = Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue
    if (-not $task) { Write-Warn "$Name task not found."; return }
    Write-Step "Stopping $Name"
    Stop-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 1
    $state = (Get-ScheduledTask -TaskName $taskName).State
    Write-Ok  "$Name task state: $state"
    Write-Log "$Name stopped - state: $state"
}

function Get-ComponentStatus {
    param([string]$Name)

    $def      = $ComponentDefs[$Name]
    $taskName = $def.TaskName
    $logDir   = Join-Path $LogRoot $Name
    $envFile  = Get-EnvFilePath $Name

    $task = Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue
    if (-not $task) {
        Write-Host "  [$Name]  NOT INSTALLED" -ForegroundColor Red
        return
    }

    $info    = Get-ScheduledTaskInfo -TaskName $taskName -ErrorAction SilentlyContinue
    $state   = $task.State
    $lastRun = if ($info) { $info.LastRunTime } else { "n/a" }
    $lastRes = if ($info) { "0x{0:X8}" -f $info.LastTaskResult } else { "n/a" }

    $col = switch ($state) { "Running" { "Green" } "Ready" { "Yellow" } default { "Red" } }

    $envVars = Read-EnvFile $envFile
    $port    = if ($envVars["PORT"]) { $envVars["PORT"] } else { $def.DefaultPort }

    Write-Host ""
    Write-Host ("  [{0}]  State: " -f $Name) -ForegroundColor Cyan -NoNewline
    Write-Host $state -ForegroundColor $col
    Write-Host "    Task name : $taskName"
    Write-Host "    Config    : $envFile"
    $proto = $def.Protocol
    Write-Host "    Listen    : $($proto.ToLower())://127.0.0.1:$port"
    Write-Host "    Last run  : $lastRun  (result: $lastRes)"
    Write-Host "    Logs      : $logDir"

    if ($proto -eq "HTTP") {
        try {
            Invoke-RestMethod -Uri "http://127.0.0.1:$port/center/" -TimeoutSec 2 -ErrorAction Stop | Out-Null
            Write-Host "    Endpoint  : OK" -ForegroundColor Green
        } catch {
            Write-Host "    Endpoint  : not responding on port $port" -ForegroundColor Yellow
        }
    } else {
        Write-Host "    Endpoint  : gRPC (no HTTP health check)"
    }

    $stdLog = Join-Path $logDir "$Name.stdout.log"
    if (Test-Path $stdLog) {
        Write-Host ""
        Write-Host "    -- last $LogLines lines of stdout --" -ForegroundColor DarkGray
        Get-Content $stdLog -Tail $LogLines | ForEach-Object { Write-Host "    $_" -ForegroundColor DarkGray }
    }
}

# ---------------------------------------------------------------------------
# LOGS
# ---------------------------------------------------------------------------
function Show-Logs {
    param([string]$Name)
    $logDir    = Join-Path $LogRoot $Name
    $stdoutLog = Join-Path $logDir "$Name.stdout.log"
    $stderrLog = Join-Path $logDir "$Name.stderr.log"

    Write-Host ""
    Write-Host "=== STDOUT  $Name  ($stdoutLog) ===" -ForegroundColor Cyan
    if (Test-Path $stdoutLog) { Get-Content $stdoutLog -Tail $LogLines }
    else { Write-Host "  (no stdout log yet)" -ForegroundColor DarkGray }

    Write-Host ""
    Write-Host "=== STDERR  $Name  ($stderrLog) ===" -ForegroundColor Yellow
    if (Test-Path $stderrLog) { Get-Content $stderrLog -Tail $LogLines }
    else { Write-Host "  (no stderr log yet)" -ForegroundColor DarkGray }
}

# ---------------------------------------------------------------------------
# UNINSTALL
# ---------------------------------------------------------------------------
function Uninstall-Component {
    param([string]$Name)
    $taskName = $ComponentDefs[$Name].TaskName
    Write-Step "Uninstalling $Name"
    Write-Log  "Uninstalling $Name"

    Stop-ScheduledTask       -TaskName $taskName -ErrorAction SilentlyContinue
    Unregister-ScheduledTask -TaskName $taskName -Confirm:$false -ErrorAction SilentlyContinue
    Write-Ok "Task removed: $taskName"

    Remove-FirewallRules -Name $Name

    if ($Force) {
        $compDir = Join-Path $AppRoot $Name
        $logDir  = Join-Path $LogRoot $Name
        if (Test-Path $compDir) { Remove-Item $compDir -Recurse -Force; Write-Ok "Removed: $compDir" }
        if (Test-Path $logDir)  { Remove-Item $logDir  -Recurse -Force; Write-Ok "Removed: $logDir" }
        Write-Log "Files removed for $Name"
    } else {
        Write-Warn "Config and logs retained. Use -Force to delete them."
    }
}

# ---------------------------------------------------------------------------
# HELP
# ---------------------------------------------------------------------------
function Show-Help {
    Get-Help $PSCommandPath -Full

    Write-Host ""
    Write-Host "COMPONENT REFERENCE" -ForegroundColor Cyan
    Write-Host "-------------------------------------------------------------------"
    Write-Host "  smiley  HTTP backend - returns an emoji"
    Write-Host "          Default port: 8080   Metrics: 9090"
    Write-Host "          Key vars: SMILEY, ERROR_FRACTION, LATCH_FRACTION"
    Write-Host ""
    Write-Host "  color   gRPC backend - returns a hex colour"
    Write-Host "          Default port: 8081   Metrics: 9091"
    Write-Host "          Key vars: COLOR, ERROR_FRACTION"
    Write-Host ""
    Write-Host "  face    HTTP backend - calls smiley + color"
    Write-Host "          Default port: 8082   Metrics: 9092"
    Write-Host "          Key vars: SMILEY_SERVICE, COLOR_SERVICE"
    Write-Host ""
    Write-Host "  gui     HTTP front-end - serves the browser UI"
    Write-Host "          Default port: 8083   Metrics: 9093"
    Write-Host "          Key vars: FACE_SERVICE, DATA_PATH, NUM_ROWS, NUM_COLS"
    Write-Host ""
    Write-Host "CONFIG FILES" -ForegroundColor Cyan
    Write-Host "-------------------------------------------------------------------"
    Write-Host "  $AppRoot\smiley\env\smiley.env"
    Write-Host "  $AppRoot\color\env\color.env"
    Write-Host "  $AppRoot\face\env\face.env"
    Write-Host "  $AppRoot\gui\env\gui.env"
    Write-Host ""
    Write-Host "LOG FILES" -ForegroundColor Cyan
    Write-Host "-------------------------------------------------------------------"
    Write-Host "  $LogRoot\smiley\smiley.stdout.log"
    Write-Host "  $LogRoot\smiley\smiley.stderr.log"
    Write-Host "  (same pattern for color, face, gui)"
    Write-Host "  $LogRoot\manager\manage-faces-demo.log"
    Write-Host ""
    Write-Host "QUICK WORKFLOW" -ForegroundColor Cyan
    Write-Host "-------------------------------------------------------------------"
    Write-Host "  1. .\Faces-Demo-Windows-Manager.ps1 -Command Install"
    Write-Host "  2. Edit $AppRoot\gui\env\gui.env if needed"
    Write-Host "  3. .\Faces-Demo-Windows-Manager.ps1 -Command Restart -Component gui"
    Write-Host "  4. .\Faces-Demo-Windows-Manager.ps1 -Command Status"
    Write-Host "  5. Open http://localhost:8083 in a browser"
    Write-Host ""
}

# ---------------------------------------------------------------------------
# MAIN
# ---------------------------------------------------------------------------
Ensure-Dir (Join-Path $LogRoot "manager")
Write-Log "====== Faces-Demo-Windows-Manager.ps1 started  Command=$Command  Component=$Component ======"

$id        = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = New-Object Security.Principal.WindowsPrincipal($id)
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    throw "Must be run from an elevated (Administrator) PowerShell session."
}

$targets = Resolve-Components $Component

switch ($Command) {

    "Help" {
        Show-Help
    }

    "Build" {
        Write-Step "Checking prerequisites"
        Install-GitDirect
        Install-GoDirect
        Sync-Repo | Out-Null
        Write-Step "Building: $($targets -join ', ')"
        Build-Components -Names $targets
        Write-Log "Build complete: $($targets -join ', ')"
    }

    "Install" {
        Write-Step "Checking prerequisites"
        Install-GitDirect
        Install-GoDirect

        if (-not $SkipBuild) {
            Sync-Repo | Out-Null
            Write-Step "Building: $($targets -join ', ')"
            Build-Components -Names $targets
        } else {
            Write-Warn "Skipping build (-SkipBuild set)"
        }

        foreach ($name in $targets) {
            Write-Step "Deploying $name"
            Ensure-Dir (Join-Path $AppRoot $name)
            Ensure-Dir (Join-Path $LogRoot $name)
            Write-DefaultEnvFile -Name $name
            Apply-PortOverride   -Name $name -NewPort $Port
            $launcher = Write-LauncherScript -Name $name
            Register-ComponentTask -Name $name -LauncherPath $launcher
            Add-FirewallRules      -Name $name
        }

        Write-Step "Starting all deployed components"
        foreach ($name in $targets) { Start-Component -Name $name }

        Write-Host ""
        Write-Host "==========================================================" -ForegroundColor Green
        Write-Host "  Faces Demo install complete!" -ForegroundColor Green
        Write-Host "==========================================================" -ForegroundColor Green
        Write-Host ""
        Write-Host "  GUI URL     : http://localhost:8083"
        Write-Host "  Config dir  : $AppRoot\<component>\env\"
        Write-Host "  Log dir     : $LogRoot\<component>\"
        Write-Host ""
        Write-Host "  .\Faces-Demo-Windows-Manager.ps1 -Command Status"
        Write-Host "  .\Faces-Demo-Windows-Manager.ps1 -Command Logs -Component gui"
        Write-Host ""
        Write-Log "Install complete: $($targets -join ', ')"
    }

    "Configure" {
        # Set up dirs, env files, launcher scripts, scheduled tasks, and firewall
        # rules for each component - but do NOT build binaries and do NOT start
        # anything. Use this after Build to get everything ready to edit and then
        # start components individually with -Command Start -Component <name>.
        foreach ($name in $targets) {
            Write-Step "Configuring $name"
            Ensure-Dir (Join-Path $AppRoot $name)
            Ensure-Dir (Join-Path $LogRoot $name)
            Write-DefaultEnvFile   -Name $name
            Apply-PortOverride     -Name $name -NewPort $Port
            $launcher = Write-LauncherScript -Name $name
            Register-ComponentTask -Name $name -LauncherPath $launcher
            Add-FirewallRules      -Name $name
        }

        Write-Host ""
        Write-Host "==========================================================" -ForegroundColor Cyan
        Write-Host "  Configure complete - nothing started yet." -ForegroundColor Cyan
        Write-Host "==========================================================" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "  Edit config files, then start each component when ready:"
        Write-Host ""
        foreach ($name in $targets) {
            $envFile = Get-EnvFilePath $name
            Write-Host "  $envFile"
        }
        Write-Host ""
        Write-Host "  .\Faces-Demo-Windows-Manager.ps1 -Command Start -Component smiley"
        Write-Host "  .\Faces-Demo-Windows-Manager.ps1 -Command Start -Component color"
        Write-Host "  .\Faces-Demo-Windows-Manager.ps1 -Command Start -Component face"
        Write-Host "  .\Faces-Demo-Windows-Manager.ps1 -Command Start -Component gui"
        Write-Host ""
        Write-Log "Configure complete: $($targets -join ', ')"
    }

    "Repair" {
        # Adds any env var keys that exist in the defaults but are missing from
        # the on-disk config files. Safe to run at any time - never overwrites
        # values you have already set. Useful after upgrading the script.
        foreach ($name in $targets) {
            Write-Step "Repairing $name config"
            Repair-EnvFile -Name $name
        }
        Write-Host ""
        Write-Host "Repair complete. Restart affected components to pick up changes:"
        Write-Host "  .\Faces-Demo-Windows-Manager.ps1 -Command Restart -Component <name>"
        Write-Host ""
        Write-Log "Repair complete: $($targets -join ', ')"
    }

    "Start" {
        foreach ($name in $targets) { Start-Component -Name $name }
    }

    "Stop" {
        foreach ($name in $targets) { Stop-Component -Name $name }
    }

    "Restart" {
        foreach ($name in $targets) { Stop-Component  -Name $name }
        Start-Sleep -Seconds 2
        foreach ($name in $targets) { Start-Component -Name $name }
    }

    "Status" {
        Write-Host ""
        Write-Host "Faces Demo Component Status" -ForegroundColor Cyan
        Write-Host "=========================================================="
        foreach ($name in $targets) { Get-ComponentStatus -Name $name }
        Write-Host ""
    }

    "Logs" {
        foreach ($name in $targets) { Show-Logs -Name $name }
    }

    "Uninstall" {
        Write-Warn "Uninstalling: $($targets -join ', ')"
        if ($Force) { Write-Warn "Files will be permanently deleted (-Force set)." }
        foreach ($name in $targets) { Uninstall-Component -Name $name }
        if ($Component -eq "all" -and $Force) {
            $binDir = Join-Path $AppRoot "bin"
            if (Test-Path $AppRoot) {
                if (-not (Test-Path $binDir) -or -not (Get-ChildItem $binDir -ErrorAction SilentlyContinue)) {
                    Remove-Item $AppRoot -Recurse -Force -ErrorAction SilentlyContinue
                    Write-Ok "Removed: $AppRoot"
                }
            }
        }
        Write-Log "Uninstall complete"
    }
}

Write-Log "====== Faces-Demo-Windows-Manager.ps1 finished  Command=$Command ======"