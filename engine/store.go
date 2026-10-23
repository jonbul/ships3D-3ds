//go:build !nintendo3ds

package main

import (
	"os"
	"path/filepath"
)

// profileFile is where the desktop keeps the profile: the user's configuration folder.
func profileFile() (string, error) {
	folder, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(folder, "ships3d", "profile.json"), nil
}

func readProfile() ([]byte, error) {
	file, err := profileFile()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(file)
}

func saveProfile(bytes []byte) error {
	file, err := profileFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
		return err
	}
	return os.WriteFile(file, bytes, 0o600)
}
