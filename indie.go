package main

import (
	"os"
	"github.com/jawher/mow.cli"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"io/ioutil"
	"github.com/fsouza/go-dockerclient"
	"log"
	"strings"
	"github.com/satori/go.uuid"
)

type Config struct {
	Image       string
	Workspace   string
	Environment [] string
}

type DockerClient struct {
	client     *docker.Client
	authConfig docker.AuthConfiguration
}

func main() {
	app := cli.App("indie", "Docker command execution")

	app.Spec = "[OPTIONS] CMD ARG..."
	app.Version("v version", "indie 0.2.0")

	var (
		endpoint = app.String(cli.StringOpt{
			Name:   "docker-endpoint",
			Value:  "unix:///var/run/docker.sock",
			EnvVar: "DOCKER_HOST",
			Desc:   "Docker host or socket",
		})

		registry = app.String(cli.StringOpt{
			Name:   "registry-host",
			Value:  "docker.io",
			EnvVar: "DOCKER_REGISTRY_HOST",
			Desc:   "Docker registry host",
		})

		username = app.String(cli.StringOpt{
			Name:   "username",
			EnvVar: "DOCKER_USERNAME",
			Desc:   "Username to log in to a Docker registry",
		})
		password = app.String(cli.StringOpt{
			Name:      "password",
			EnvVar:    "DOCKER_PASSWORD",
			Desc:      "Password to log in to a Docker registry",
			HideValue: true,
		})

		cmd  = app.StringsArg("CMD", nil, "Command")
		args = app.StringArg("ARG", "", "Arguments")
	)

	app.Action = func() {
		config := configFromFile()
		dockerClient := configureDockerClient(*endpoint, *username, *password, *registry)

		repository := config.Image
		tag := ""

		if strings.Contains(config.Image, ":") {
			repositoryWithTag := strings.Split(config.Image, ":")

			repository = repositoryWithTag[0]
			tag = repositoryWithTag[1]
		}

		if err := dockerClient.client.PullImage(docker.PullImageOptions{
			Repository: repository,
			Tag:        tag,
		}, dockerClient.authConfig); err != nil {
			log.Fatal(err)
		}

		dir, _ := os.Getwd()
		name := "indie-" + uuid.NewV4().String()

		container, err := dockerClient.client.CreateContainer(docker.CreateContainerOptions{
			Name: name,
			Config: &docker.Config{
				Image:        config.Image,
				WorkingDir:   config.Workspace,
				Env:          config.Environment,
				Cmd:          append(*cmd, *args),
				AttachStdout: true,
				AttachStderr: true,
				AttachStdin:  true,
				Tty:          true,
				OpenStdin:    true,
				StdinOnce:    true,
			},
			HostConfig: &docker.HostConfig{
				Binds:      []string{dir + ":" + config.Workspace},
			},
		})

		if err != nil {
			log.Fatal(err)
		}

		if err := dockerClient.client.StartContainer(container.ID, nil); err != nil {
			log.Fatal(err)
		}

		opts := docker.AttachToContainerOptions{
			Container:    container.ID,
			Logs:         true,
			Stdout:       true,
			Stderr:       true,
			Stdin:        true,
			RawTerminal:  true,
			Stream:       true,
			ErrorStream:  os.Stderr,
			InputStream:  os.Stdin,
			OutputStream: os.Stdout,
		}

		if err := dockerClient.client.AttachToContainer(opts); err != nil {
			log.Fatal(err)
		}

		if err := dockerClient.client.RemoveContainer(docker.RemoveContainerOptions{
			ID:            container.ID,
			RemoveVolumes: true,
		}); err != nil {
			log.Fatal(err)
		}
	}

	app.Run(os.Args)
}

func configFromFile() Config {
	dir, _ := os.Getwd()
	indieFile, _ := ioutil.ReadFile(filepath.Join(dir, "indie.yml"))

	var config Config

	yaml.Unmarshal(indieFile, &config)

	if config.Workspace == "" {
		config.Workspace = "/workspace"
	}

	return config
}

func configureDockerClient(endpoint string, username string, password string, registry string) *DockerClient {
	var client *docker.Client
	var err error

	if len(os.Getenv("DOCKER_CERT_PATH")) != 0 {
		client, err = docker.NewTLSClient(endpoint,
			os.Getenv("DOCKER_CERT_PATH")+"/cert.pem",
			os.Getenv("DOCKER_CERT_PATH")+"/key.pem",
			os.Getenv("DOCKER_CERT_PATH")+"/ca.pem")

		if err != nil {
			log.Fatal(err)
		}

	} else {
		client, err = docker.NewClient(endpoint)

		if err != nil {
			log.Fatalf("Failed to connect to docker endpoint %q: %v", endpoint, err)
		}
	}

	var authConfig docker.AuthConfiguration

	if len(os.Getenv("DOCKER_USERNAME")) != 0 &&
		len(os.Getenv("DOCKER_PASSWORD")) != 0 {
		authConfig = docker.AuthConfiguration{
			Username:      username,
			Password:      password,
			ServerAddress: registry,
		}
	} else {
		authConfig = docker.AuthConfiguration{}
	}

	return &DockerClient{
		client:     client,
		authConfig: authConfig,
	}
}
