# Stratavore for Windows 🪟

Quick start guide for Windows users.

---

## ⚡ Quick Start (5 Minutes)

### 1. Prerequisites

**Required:**
- Windows 10/11
- Go 1.22+ ([download](https://go.dev/dl/))
- PostgreSQL 14+ (running somewhere)
- RabbitMQ 3.12+ (running somewhere)

**Optional:**
- Docker Desktop (for infrastructure)
- Git (for version info in builds)

---

### 2. Build

**PowerShell (Recommended):**
```powershell
.\build.ps1
```

**Command Prompt:**
```batch
build.bat
```

**Output:**
```
bin\stratavore.exe        (CLI)
bin\stratavored.exe       (Daemon)
bin\stratavore-agent.exe  (Agent wrapper)
```

---

### 3. Configure

Edit `configs\stratavore.yaml`:

```yaml
database:
  postgresql:
    host: localhost          # Or your PostgreSQL server
    port: 5432
    database: stratavore_state
    user: stratavore
    password: your_password  # CHANGE THIS!
    sslmode: disable         # Use 'require' in production

docker:
  rabbitmq:
    host: localhost          # Or your RabbitMQ server
    port: 5672
    user: guest
    password: guest
```

---

### 4. Run

**Start Daemon:**
```powershell
# PowerShell
.\bin\stratavored.exe

# Or in background (PowerShell)
Start-Process -NoNewWindow .\bin\stratavored.exe
```

**Use CLI (New Terminal):**
```powershell
# Create a project
.\bin\stratavore.exe new my-project

# Launch a runner
.\bin\stratavore.exe launch my-project

# Check status
.\bin\stratavore.exe status

# List runners
.\bin\stratavore.exe runners

# Stop a runner
.\bin\stratavore.exe kill <runner-id>
```

---

## 🐳 Docker Option (Easiest)

### Use Docker for Infrastructure

**Start services:**
```powershell
docker-compose up -d
```

This starts:
- PostgreSQL
- RabbitMQ
- Redis
- Prometheus
- Grafana
- Qdrant

Then run Stratavore daemon on Windows:
```powershell
.\bin\stratavored.exe
```

---

## 🔧 Configuration

### Environment Variables

**PowerShell:**
```powershell
$env:STRATAVORE_DATABASE_POSTGRESQL_HOST = "192.168.1.100"
$env:STRATAVORE_DATABASE_POSTGRESQL_PASSWORD = "secure_password"
$env:STRATAVORE_DOCKER_TELEGRAM_TOKEN = "bot123456:ABC..."
$env:STRATAVORE_DOCKER_TELEGRAM_CHAT_ID = "123456789"

.\bin\stratavored.exe
```

**Command Prompt:**
```batch
set STRATAVORE_DATABASE_POSTGRESQL_HOST=192.168.1.100
set STRATAVORE_DATABASE_POSTGRESQL_PASSWORD=secure_password

bin\stratavored.exe
```

---

## 🎯 Common Issues

### "Cannot connect to daemon"

**Solution:**
```powershell
# Check if daemon is running
Get-Process stratavored

# If not, start it
.\bin\stratavored.exe
```

### "Database connection failed"

**Check:**
1. PostgreSQL is running
2. Firewall allows port 5432
3. Host/password in config is correct
4. Database `stratavore_state` exists

**Create database:**
```sql
CREATE DATABASE stratavore_state;
CREATE USER stratavore WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE stratavore_state TO stratavore;
```

### "RabbitMQ connection failed"

**Check:**
1. RabbitMQ is running
2. Firewall allows port 5672
3. Guest user can connect (or use custom user)

---

## 📁 File Locations

### Windows Paths

**Config file:**
```
%USERPROFILE%\.config\stratavore\stratavore.yaml
or
configs\stratavore.yaml (current directory)
```

**Binaries:**
```
bin\stratavore.exe
bin\stratavored.exe
bin\stratavore-agent.exe
```

**Logs:**
```
(stdout - redirect to file if needed)
.\bin\stratavored.exe > daemon.log 2>&1
```

---

## 🚀 Advanced: Run as Windows Service

### Using NSSM (Non-Sucking Service Manager)

**1. Download NSSM:**
https://nssm.cc/download

**2. Install service:**
```powershell
nssm install stratavored "C:\path\to\bin\stratavored.exe"
nssm set stratavored AppDirectory "C:\path\to\stratavore"
nssm set stratavored DisplayName "Stratavore Daemon"
nssm set stratavored Description "AI Workspace Orchestrator"
nssm set stratavored Start SERVICE_AUTO_START

nssm start stratavored
```

**3. Manage service:**
```powershell
nssm stop stratavored
nssm restart stratavored
nssm remove stratavored confirm
```

---

## 🔍 Debugging

### Enable Detailed Logging

**PowerShell:**
```powershell
$env:LOG_LEVEL = "debug"
.\bin\stratavored.exe
```

### View Daemon Output

```powershell
# Run in foreground (see all logs)
.\bin\stratavored.exe

# Or redirect to file
.\bin\stratavored.exe > daemon.log 2>&1

# View log file
Get-Content daemon.log -Wait  # Like tail -f
```

### Check API Endpoint

```powershell
# PowerShell
Invoke-WebRequest http://localhost:50051/health

# Or use curl (if installed)
curl http://localhost:50051/health
```

---

## 📊 Monitoring

### Grafana Dashboard

If using Docker Compose:

**Access:** http://localhost:3000  
**Login:** admin / admin

**Available:**
- Runner metrics
- Token usage
- Performance graphs

---

## 🎓 Tips for Windows Users

### Use PowerShell (Not CMD)

PowerShell has better:
- Tab completion
- Error handling
- Colors
- Modern features

### Set Execution Policy

If you get script execution errors:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Add to PATH

**PowerShell (temporary):**
```powershell
$env:PATH += ";C:\path\to\stratavore\bin"
```

**Permanent:**
1. Windows Settings → System → About
2. Advanced system settings
3. Environment Variables
4. Edit PATH
5. Add `C:\path\to\stratavore\bin`

Then just use:
```powershell
stratavore new my-project
```

### Create Shortcuts

**Desktop shortcut to daemon:**
1. Right-click Desktop → New → Shortcut
2. Location: `C:\path\to\stratavore\bin\stratavored.exe`
3. Name: "Stratavore Daemon"

---

## 🐛 Known Windows-Specific Issues

### None Currently! ✅

v1.2 was specifically tested on Windows.

---

## 📞 Getting Help

### Check daemon is running:
```powershell
.\bin\stratavore.exe status
```

### View configuration:
```powershell
Get-Content configs\stratavore.yaml
```

### Test database connection:
```powershell
# From PowerShell with psql installed
$env:PGPASSWORD = "your_password"
psql -h localhost -U stratavore -d stratavore_state -c "\dt"
```

---

## 🎉 You're Ready!

Stratavore works great on Windows. Happy orchestrating!

---

**Need more help?** Check the main README.md and documentation files.
