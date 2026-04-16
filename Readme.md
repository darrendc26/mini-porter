# mini-porter — CLI PaaS for Kubernetes

## Demo

![demo](examples/node-app/porter-demo.gif)

---

## About

mini-porter is a minimal Platform-as-a-Service (PaaS) CLI that lets you deploy applications to Kubernetes with a single command.

To deploy your app, run:
```bash
mini-porter deploy
```

It handles:

* Docker image build
* Image push (optional)
* Kubernetes deployment
* Service exposure
* Ingress + domain setup

---

## Features

* mini-porter.yaml file generation (`mini-porter init`)
* One-command deploy (`mini-porter deploy`)
* Docker build & push
* Kubernetes Deployment & Service creation
* Ingress with custom domain (`.miniporter`)
* Status inspection (`mini-porter status`)
* Clean teardown (`mini-porter delete`)
* Clean architecture (CLI → Orchestrator → Infra modules)

---

## Architecture

```
CLI (Cobra)
   ↓
Orchestrator (deploy logic)
   ↓
--------------------------------
| Docker | Kubernetes | Config |
--------------------------------
```

* **cmd/** → CLI commands
* **internal/** → core logic (Docker, K8s, deploy orchestration)
* **examples/** → sample app for testing

---

## Installation

### Option 1 — Build locally

```bash
git clone https://github.com/darrendc26/mini-porter.git
cd mini-porter
go build -o mini-porter
sudo mv mini-porter /usr/local/bin/
```

---

### Option 2 — Go install

```bash
go install github.com/darrendc26/mini-porter@latest
```

---

## Prerequisites

Ensure the following are installed and running before using mini-porter:

* Docker (for building and pushing container images)
* Kubernetes cluster (e.g., Minikube or kind)
* kubectl configured to access your cluster

---

### Verify setup

```bash
docker --version
kubectl get nodes
```

---

### Start a local cluster (if needed)

Using Minikube:

```bash
minikube start
minikube addons enable ingress
```

Or using kind:

```bash
kind create cluster
```
---

## Requirements

To deploy an application, your project must include:
Literally nothing... Just a working application is all that's needed.

---

### Example `porter.yaml`

```yaml
name: my-app
image: yourdockerhubusername/my-app
port: 3000
replicas: 1
services:
  - name: my-app
    path: .
    port: 3000
```

---

## Example App Usage

### 1. Navigate to your app

```bash
cd examples/node-app
# Generates a mini-porter.yaml file
mini-porter init
```

---

### 2. Deploy

```bash
mini-porter deploy
```

---

### 3. Access your app

After deployment:

```text
http://my-app.local
```

If not configured yet, run:

```bash
mini-porter host add
```

---

### 4. Check status

```bash
mini-porter status
```

---

### 5. Delete deployment

```bash
# Deletes the entire app deployment
mini-porter delete

# To delete a single service
mini-porter delete <service-name>
```

---

## How it works

mini-porter performs:

1. Builds Docker image
2. Pushes image to registry
3. Creates Kubernetes Deployment
4. Exposes Service (NodePort)
5. Creates Ingress for domain routing
6. Optionally adds local DNS entry
7. Sets up postgres and redis databases (if needed)

---

## Design Decisions

* **NodePort over LoadBalancer**
  → Simpler local development with Minikube

* **YAML config (****`mini-porter.yaml`****)**
  → Explicit and predictable deployments

* **No Helm**
  → Full control over Kubernetes resources

* **Orchestrator pattern**
  → Clean separation between CLI and infra logic

---

## Example App

A sample Node.js app is included:

```bash
examples/node-app/
```

Use it to test the full deployment flow.

---

## Future Improvements

* Ingress auto-setup (no manual hosts entry)
* HTTPS support (cert-manager)
* Autoscaling (HPA)
* Remote cluster support (AWS/GCP)
* Web dashboard

---

## Why this project

mini-porter was built to understand and demonstrate:

* Containerization workflows
* Kubernetes resource management
* Infrastructure orchestration
* CLI design patterns in Go

---

## License

MIT License
