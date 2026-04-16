package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/darrendc26/mini-porter/internal/detector"
)

func CreateDockerfile(ctx context.Context, project detector.Project, port int) error {

	var dockerfileContent string

	dockerfilePath := filepath.Join(project.Path, "Dockerfile")

	if _, err := os.Stat(dockerfilePath); err == nil {
		fmt.Println("[1/6] Dockerfile                                Already Exists")
		return nil
	}

	switch project.Type {
	case "nodejs":
		dockerfileContent = fmt.Sprintf(`FROM node:18-alpine 
WORKDIR /app
COPY package*.json ./
RUN npm install --omit=dev

COPY . .
EXPOSE %d
CMD ["npm", "start"]`, port)
		WriteDockerfile(project.Path, dockerfileContent)

	case "python":
		dockerfileContent = fmt.Sprintf(`FROM python:3.11-slim AS builder
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

FROM python:3.11-slim
WORKDIR /app
COPY --from=builder /usr/local/lib/python3.11 /usr/local/lib/python3.11
COPY . .
EXPOSE %d
CMD ["python", "app.py"]`, port)
		WriteDockerfile(project.Path, dockerfileContent)

	case "golang":
		dockerfileContent = fmt.Sprintf(`FROM golang:latest AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo -o app .
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
EXPOSE %d
CMD ["/app/app"]`, port)
		WriteDockerfile(project.Path, dockerfileContent)

	case "java":
		dockerfileContent = fmt.Sprintf(`FROM maven:3.9-eclipse-temurin-17 AS builder

WORKDIR /app
COPY pom.xml .
RUN mvn dependency:go-offline
COPY . .
RUN mvn package -DskipTests

FROM eclipse-temurin:17-jdk-jammy
WORKDIR /app
COPY --from=builder /app/target/*.jar app.jar
EXPOSE %d
CMD ["java", "-jar", "app.jar"]`, port)
		WriteDockerfile(project.Path, dockerfileContent)

	case "rust":
		dockerfileContent = fmt.Sprintf(`FROM rust:1.75 AS builder

WORKDIR /app
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main(){}" > src/main.rs
RUN cargo build --release
COPY . .
RUN cargo build --release

FROM debian:buster-slim
WORKDIR /app
COPY --from=builder /app/target/release/app .
EXPOSE %d
CMD ["./app"]`, port)
		WriteDockerfile(project.Path, dockerfileContent)

	default:
		return fmt.Errorf("unsupported project type: %s", project.Type)
	}
	fmt.Println("[1/6] Dockerfile                                Created")
	return nil
}

func WriteDockerfile(path string, dockerfileContent string) error {
	fmt.Println("[1/6] Dockerfile                                Generating")
	filePath := filepath.Join(path, "Dockerfile")
	if err := os.WriteFile(filePath, []byte(dockerfileContent), 0644); err != nil {
		return fmt.Errorf("Failed to write Dockerfile: %w", err)
	}
	return nil
}
