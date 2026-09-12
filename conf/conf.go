package conf

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	/*
		Configuration file path and
		variable to store settings
	*/
	cachedConfig map[string]string
	configHash   [32]byte
	configPath   = "conf/.conf"
	mu sync.RWMutex
)

/*
	Calculates sha256 for a given string, we use it to store
	the configuration file hash in order to detect changes
*/
func calculateHash(path string) ([32]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return [32]byte{}, fmt.Errorf("read config file: %w", err)
	}
	return sha256.Sum256(data), nil
}

func readConfig(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config file %s: %w", path, err)
	}
	defer file.Close()

	config := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid config line: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("empty config key in line: %s", line)
		}

		config[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan config: %w", err)
	}

	return config, nil
}

func loadConfig() (map[string]string, error) {
	mu.Lock()
	defer mu.Unlock()

	newHash, err := calculateHash(configPath)
	if err != nil {
		return nil, err
	}

	/*
		We compare the cached settings hash with the new hash,
		if they're different we read and store the settings again.
		This reduces reading operations.
	*/
	if cachedConfig == nil || newHash != configHash {
		cfg, err := readConfig(configPath)
		if err != nil {
			return nil, err
		}

		cachedConfig = cfg
		configHash = newHash
	}

	return cachedConfig, nil
}
