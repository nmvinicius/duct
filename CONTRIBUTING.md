# Contributing to Duct

Thanks for your interest in contributing!

## Development Setup

```bash
git clone https://github.com/nmvinicius/duct.git
cd duct
go mod tidy
make build
```

## Running Tests

```bash
make test              # Run all tests
make test-coverage     # With coverage report
make lint              # Format and vet
```

## Project Structure

|          Path          |               Purpose               |
|:-----------------------|:------------------------------------|
|`cmd/duct/`             | CLI entrypoint                      |
|`pkg/ductfile/`         | Core types and logic                |
|`internal/parser/`      | Ductfile parser                     |
|`internal/executor/`    | Step execution engine               |
|`internal/platform/`    | CI platform detection               |
|`scripts/`              | Shell scripts for tools and runners |

## Submitting Changes

1. Clone the repo
2. Create a branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Add tests
5. Run `make test lint`
6. Submit a PR

## Code Style

- Go: standard `gofmt`
- Shell: `shfmt` if available
- Commit messages: conventional commits preferred
```

---

## .gitignore (bônus)

```gitignore
# Build
/build/
/dist/
*.exe

# Go
*.out
coverage.html
coverage.out
vendor/

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Test artifacts
/tmp/
*.log
```

---

## Makefile target extra: `make ci`

Adiciona isso no seu Makefile:

```makefile
.PHONY: ci
ci: lint test build ## Run all CI checks locally
	@echo "$(GREEN)All checks passed!$(RESET)"
```

Pronto! Substitua `[Your Name]` no LICENSE pelo seu nome/organização e pode publicar. 🚀