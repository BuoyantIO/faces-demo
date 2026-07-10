# Faces Demo Windows Manager

Windows manager for the [Faces Demo](https://github.com/BuoyantIO/faces-demo) — builds, configures, and runs all four components as Windows Scheduled Tasks.

## Requirements

- Windows 10 / Server 2016+, PowerShell 5.1+
- Run as Administrator
- Internet access on first run (Git and Go auto-installed if missing)

## Quick Start

```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force
.\Faces-Demo-Windows-Manager.ps1 Install
```

Then open `http://localhost:8083` in a browser.

---

## Commands

| Command | Description |
|---|---|
| `Install` | Install prereqs, build, configure, and start everything |
| `Build` | Compile binaries from source (clones repo if needed) |
| `Configure` | Create dirs, config files, tasks, firewall rules — no build, no start |
| `Repair` | Add missing config keys to existing env files (never overwrites) |
| `Start` | Start one or all component tasks |
| `Stop` | Stop one or all component tasks |
| `Restart` | Stop then start |
| `Status` | Show task state, ports, and recent log lines |
| `Logs` | Tail stdout/stderr logs |
| `Uninstall` | Remove tasks and firewall rules (add `-Force` to delete files too) |
| `Help` | Full built-in help |

**Target a single component with `-Component`:**

```powershell
.\Faces-Demo-Windows-Manager.ps1 Start   -Component smiley
.\Faces-Demo-Windows-Manager.ps1 Restart -Component face
.\Faces-Demo-Windows-Manager.ps1 Logs    -Component gui -LogLines 100
```

Valid values: `smiley` | `color` | `face` | `gui` | `all` (default)

---

## Recommended Workflow (build first, start when ready)

```powershell
.\Faces-Demo-Windows-Manager.ps1 Build
.\Faces-Demo-Windows-Manager.ps1 Configure

# Edit any config files you want before starting
notepad <AppRoot>\smiley\env\smiley.env

.\Faces-Demo-Windows-Manager.ps1 Start -Component smiley
.\Faces-Demo-Windows-Manager.ps1 Start -Component color
.\Faces-Demo-Windows-Manager.ps1 Start -Component face
.\Faces-Demo-Windows-Manager.ps1 Start -Component gui
.\Faces-Demo-Windows-Manager.ps1 Status
```

---

## Components

| Component | Role | HTTP Port | Metrics Port | Task Name |
|---|---|---|---|---|
| `smiley` | Returns an emoji | 8080 | 9090 | FacesDemoSmiley |
| `color` | Returns a hex colour (gRPC) | 8081 | 9091 | FacesDemoColor |
| `face` | Calls smiley + color | 8082 | 9092 | FacesDemoFace |
| `gui` | Browser UI, calls face | 8083 | 9093 | FacesDemoGUI |

---

## Config Files

One `.env` file per component — written on first run, **never overwritten**.

Paths use the `-AppRoot` parameter (default `C:\faces-demo`):

```
<AppRoot>\smiley\env\smiley.env   SMILEY, PORT, METRICS_PORT, ERROR_FRACTION ...
<AppRoot>\color\env\color.env     COLOR, PORT, METRICS_PORT ...
<AppRoot>\face\env\face.env       SMILEY_SERVICE, COLOR_SERVICE, PORT ...
<AppRoot>\gui\env\gui.env         FACE_SERVICE, DATA_PATH, NUM_ROWS, NUM_COLS, START_ACTIVE ...
```

Edit a file, then restart that component:

```powershell
.\Faces-Demo-Windows-Manager.ps1 Restart -Component smiley
```

If you upgrade the script and configs are missing keys, run:

```powershell
.\Faces-Demo-Windows-Manager.ps1 Repair
```

---

## Logs

Log paths use the `-LogRoot` parameter (default `C:\temp\faces-demo`):

```
<LogRoot>\<component>\<component>.stdout.log
<LogRoot>\<component>\<component>.stderr.log
<LogRoot>\manager\manage-faces-demo.log
```

---

## Key Parameters

| Parameter | Default | Notes |
|---|---|---|
| `-Component` | `all` | `smiley`, `color`, `face`, `gui`, or `all` |
| `-Port` | component default | Override HTTP listen port |
| `-LogLines` | `50` | Lines shown by `Status` / `Logs` |
| `-SourceRoot` | `C:\src` | Where the repo is cloned |
| `-AppRoot` | `C:\faces-demo` | Parent directory for binaries and config |
| `-LogRoot` | `C:\temp\faces-demo` | Parent directory for log files |
| `-Branch` | repo default (`main`) | Git branch to build from. The faces-admin linky images require workloads built from `faces-demo-3.0` until it merges to `main` — use `-Branch faces-demo-3.0` (with `-UpdateRepo` to switch an existing clone) |
| `-UpdateRepo` | off | Pull latest source before building (also switches to `-Branch` if set) |
| `-SkipBuild` | off | Skip Go compile during `Install` |
| `-ResetConfig` | off | Wipe and regenerate env files before `Configure` / `Install` |
| `-Force` | off | `Uninstall`: also delete files |

---

## Troubleshooting

**Task state is `Ready` instead of `Running`** — the process exited. Check logs:
```powershell
.\Faces-Demo-Windows-Manager.ps1 -Command Logs -Component smiley
```

**Metrics port conflict** — run `Repair` then rebuild and restart:
```powershell
.\Faces-Demo-Windows-Manager.ps1 -Command Repair
.\Faces-Demo-Windows-Manager.ps1 -Command Build
.\Faces-Demo-Windows-Manager.ps1 -Command Restart
```

**Services running but GUI shows errors / components calling wrong ports** — a stale config from a previous install is being preserved. Reset all env files to defaults and restart:
```powershell
.\Faces-Demo-Windows-Manager.ps1 -Command Configure -ResetConfig
.\Faces-Demo-Windows-Manager.ps1 -Command Restart
```
To reset a single component only:
```powershell
.\Faces-Demo-Windows-Manager.ps1 -Command Configure -Component face -ResetConfig
.\Faces-Demo-Windows-Manager.ps1 -Command Restart -Component face
```
⚠️ `-ResetConfig` permanently discards any customisations in the env files.

**Build hangs on git fetch** — omit `-UpdateRepo`; the local clone is used by default.

**Git/Go not found after install** — close and reopen the elevated PowerShell window.
