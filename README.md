# TCVM - Tencent Cloud VM Manager

A CLI tool for managing Tencent Cloud CVM (Cloud Virtual Machine) instances through APIs, designed especially for **instances without public network access**.

## Features

- **Interactive Workbench** - Run `./tcvm` to enter a visual menu-driven interface
- **Instance Management** - List, connect, execute commands, and upload files to CVM instances
- **Pseudo-Shell** - Interactive command-line experience via Tencent Automation Tools (TAT)
- **File Transfer**
  - Small files (<= 24KB): Direct transfer via TAT
  - Large files: Automatic COS (Cloud Object Storage) relay transfer with MD5 verification
- **VNC Connection** - Retrieve VNC URLs for graphical access
- **Command History** - Query execution results by invocation ID
- **Multi-Region Support** - Override region via CLI flags

## Installation

### Prerequisites

- Go 1.23+ (for building from source)
- Tencent Cloud API credentials (SecretId / SecretKey)
- (Optional) COS bucket for large file transfers

### Build from Source

```bash
# Clone the repository
git clone <repo-url>
cd tcvm

# Build for current platform
go build -o tcvm .

# Or cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o tcvm .
```

### Quick Start

```bash
# 1. Configure credentials (interactive)
./tcvm config
# Or manually create ~/.tcvm/config.yaml based on config.yaml.example

# 2. Launch interactive workbench
./tcvm

# 3. Or use subcommands directly
./tcvm list
./tcvm connect ins-xxxxxx
./tcvm exec ins-xxxxxx "uname -a"
./tcvm upload ins-xxxxxx ./local.jar /opt/app/app.jar
```

## Security Notes

- **Credentials are sensitive**: Your Tencent Cloud SecretId/SecretKey grant access to your cloud resources. Never commit them to version control.
- **File permissions**: `~/.tcvm/config.yaml` is created with `0600` permissions (owner read/write only).
- **Environment variables preferred**: For CI/CD or shared machines, use `TCVM_SECRET_ID` and `TCVM_SECRET_KEY` environment variables instead of the config file.
- **Example config provided**: See `config.yaml.example` for the required format.

## Configuration

Configuration is stored in `~/.tcvm/config.yaml` (created automatically on first run):

```yaml
secret_id: <YOUR_TENCENT_SECRET_ID>
secret_key: xxxxxxxxxxxxxxxx
region: ap-guangzhou
cos_bucket: mybucket-1250000000    # Optional, for large file transfers
cos_region: ap-guangzhou            # Optional, COS bucket region
```

### Environment Variables

You can also use environment variables (highest priority):

```bash
export TCVM_SECRET_ID=<YOUR_TENCENT_SECRET_ID>
export TCVM_SECRET_KEY=xxx
export TCVM_REGION=ap-guangzhou
export TCVM_COS_BUCKET=mybucket-1250000000
export TCVM_COS_REGION=ap-guangzhou
```

## Usage

### Interactive Mode (Recommended)

```bash
./tcvm
```

Opens the main menu:

```
TCVM - Tencent Cloud VM Manager

[1] List instances
[2] Connect to instance
[3] Execute command
[4] Upload file
[5] View tasks
[6] Configure
[h] Help
[q] Quit
```

### Subcommands

| Command | Description | Example |
|---------|-------------|---------|
| `list` | List all CVM instances | `./tcvm list` |
| `connect` | Connect to instance interactively | `./tcvm connect ins-xxx` |
| `exec` | Execute a command via TAT | `./tcvm exec ins-xxx "df -h"` |
| `upload` | Upload a file | `./tcvm upload ins-xxx ./file.txt /tmp/file.txt` |
| `tasks` | Query command execution results | `./tcvm tasks inv-xxx` |
| `config` | Configure credentials | `./tcvm config` |
| `shell` | Start interactive shell (alias) | `./tcvm shell ins-xxx` |

### Inside the Shell

When connected to an instance, you can:

```bash
ins-xxx:/root$ ls -la
ins-xxx:/root$ cd /tmp
ins-xxx:/tmp$ cat /etc/os-release
ins-xxx:/tmp$ upfile ./local.txt /remote.txt   # Upload file from local to remote
ins-xxx:/tmp$ exit                              # Disconnect
```

**Special Commands:**
- `upfile <local-path> <remote-path>` - Upload file from local machine to the connected instance
- `exit` / `quit` - Disconnect

### File Transfer Modes

**Small files (<= 24KB):**
Directly encoded as base64 and sent through TAT command channel.

**Large files (> 24KB):**
Requires COS bucket configuration. The transfer process:
1. Upload file to COS bucket
2. Generate a presigned download URL (5-minute expiry)
3. Execute `wget` / `curl` on the instance to download
4. Verify file integrity via MD5 checksum
5. Clean up COS temporary object

## Architecture

```
tcvm/
├── cmd/                  # CLI commands (cobra)
│   ├── root.go          # Main menu & workbench
│   ├── connect.go       # Interactive shell
│   ├── exec.go          # One-shot command execution
│   ├── upload.go        # File upload
│   ├── list.go          # Instance listing
│   ├── tasks.go         # Task query
│   └── config.go        # Credential configuration
├── internal/
│   ├── config/          # Configuration management
│   └── tencent/         # Tencent Cloud SDK wrappers
│       ├── client.go    # Base client
│       ├── cvm.go       # CVM API (DescribeInstances, VNC)
│       ├── tat.go       # TAT API (RunCommand, UploadFile)
│       └── cos.go       # COS API (Upload, PresignedURL)
└── main.go
```

## Limitations

- **Not a real SSH session**: Each command is an independent API call via TAT, with ~1-3s latency
- **No interactive programs**: `vim`, `top`, `htop`, password prompts are not supported
- **Environment state**: Environment variables do not persist between commands (TAT limitation)
- **File size limit**: Direct TAT upload limited to ~24KB. Larger files require COS relay
- **TAT Agent required**: Instances must have Tencent Automation Tools agent installed (pre-installed on official images)

## License

MIT
