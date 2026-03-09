# OrigAdmin GitHub Template for Go Projects

This repository contains the default community health files and GitHub Actions workflows for Go projects within the [OrigAdmin](https://github.com/origadmin) organization.

## 📋 What's Included

### 🛠️ GitHub Actions Workflows

All workflows are located in the `.github/workflows` directory and are configured to **not** run in this template repository itself (using an `if` condition checking `github.repository != 'origadmin/go-project-template'`). They will automatically activate when you create a new repository from this template.

- **Test (`test.yml`)** - Runs `go build` and `go test` across multiple platforms (Linux, macOS, Windows) and Go versions.
- **Release (`release.yml`)** - Automates versioning and release creation using [GoReleaser](https://goreleaser.com/) when a new `v*` tag is pushed.
- **CodeQL (`codeql.yml`)** - Performs static code analysis to find security vulnerabilities.
- **Dependabot (`dependabot.yml`)** - Keeps `go.mod` dependencies up-to-date by creating automated pull requests.

### 📝 Community Health Files

- **`CONTRIBUTING.md`** - Guidelines for contributing to projects.
- **`CODE_OF_CONDUCT.md`** - The Contributor Covenant Code of Conduct.
- **`SECURITY.md`** - Instructions for reporting security vulnerabilities.
- **`PULL_REQUEST_TEMPLATE.md`** - A standard template for submitting pull requests.
- **Issue Templates** - Templates for submitting bug reports and feature requests.
- **`.gitignore`** - A simplified `.gitignore` file tailored for Go, GoLand, and VSCode development.
- **`LICENSE`** - MIT License.

## 🚀 Getting Started

1.  Click the "**Use this template**" button at the top of this page to create a new repository.
2.  Clone your new repository and start building your project!
3.  Update the `README.md` and other files (like `SECURITY.md` to include a contact email) with your project-specific information.

## 🤝 Contributing

Contributions to this template are welcome! Please see [CONTRIBUTING.md](.github/CONTRIBUTING.md) for guidelines.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
