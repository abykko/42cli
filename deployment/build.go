package deployment

import (
    "crypto/rand"
    "encoding/hex"
    "log"
    "fmt"
    "os"
    "os/exec"

    "42tui/conf"
)

func generateSecret() string {
    b := make([]byte, 32) // 256 bits

    _, err := rand.Read(b)
    if err != nil {
        panic(err)
    }

    return hex.EncodeToString(b)
}

func Build() error {
    
    /*
        En caso de haber una instancia previa (no debería ocurrir)
    */
    StopByImage()

    /*
        Genera una clave que se usa para validar las
        peticiones al contenedor del servidor.

        La clave se almacena en una variable de entorno
        correspondiente a la ejecución del programa.
    */
    secret := generateSecret()

    secretEnv, err := conf.GetString("secret_env_var_name")
    if err != nil {
        log.Println("[Build] error getting secret env name:", err)
        return err
    }

    if err := os.Setenv(secretEnv, secret); err != nil {
        log.Println("[Build] error setting secret env var:", err)
        return err
    }

    log.Println("[Build] secret generated")

    /*
        Montamos el contenedor del servidor
    */
    servDir, err := conf.GetString("server_dir")
    if err != nil {
        log.Println("[Build] error getting server_dir:", err)
        return err
    }

    port, err := conf.GetInt("server_port")
	if err != nil {
		return fmt.Errorf("getting server_port: %w", err)
	}

    imageName, err := conf.GetString("podman_image_name")
    if err != nil {
        log.Println("[Build] error getting podman_image_name:", err)
        return err
    }

    log.Printf("[Build] building podman image %s from %s\n", imageName, servDir)

    buildCmd := exec.Command(
        "podman", "build",
        "--rm",
        "--build-arg", fmt.Sprintf("PORT=%d", port),
        "-t", imageName, ".",
    )
    buildCmd.Dir = servDir
    buildCmd.Stdout = log.Writer()
	buildCmd.Stderr = log.Writer()

    log.Println("=== Building process ===")

    if err := buildCmd.Run(); err != nil {
        log.Println("[Build] error building podman image:", err)
        return err
    }

    log.Println("[Build] podman build completed successfully")

    return nil
}