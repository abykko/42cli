package conf

import (
	"fmt"
	"strconv"
	"strings"
)

/*
	Get setting as a string.
*/
func GetString(key string) (string, error) {
	cfg, err := loadConfig()
	if err != nil {
		return "", fmt.Errorf("config error: %w", err)
	}

	value, ok := cfg[key]
	if !ok {
		return "", fmt.Errorf("config key not found: %s", key)
	}

	return value, nil
}

/*
	Get() is an alias method refering to GetString()
*/
func Get(key string) (string, error) {
	return GetString(key)
}

/*
	GetString plus Atoi
*/
func GetInt(key string) (int, error) {
	value, err := GetString(key)
	if err != nil {
		return 0, err
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid int for %s: %w", key, err)
	}

	return i, nil
}

func GetBool(key string) (bool, error) {
	value, err := GetString(key)
	if err != nil {
		return false, err
	}

	/*
		Support for different bool indications
	*/
	switch strings.ToLower(value) {
	case "true", "1", "yes", "y", "on":
		return true, nil
	case "false", "0", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool for %s: %s", key, value)
	}
}