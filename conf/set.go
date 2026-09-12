package conf

import (
	"bufio"
	"fmt"
	"os"
)

/*
	saveConfigLocked writes the current cachedConfig map back to the file.
	Note: This overwrites the file and loses comments/formatting.
*/
func saveConfigLocked() error {
    file, err := os.Create(configPath)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := bufio.NewWriter(file)
    for k, v := range cachedConfig {
        line := fmt.Sprintf("%s = %s\n", k, v)
        if _, err := writer.WriteString(line); err != nil {
            return err
        }
    }
    
    if err := writer.Flush(); err != nil {
        return err
    }

    // Update the hash so the next Get doesn't trigger a reload
    newHash, err := calculateHash(configPath)
    if err == nil {
        configHash = newHash
    }

    return nil
}

/*
	Loads within an existing lock
*/
func loadConfigLocked() (map[string]string, error) {
    newHash, err := calculateHash(configPath)
    if err != nil {
        return nil, err
    }

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

/*
	Set updates an existing configuration key with a new value.
	It returns an error if the key does not already exist.
*/
func Set(key, value string) error {
    mu.Lock()
    defer mu.Unlock()

    // Ensure config is loaded before checking existence
    if cachedConfig == nil {
        if _, err := loadConfigLocked(); err != nil {
            return fmt.Errorf("failed to load config before setting: %w", err)
        }
    }

    // Check if the setting exists
    if _, exists := cachedConfig[key]; !exists {
        return fmt.Errorf("cannot modify setting: key '%s' does not exist", key)
    }

    // Update the memory cache
    cachedConfig[key] = value

    // Persist to disk
    if err := saveConfigLocked(); err != nil {
        return fmt.Errorf("failed to save config: %w", err)
    }

    return nil
}