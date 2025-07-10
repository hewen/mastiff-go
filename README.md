# mastiff-go

**mastiff-go** is a comprehensive and extensible Go toolkit for building robust and modular server-side applications. It provides a clean project structure, built-in support for HTTP/gRPC servers, logging, storage layers, message queues, and includes a powerful CLI tool (`mastiffgen`) for code generation.

<p align="left">
  <a href="https://github.com/hewen/mastiff-go/actions?query=workflow%3ATests" title="Build Status"><img src="https://img.shields.io/github/actions/workflow/status/hewen/mastiff-go/test.yml?branch=dev&style=flat-square&logo=github-actions" /></a>
  <a href="https://codecov.io/gh/hewen/mastiff-go" title="Codecov"><img src="https://img.shields.io/codecov/c/github/hewen/mastiff-go?style=flat-square&logo=codecov" /></a>
  <a href="https://github.com/hewen/mastiff-go" title="Supported Platforms"><img src="https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20*BSD%20%7C%20Windows-549688?style=flat-square&logo=launchpad" /></a>
  <a href="https://github.com/hewen/mastiff-go" title="Minimum Go Version"><img src="https://img.shields.io/badge/go-%3E%3D1.24-30dff3?style=flat-square&logo=go" /></a>
  <br />
  <a href="https://goreportcard.com/report/github.com/hewen/mastiff-go" title="Go Report Card"><img src="https://goreportcard.com/badge/github.com/hewen/mastiff-go?style=flat-square" /></a>
  <a href="https://pkg.go.dev/github.com/hewen/mastiff-go#section-documentation" title="Documentation"><img src="https://img.shields.io/badge/go.dev-doc-007d9c?style=flat-square&logo=read-the-docs" /></a>
  <a href="https://github.com/hewen/mastiff-go/releases" title="Releases"><img src="https://img.shields.io/github/v/release/hewen/mastiff-go.svg?color=161823&style=flat-square&logo=smartthings" /></a>
  <a href="https://github.com/hewen/mastiff-go/tags" title="Tags"><img src="https://img.shields.io/github/v/tag/hewen/mastiff-go?color=%23ff8936&logo=fitbit&style=flat-square" /></a>
</p>

---

## 📦 Installation

```bash
go get github.com/hewen/mastiff-go
````

Install the `mastiffgen` CLI tool:

```bash
go install github.com/hewen/mastiff-go/cmd/mastiffgen@latest
```

---

## 🚀 Quick Start

### 1. Scaffold a new project

```bash
mastiffgen init --package github.com/yourname/myproject --project myproject --dir ./myproject
```

This command creates a new project with a clean architecture under the `./myproject` directory.

### 2. Add a module

```bash
mastiffgen module user --package github.com/yourname/myproject --dir ./myproject
```

This generates the module code under `core/user/`, and automatically updates:

* Import paths
* Struct fields
* Initialization code
* Route registration in `core/core.go`

---

## 🧱 Example Scaffolded Project

```text
myproject/
├── core/
│   ├── core.go         # Module registration logic
│   └── user/           # Example generated module
├── pkg/
│   └── model/          # Database models
├── config/             # Configuration handling
├── main.go             # Application entry point
```

---

## 🏗️ Project Structure Overview

```text
.
├── cmd/mastiffgen        # CLI tool for project/module scaffolding
├── logger/               # Logging system abstraction
├── server/               # HTTP/gRPC server setup and queue handler
├── store/                # Storage layer abstraction (MySQL, Redis)
├── util/                 # Common utility functions (port, time, etc.)
├── go.mod / go.sum       # Go module definition and dependencies
├── Makefile              # Build and run shortcuts
└── README.md             # Project documentation
```

---

## 🔧 CLI Tools

* **mastiffgen**: Powerful code generator for:

  * Project initialization
  * Module scaffolding
  * Template-based file generation and injection

---

## 📋 Requirements

* Go 1.24 or later
* Redis (for queue handling)

---

## 🧪 Testing

Run the full test suite:

```bash
make test
```

---

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch: `git checkout -b feature/awesome-feature`
3. Commit your changes: `git commit -m 'Add awesome feature'`
4. Push to the branch: `git push origin feature/awesome-feature`
5. Open a Pull Request 🚀

---

## 📄 License

This project is licensed under the Apache‑2.0. See [LICENSE](LICENSE) for details.

---

## 📚 Support & Resources

* 📖 [API Documentation](https://pkg.go.dev/github.com/hewen/mastiff-go)
* 🐛 [Issue Tracker](https://github.com/hewen/mastiff-go/issues)
* 💬 [Discussions](https://github.com/hewen/mastiff-go/discussions)

---

Made with ❤️ by the mastiff-go team
