# Indie - Docker command execution

Indie is a small binary which allows you to execute a command inside a container. Its main usage is to implement
the pattern of build container for your project without having to create a Makefile or something else.

## Installation

Simply download the binary from GitHub and add it to your `/usr/local/bin` directory.

```shell
$ curl -L "https://github.com/jlrigau/indie/releases/download/v0.1.0/indie-$(uname -s)-$(uname -m)" -o /usr/local/bin/indie && \
  chmod +x /usr/local/bin/indie
```

## Build

You can build Indie using Indie with one command.
 
```shell
$ indie go build
```