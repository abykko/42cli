package deployment

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	
	"42tui/conf"
)

func Run() (string, error) {

	/*
		Obtenemos todos los datos necesarios para correr
		el contenedor del servidor y lo ejecutamos.
	*/

	log.Println("[Run] Preparing settings to run")

	addr, err := conf.GetString("server_addr")
	if err != nil {
		return "", fmt.Errorf("[Run] getting server_addr: %w", err)
	}

	port, err := conf.GetInt("server_port")
	if err != nil {
		return "", fmt.Errorf("[Run] getting server_port: %w", err)
	}

	imageName, err := conf.GetString("podman_image_name")
	if err != nil {
		return "", fmt.Errorf("[Run] getting podman_image_name: %w", err)
	}

	containerIDEnv, err := conf.GetString("container_id_env_var_name")
	if err != nil {
		return "", fmt.Errorf("[Run] getting container_id_env_var_name: %w", err)
	}

	secretEnv, err := conf.GetString("secret_env_var_name")
	if err != nil {
		return "", fmt.Errorf("[Run] getting secret_env_var_name: %w", err)
	}

	secret := os.Getenv(secretEnv)
	if secret == "" {
		return "", fmt.Errorf("[Run] environment variable %q is empty", secretEnv)
	}

	// Build command
	runCmd := exec.Command(
		"podman", "run",
		"-d",
		"-p", fmt.Sprintf("%s:%d:%d", addr, port, port),
		"-e", fmt.Sprintf("SECRET=%s", secret),
		fmt.Sprintf("%s:latest", imageName),
	)

	// Execute command
	output, err := runCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("[Run] running podman container: %w - %s", err, strings.TrimSpace(string(output)))
	}

	containerID := strings.TrimSpace(string(output))

	// Save container ID in env var
	if err := os.Setenv(containerIDEnv, containerID); err != nil {
		return "", fmt.Errorf("[Run] setting env var %q: %w", containerIDEnv, err)
	}

	return containerID, nil
}