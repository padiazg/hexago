package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateDir creates a directory with all parent directories if they don't exist
func CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// WriteFile writes content to a file, creating parent directories if needed
func WriteFile(path string, content []byte) error {
	// Create parent directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := CreateDir(dir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// FileExists checks if a file or directory exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory checks if the path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// CreateDirs creates multiple directories
func CreateDirs(basePath string, dirs []string) error {
	for _, dir := range dirs {
		path := filepath.Join(basePath, dir)
		if err := CreateDir(path); err != nil {
			return err
		}
	}
	return nil
}

// TouchFile creates an empty file (like Unix touch command)
func TouchFile(path string) error {
	return WriteFile(path, []byte{})
}

// ReadDir reads directory entries
func ReadDir(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names, nil
}
