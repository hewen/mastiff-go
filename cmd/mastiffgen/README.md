---

### 📄 `README.md`

````markdown
# MastiffGen - Project Code Generator

MastiffGen is a command-line tool written in Go that helps bootstrap Go projects with a clean and modular structure. It supports initialization of a project scaffold as well as adding new modules.

## ✨ Features

- 📦 Initialize a new Go project (`init` command)
- 🧩 Add new modules with HTTP/gRPC support (`module` command)
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

### 1. Initialize a new project

```bash
mastiffgen init --package github.com/yourname/myproject --project myproject --dir ./myproject
```

This will generate the base project structure in the `./myproject` directory.

### 2. Add a new module

```bash
mastiffgen module user --package github.com/yourname/myproject --dir ./myproject
```

This will generate module code under `core/user/` and automatically inject required lines into `core/core.go`:

- Fields
- Initializations
- Route registration
- Import statement

---

## 🧱 Project Structure

```
myproject/
├── core/
│   ├── core.go         # Main registration logic for modules
│   └── user/           # Example module
├── pkg/
│   └── model/          # Database models
├── config/             # Configuration logic
├── main.go             # Application entry point
```

---

## 🔖 Template Markers

`core/core.go` is modified using the following comment markers:

- `// MODULE_PACKAGE_START` / `// MODULE_PACKAGE_END`
- `// MODULE_FIELDS_START` / `// MODULE_FIELDS_END`
- `// MODULE_INITS_START` / `// MODULE_INITS_END`
- `// MODULE_ROUTES_START` / `// MODULE_ROUTES_END`

These markers indicate where new module code should be injected.

---

## 📂 Templates

Template files are stored under:

```
templates/
├── init/       # Used for project scaffolding
└── module/     # Used for module generation
```

They are loaded using Go's `embed.FS` and rendered via Go's `text/template`.
