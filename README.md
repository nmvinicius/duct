# Duct


![Go](https://img.shields.io/badge/go-1.26+-blue?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)
![Status](https://img.shields.io/badge/status-alpha-orange)

> **Pipeline as Code** — Define, extend and run CI/CD pipelines anywhere.

Define your pipeline once in a `Ductfile`, run it on **GitHub Actions**, **Bitbucket Pipelines**, **GitLab CI**, or **locally** on your machine.

```
┌──────────────┐     ┌─────────────┐     ┌─────────────┐
│  Ductfile    │───▶│  duct run   │───▶│  Any CI/CD  │
│ (declarative)│     │   (engine)  │     │  platform   │
└──────────────┘     └─────────────┘     └─────────────┘
```

---

## 🚀 Quick Start

### 1. Install Duct

```bash
# Install latest stable release
curl -fsSL https://get.duct.dev | sh

# Or build from source
git clone https://github.com/seu-org/duct.git
cd duct && make install
```

### 2. Create a `Ductfile`

```bash
duct init
```

### 3. Run locally

```bash
duct run --local
```

---

## 📄 Ductfile Syntax

```ductfile
VERSION 1.0

PROJECT my-api
TEAM backend

# Global variables
GLOBAL NODE_VERSION=20
GLOBAL REGISTRY=myregistry.io

# Inherit from a template
EXTENDS github.com/seu-org/duct-templates//node/base

# Required secrets
REQUIRE SECRET AWS_ACCESS_KEY_ID

# Pipeline steps
STEP lint
    USE node
    RUN npm ci
    RUN npm run lint

STEP test
    USE node
    NEEDS lint
    RUN npm test
    ARTIFACTS coverage/

STEP build
    USE docker
    NEEDS test
    RUN docker build -t my-api:$GIT_COMMIT .

STEP deploy
    USE kubectl
    NEEDS build
    WHEN branch == "main"
    RUN kubectl apply -f k8s/
    ROLLBACK kubectl rollout undo deployment/my-api
```

---

## 🖥️ Local Development

Test your pipeline before pushing:

```bash
duct validate              # Validate Ductfile syntax
duct run --local           # Run full pipeline locally
duct run --local --dry-run # See what would execute
duct graph                 # Show dependency graph
duct run step build        # Run single step
```

---

## 🔷 GitHub Actions

```yaml
# .github/workflows/duct.yml
name: Pipeline
on: [push, pull_request]

jobs:
  duct:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: |
          git clone --branch stable --depth 1 https://github.com/seu-org/duct.git .duct
          sh .duct/bin/duct run
```

---

## 🔶 Bitbucket Pipelines

```yaml
# bitbucket-pipelines.yml
definitions:
  caches:
    duct: .duct

pipelines:
  default:
    - step:
        name: Pipeline
        caches: [duct]
        script:
          - "[ -d .duct ] || git clone --branch stable --depth 1 https://bitbucket.org/seu-org/duct.git .duct"
          - "sh .duct/bin/duct run"
```

---

## 🦊 GitLab CI

```yaml
# .gitlab-ci.yml
run-duct:
  image: alpine/git:latest
  cache:
    key: duct
    paths: [.duct/]
  before_script:
    - "[ -d .duct ] || git clone --branch stable --depth 1 https://gitlab.com/seu-org/duct.git .duct"
  script:
    - sh .duct/bin/duct run
```

---

## 🧩 Templates

Extend pre-built templates to avoid boilerplate:

```ductfile
EXTENDS github.com/seu-org/duct-templates//node/docker

# Add your custom steps
STEP deploy
    USE kubectl
    RUN kubectl apply -f k8s/
```

---

## 🔧 CLI Reference

| Command | Description |
|---------|-------------|
| `duct run` | Run full pipeline |
| `duct run --local` | Run locally (simulates CI) |
| `duct run --dry-run` | Show what would run |
| `duct run step <name>` | Run specific step |
| `duct validate` | Validate Ductfile syntax |
| `duct init` | Create Ductfile interactively |
| `duct graph` | Show dependency graph |
| `duct version` | Show version |

---

## 🏗️ Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Ductfile   │───▶│  duct run   │───▶│   Engine    │───▶│   Runner    │
│(declarative)│     │   (CLI)     │     │  (Go+Shell) │     │ (executes)  │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                                              │
                                              ▼
                                       ┌─────────────┐
                                       │  Platform   │
                                       │  Adapter    │
                                       │ (GA/BB/GL)  │
                                       └─────────────┘
```