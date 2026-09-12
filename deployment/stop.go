package deployment

import (
	"fmt"
	"os"
	"os/exec"

	"42tui/conf"
)

func StopByImage() error {

	imageName, err := conf.GetString("podman_image_name")
	if err != nil {
		return fmt.Errorf("error getting podman_image_name: %w", err)
	}

	cmd := exec.Command(
		"sh",
		"-c",
		fmt.Sprintf("podman rm -f $(podman ps -aq --filter ancestor=%s)", imageName),
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error removing containers: %w - %s", err, string(out))
	}

	return nil
}

func Stop() error {

	// Get env var name that stores the container ID
	contEnvId, err := conf.GetString("container_id_env_var_name")
	if err != nil {
		return fmt.Errorf("error getting container env var name: %w", err)
	}

	container := os.Getenv(contEnvId)
	if container == "" {
		return nil
	}

	// Try to kill container (best effort)
	killCmd := exec.Command("podman", "kill", container)
	if err := killCmd.Run(); err != nil {
		fmt.Printf("Warning: could not kill container (maybe already stopped): %v\n", err)
	}

	// Force remove container
	rmCmd := exec.Command("podman", "rm", "-f", container)
	output, err := rmCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error removing container: %w - %s", err, string(output))
	}

	if err := os.Unsetenv(contEnvId); err != nil {
		return fmt.Errorf("error unsetting env var: %w", err)
	}

	return nil
}
