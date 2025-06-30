FROM python:3.11-slim AS base

# Instala dependencias necesarias y Go
RUN apt-get update && apt-get install -y \
    curl \
    git \
    build-essential \
    wget \
    gnupg \
    ca-certificates \
    && wget https://go.dev/dl/go1.24.3.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.24.3.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:$PATH"

# Instala GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | \
      gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] \
         https://cli.github.com/packages stable main" | \
         tee /etc/apt/sources.list.d/github-cli.list > /dev/null && \
    apt-get update && apt-get install -y gh

# Instala Poetry
ENV POETRY_HOME="/opt/poetry"
RUN curl -sSL https://install.python-poetry.org | python3 - && \
    ln -s $POETRY_HOME/bin/poetry /usr/local/bin/poetry

WORKDIR /app
COPY . .

# Compila el binario Go
RUN go build -o sidecar .

ENTRYPOINT ["/app/sidecar"]
