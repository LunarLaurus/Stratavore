# Protobuf Setup Guide for Stratavore

This guide helps you set up Protocol Buffers (protobuf) for gRPC support in Stratavore.

**Note:** Protobuf is **optional**. Stratavore works with HTTP API by default.

---

## Why Protobuf?

**With Protobuf:**
- Full gRPC support (faster binary protocol)
- Streaming capabilities
- Type-safe API contracts
- Better performance for high-throughput scenarios

**Without Protobuf (HTTP API):**
- ✅ Still fully functional
- ✅ JSON-based REST API
- ✅ Easy debugging
- ✅ Works out of the box

---

## Quick Check

See if you already have the tools:

```bash
# Check protoc
protoc --version
# Should show: libprotoc 3.x.x or higher

# Check Go plugins
protoc-gen-go --version
protoc-gen-go-grpc --version
```

If all three work, you're ready! Just run `make` or `.\build.ps1`

---

## Installation

### Linux (Ubuntu/Debian)

**1. Install protoc:**
```bash
# Download latest release
PB_REL="https://github.com/protocolbuffers/protobuf/releases"
curl -LO $PB_REL/download/v25.1/protoc-25.1-linux-x86_64.zip

# Extract
unzip protoc-25.1-linux-x86_64.zip -d $HOME/.local

# Add to PATH (add to ~/.bashrc)
export PATH="$PATH:$HOME/.local/bin"

# Verify
protoc --version
```

**2. Install Go plugins:**
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Make sure $GOPATH/bin is in PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

**3. Build Stratavore:**
```bash
make
# Protobuf code will be generated automatically
```

---

### macOS

**1. Install protoc (via Homebrew):**
```bash
brew install protobuf

# Verify
protoc --version
```

**2. Install Go plugins:**
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Ensure $(go env GOPATH)/bin is in PATH
```

**3. Build Stratavore:**
```bash
make
```

---

### Windows

**1. Install protoc:**

**Option A: Manual**
1. Download from: https://github.com/protocolbuffers/protobuf/releases
2. Get: `protoc-25.1-win64.zip`
3. Extract to `C:\protoc` (or anywhere)
4. Add `C:\protoc\bin` to System PATH:
   - Windows Settings → System → About
   - Advanced system settings
   - Environment Variables
   - Edit PATH → Add `C:\protoc\bin`

**Option B: Chocolatey**
```powershell
choco install protoc
```

**2. Install Go plugins:**
```powershell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Add Go bin to PATH if not already
$env:PATH += ";$env:USERPROFILE\go\bin"
```

**3. Build Stratavore:**
```powershell
.\build.ps1
# Will auto-detect and generate protobuf code
```

---

## Verification

After installation:

```bash
# All three should work
protoc --version
protoc-gen-go --version  
protoc-gen-go-grpc --version

# Build Stratavore
make          # Linux/Mac
.\build.ps1   # Windows

# You should see:
# [OK] protoc found
# [OK] Go plugins found
# Generating protobuf code...
# [OK] Protobuf code generated
```

---

## Generated Files

When protobuf tools are installed, building creates:

```
pkg/api/generated/
├── stratavore.pb.go         (protobuf types)
└── stratavore_grpc.pb.go    (gRPC service)
```

These files provide:
- Full gRPC support
- Type-safe API definitions
- Streaming capabilities
- Better performance

---

## Fallback Mode

**If protobuf tools are NOT installed:**

Build scripts will show:
```
[WARN] protoc not found - using fallback types
[INFO] Using fallback API types
```

**This is fine!** Stratavore will use:
- `pkg/api/types.go` (hand-written types)
- HTTP REST API (JSON)
- Full functionality (just without gRPC)

---

## Troubleshooting

### "protoc: command not found"

**Check installation:**
```bash
which protoc
echo $PATH
```

**Fix:** Add protoc bin directory to PATH

**Linux/Mac:**
```bash
export PATH="$PATH:$HOME/.local/bin"
# Add to ~/.bashrc or ~/.zshrc
```

**Windows:**
Add `C:\protoc\bin` to System PATH (see installation above)

---

### "protoc-gen-go: program not found"

**Check Go bin in PATH:**
```bash
echo $GOPATH
ls $(go env GOPATH)/bin
```

**Fix:**
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
# Add to shell rc file
```

**Windows:**
```powershell
$env:PATH += ";$env:USERPROFILE\go\bin"
```

---

### "cannot find package" errors when building

**Usually means protobuf tools partially installed.**

**Fix - Clean and rebuild:**
```bash
make clean
make install-proto-tools  # Installs Go plugins
make                      # Regenerates code
```

---

## Advanced: Manual Generation

If you want to generate protobuf code manually:

```bash
# Create output directory
mkdir -p pkg/api/generated

# Generate code
protoc --go_out=pkg/api/generated \
       --go_opt=paths=source_relative \
       --go-grpc_out=pkg/api/generated \
       --go-grpc_opt=paths=source_relative \
       --proto_path=pkg/api \
       pkg/api/stratavore.proto
```

---

## Should I Install Protobuf?

**Install if:**
- You want gRPC support
- You need streaming APIs
- You want maximum performance
- You're building from source for production

**Skip if:**
- You just want to try Stratavore
- HTTP API is sufficient for your use case
- You don't want to install additional tools
- You prefer JSON debugging

**Recommendation:** Start without protobuf, add it later if you need gRPC.

---

## References

- **Protobuf:** https://protobuf.dev/
- **gRPC Go:** https://grpc.io/docs/languages/go/
- **Installation:** https://grpc.io/docs/protoc-installation/

---

## Summary

```bash
# Quick Install (Linux/Mac)
brew install protobuf  # or download binary
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
make

# Quick Install (Windows)
choco install protoc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
.\build.ps1

# Or just build without protobuf - still works!
make    # Will use HTTP API fallback
```

---

**Stratavore works either way. Choose what's best for you!** 🚀
