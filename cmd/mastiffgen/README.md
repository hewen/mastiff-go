---

### 📄 `README.md`

````markdown
# MastiffGen - Project Code Generator

MastiffGen is a command-line tool written in Go that helps bootstrap Go projects with a clean and modular structure. It supports initialization of a project scaffold as well as adding new modules.

## ✨ Features

- 📦 Initialize a new Go project (`init` command)
- 🛠 Template-based file generation using `embed.FS`
- 🔧 Easy integration with your existing project structure

---

## 🚀 Getting Started

### Installation

Install the `mastiffgen` CLI tool:

```bash
go install github.com/hewen/mastiff-go/cmd/mastiffgen@latest
```

---

## 🧪 Usage

### Initialize a new project

```bash
mastiffgen init --package github.com/yourname/myproject --project myproject --dir ./myproject
```

This will generate the base project structure in the `./myproject` directory.

---

## 🧱 Project Structure

```
myproject/
├── cmd/
│   ├── root.go                   # Cobra root command
│   └── run.go                    # command that calls di.InitApp(), starts the service, loads config, and launches HTTP/gRPC services
├── internal/
│   ├── di/
│   │   └── init.go               # Generated dependency graph
│   ├── core/
│   │   ├── domain/               # Entities + domain services
│   │   ├── usecase/              # Business logic orchestration
│   │   └── interfaces/           # Adapter implementations
│   │       ├── http/             # HTTP adapter
│   │       │    ├── handler/     # Handlers (grouped by functionality)
│   │       │    ├── route.go     # Route definitions
│   │       │    └── server.go    # HTTP server creation
│   │       ├── rpc/              # RPC adapter
│   │       │    ├── handler/     # Handlers (grouped by functionality)
│   │       │    └── server.go    # gRPC server creation
│   │       ├── queue/            # Queue adapter
│   │       │    ├── handler/     # Handlers (grouped by functionality)
│   │       │    └── server.go    # Queue server creation
│   │       ├── socket/           # Socket adapter
│   │       │    ├── handler/     # Handlers (grouped by functionality)
│   │       │    └── server.go    # Socket server creation
│   │       ├── websocket/        # WebSocket adapter
│   │       │    ├── handler/     # Handlers (grouped by functionality)
│   │       │    └── server.go    # WebSocket server creation
│   │       └── repository/       # MySQL/Redis implementations
│   │            ├── sqlc/        # sqlc-generated code
│   │            ├── custom/      # Custom repository code
│   │            └── sql/         # SQL definitions for sqlc
│   │                 ├── schema  # Table schema definitions
│   │                 └── queries # Query definitions
│   ├── config/                   # Configuration loading (via viper/env, etc.)
│   └── pkg/                      # Shared internal libraries (logger, utils, etc.)
│        └── constants/           # Global constants
├── sqlc.yaml                     # Configuration for sqlc code generation
├── go.mod
├── .githooks
├── .gitignore
├── .golangci.yml
├── Dockerfile
├── Makefile
├── README.md
└── main.go                       # Only calls cmd.Execute()

```

---

## 📂 Templates

Template files are stored under:

```
templates/
└── init/       # Used for project scaffolding
```

They are loaded using Go's `embed.FS` and rendered via Go's `text/template`.
