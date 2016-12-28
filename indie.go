package main

import (
	"os"
	"github.com/jawher/mow.cli"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"io/ioutil"
	"github.com/fsouza/go-dockerclient"
	"log"
)

type Config struct {
	Image       string
	Workspace   string
	Environment [] string
}

func main() {
	app := cli.App("indie", "Docker command execution")

	app.Spec = "[OPTIONS] CMD ARG..."
	app.Version("v version", "indie 0.1.0")

	var (
		endpoint = app.String(cli.StringOpt{
			Name:   "docker-endpoint",
			Value:  "unix:///var/run/docker.sock",
			Desc:   "The Docker Host or socket",
			EnvVar: "DOCKER_HOST",
		})

		cmd  = app.StringsArg("CMD", nil, "Command")
		args = app.StringArg("ARG", "", "Arguments")
	)

	app.Action = func() {
		config := configFromFile()
		client := configureDockerClient(*endpoint)

		err := client.PullImage(docker.PullImageOptions{
			Repository:   config.Image,
		}, docker.AuthConfiguration{})

		if err != nil {
			log.Fatal(err)
		}

		dir, _ := os.Getwd()

		container, err := client.CreateContainer(docker.CreateContainerOptions{
			Name: "indie",
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
				AutoRemove: true,
			},
		})

		if err != nil {
			log.Fatal(err)
		}

		err = client.StartContainer(container.ID, nil)

		if err != nil {
			log.Fatal(err)
		}

		err = client.AttachToContainer(docker.AttachToContainerOptions{
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
		})

		if err != nil {
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

	return config
}

func configureDockerClient(endpoint string) *docker.Client {
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

	return client
}
