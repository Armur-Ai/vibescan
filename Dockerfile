# ============================================================
# Stage 1: Go binary builder
# ============================================================
FROM golang:1.27-alpine AS go-builder

WORKDIR /build
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /armur-server ./cmd/server/main.go

# ============================================================
# Stage 2: Go security tools
# ============================================================
FROM golang:1.27-alpine AS go-tools

RUN apk add --no-cache git
ENV GOBIN=/go-tools
RUN mkdir -p /go-tools

RUN go install github.com/securego/gosec/v2/cmd/gosec@latest && \
    go install golang.org/x/lint/golint@latest && \
    go install honnef.co/go/tools/cmd/staticcheck@latest && \
    go install github.com/fzipp/gocyclo/cmd/gocyclo@latest && \
    go install golang.org/x/tools/cmd/deadcode@latest && \
    go install github.com/google/osv-scanner/cmd/osv-scanner@latest

# ============================================================
# Stage 3: Python security tools
# ============================================================
FROM python:3.12-slim AS python-tools

RUN pip install --no-cache-dir \
    semgrep bandit pydocstyle radon pylint checkov vulture

# ============================================================
# Runtime target: armur:go  (Go tools only)
# docker build --target armur-go -t armur:go .
# ============================================================
FROM debian:bookworm-slim AS armur-go

RUN apt-get update && apt-get install -y --no-install-recommends \
    git ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=go-builder /armur-server  /usr/local/bin/armur-server
COPY --from=go-tools   /go-tools      /usr/local/bin/
COPY . /armur
WORKDIR /armur
ENV ARMUR_REPOS_DIR=/armur/repos
RUN mkdir -p /armur/repos
EXPOSE 4500
CMD ["/usr/local/bin/armur-server"]

# ============================================================
# Runtime target: armur:python  (Python tools only)
# docker build --target armur-py -t armur:python .
# ============================================================
FROM python:3.12-slim AS armur-py

RUN apt-get update && apt-get install -y --no-install-recommends \
    git ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=go-builder   /armur-server /usr/local/bin/armur-server
COPY --from=python-tools /usr/local    /usr/local
COPY . /armur
WORKDIR /armur
ENV ARMUR_REPOS_DIR=/armur/repos
RUN mkdir -p /armur/repos
EXPOSE 4500
CMD ["/usr/local/bin/armur-server"]

# ============================================================
# Runtime target: armur:js  (JavaScript/TypeScript tools only)
# docker build --target armur-js -t armur:js .
# ============================================================
FROM node:22-slim AS armur-js

RUN apt-get update && apt-get install -y --no-install-recommends \
    git ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=go-builder /armur-server /usr/local/bin/armur-server
RUN npm install -g eslint jscpd
COPY . /armur
WORKDIR /armur
RUN npm install @eslint/js eslint-plugin-jsdoc eslint-plugin-security
COPY rule_config/eslint/eslint.config.js           /armur/eslint.config.js
COPY rule_config/eslint/eslint_jsdoc.config.js     /armur/eslint_jsdoc.config.js
COPY rule_config/eslint/eslint_security.config.js  /armur/eslint_security.config.js
COPY rule_config/eslint/eslint_deadcode.config.js  /armur/eslint_deadcode.config.js
ENV ARMUR_REPOS_DIR=/armur/repos
RUN mkdir -p /armur/repos
EXPOSE 4500
CMD ["/usr/local/bin/armur-server"]

# ============================================================
# Runtime target: armur:full  (all tools — DEFAULT)
# docker build -t armur:full .
# ============================================================
FROM python:3.12-slim AS armur-full

RUN apt-get update && apt-get install -y --no-install-recommends \
    git ca-certificates curl build-essential gcc \
    && rm -rf /var/lib/apt/lists/*

# Node.js
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Go binary + Go tools
COPY --from=go-builder /armur-server /usr/local/bin/armur-server
COPY --from=go-tools   /go-tools     /usr/local/bin/

# Python tools
COPY --from=python-tools /usr/local /usr/local

# Trivy
RUN curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh \
    | sh -s -- -b /usr/local/bin

# Node tools
RUN npm install -g eslint jscpd

COPY . /armur
WORKDIR /armur
RUN npm install @eslint/js eslint-plugin-jsdoc eslint-plugin-security

COPY rule_config/eslint/eslint.config.js           /armur/eslint.config.js
COPY rule_config/eslint/eslint_jsdoc.config.js     /armur/eslint_jsdoc.config.js
COPY rule_config/eslint/eslint_security.config.js  /armur/eslint_security.config.js
COPY rule_config/eslint/eslint_deadcode.config.js  /armur/eslint_deadcode.config.js

ENV ARMUR_REPOS_DIR=/armur/repos
RUN mkdir -p /armur/repos
EXPOSE 4500
CMD ["/usr/local/bin/armur-server"]
