//go:build nintendo3ds

package main

import "github.com/akrck02/go-3ds/storage"

// The 3DS keeps the profile on the SD card, where the 3DS program keeps it: one profile for both.
const (
	profileFolder = "ships3d"
	profileName   = "profile.json"
)

func readProfile() ([]byte, error) {
	folder, err := storage.Open(profileFolder)
	if err != nil {
		return nil, err
	}
	return folder.Load(profileName)
}

func saveProfile(bytes []byte) error {
	folder, err := storage.Open(profileFolder)
	if err != nil {
		return err
	}
	return folder.Save(profileName, bytes)
}
