FROM python:3.11-slim AS base

# Instala Go
RUN apt-get update && apt-get install -y curl git build-essential wget && \
    wget https://go.dev/dl/go1.24.3.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.24.3.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:$PATH"

# Instala Poetry
ENV POETRY_HOME="/opt/poetry"
RUN curl -sSL https://install.python-poetry.org | python3 - && \
    ln -s $POETRY_HOME/bin/poetry /usr/local/bin/poetry

WORKDIR /app
COPY . .

RUN go build -o app .

ENTRYPOINT ["/app"]
