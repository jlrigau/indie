# Indie - Docker command execution

Indie is a small binary which allows you to execute a command inside a container. Its main usage is to implement
the pattern of build container for your project without having to create a Makefile or something else.

## Installation

Simply download the binary from GitHub and add it to your `/usr/local/bin` directory.

```shell
$ curl -L "https://github.com/jlrigau/indie/releases/download/v0.1.0/indie-$(uname -s)-$(uname -m)" -o /usr/local/bin/indie && \
  chmod +x /usr/local/bin/indie
```

## Usage

Add a YAML file named `indie.yml` in your repository in which you have to indicate the image to use for building
your application.

For example, if you want to build a Go application using the official Go image, you have to add this line
in the configuration file.

```yaml
image: golang:latest
```

For building a Go application it is important to well manage your GOPATH and you can do that by using a
specific workspace by adding a new property in your configuration file.

```yaml
workspace: /go/src/github.com/jlrigau/indie
```

If you need to add specific environment variable to your build container you have also the ability to add them to the
configuration file using an `environment` bloc.

```yaml
environment:
  - "GOOS=linux"
  - "GOARCH=amd64"
```

## Build

You can build Indie by using Indie of course.
 
```shell
$ indie go build
```