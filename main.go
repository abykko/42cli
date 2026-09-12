package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"42tui/conf"
	"42tui/languages"
	"42tui/tui"
)

/*
	Verifica la existencia de una utilidad dada
	por su nombre.
*/
func checkCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

/*
	Verifica la existencia de un archivo dado
	por su nombre.
*/
func checkFile(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

/*
	Verifica si una opción de configuración tiene
	el valor esperado.
*/
func checkSetting(key string, value string) bool {
	result, err := conf.GetString(key)
	if err != nil {
		return false
	}

	cleanResult := strings.TrimSpace(result)

	if cleanResult == "" {
		return false
	}

	if cleanResult != strings.TrimSpace(value) {
		return false
	}

	return true
}

/*
	Verifica si una opción de configuración está vacía.
*/
func checkIsEmpty(key string) bool {
	result, err := conf.GetString(key)
	if err != nil {
		return false
	}

	cleanResult := strings.TrimSpace(result)

	return cleanResult == ""
}

/*
	Realiza las comprobaciones necesarias antes de
	iniciar la aplicación.

	Los mensajes de error se obtienen mediante lan para
	que puedan mostrarse en el idioma seleccionado.
*/
func troubleshootingChecker(lan *languages.Languages) error {
	// Verificar si Podman está instalado.
	if !checkCommand("podman") {
		return fmt.Errorf("%s", lan.Get("error.podman_not_installed"))
	}

	// Verificar si Go está instalado.
	if !checkCommand("go") {
		return fmt.Errorf("%s", lan.Get("error.go_not_installed"))
	}

	// Verificar si existe el archivo go.mod.
	if !checkFile("go.mod") {
		return fmt.Errorf("%s", lan.Get("error.go_mod_missing"))
	}

	// Verificar autologin y sus credenciales.
	if checkSetting("autologin", "yes") {
		if checkIsEmpty("user_login") || checkIsEmpty("password_login") {
			return fmt.Errorf("%s", lan.Get("error.autologin_credentials_empty"))
		}
	}

	return nil
}

func main() {
	/*
		Preparamos el archivo de logs.
	*/
	logFile, err := os.OpenFile(
		"debug.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)

	if err == nil {
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	/*
		Cargamos el sistema de idiomas.

		El idioma seleccionado será utilizado por las
		comprobaciones y posteriormente por la interfaz.
	*/
	lan, err := languages.New("es")
	if err != nil {
		log.Printf("Error loading language: %v", err)
		fmt.Printf("Error loading language: %v\n", err)
		os.Exit(1)
	}

	/*
		Realizamos las comprobaciones previas al arranque.
	*/
	if err := troubleshootingChecker(lan); err != nil {
		log.Printf("Troubleshooting error: %v", err)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	/*
		Arranque de la interfaz.
	*/
	tui.Tui()
}