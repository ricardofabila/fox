package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

func FileExistsInHome(filePath string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		color.Red("Error loading home directory: %v", err)
		os.Exit(1)
	}

	return FileExists(filepath.Join(home, filePath))
}

func FileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	}

	return false
}

func CreateDirectoryIfNotExistsInHome(directoryPath string) {
	home, err := os.UserHomeDir()
	if err != nil {
		color.Red("Error loading home directory: %v", err)
		os.Exit(1)
	}

	CreateDirectoryIfNotExists(filepath.Join(home, directoryPath))
}

func CreateDirectoryIfNotExists(directoryPath string) {
	if _, err := os.Stat(directoryPath); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(directoryPath, os.ModePerm)
		if err != nil {
			color.Red("Error creating file: %v", err)
			os.Exit(1)
		}
	}
}

func CreateFileIfNotExistsInHome(filePath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		color.Red("Error loading home directory: %v", err)
		os.Exit(1)
	}

	return CreateFileIfNotExists(filepath.Join(home, filePath))
}

func CreateFileIfNotExists(filePath string) error {
	_, err := os.Stat(filePath)
	if err == nil {
		return err
	}

	if errors.Is(err, os.ErrNotExist) {
		file, e := os.Create(filePath)
		if e != nil {
			return fmt.Errorf("error creating blank %s file: %v", filePath, e)
		}
		e = file.Chmod(0777)
		if e != nil {
			return fmt.Errorf("error creating blank %s file: %v", filePath, e)
		}
		e = file.Close()
		if e != nil {
			return fmt.Errorf("error closing %s file: %v", filePath, e)
		}

		return nil
	}

	// Schrödinger's file may or may not exist. See err for details.
	// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
	return fmt.Errorf("error loading %s file: %v", filePath, err)
}

func RemoveFile(path string) error {
	data, err := ExecuteCommandAndGetOutput("rm", []string{"-f", path}...)
	if err != nil {
		color.Red(data)
		return PrintAndReturnError(err.Error())
	}

	return nil
}

func RemoveDirectory(path string) error {
	data, err := ExecuteCommandAndGetOutput("rm", []string{"-rf", path}...)
	if err != nil {
		color.Red(data)
		return PrintAndReturnError(err.Error())
	}

	return nil
}

func MoveFile(file, target string) error {
	data, err := ExecuteCommandAndGetOutput("mv", []string{"-f", file, target}...)
	if err != nil {
		color.Red(data)
		return PrintAndReturnError(err.Error())
	}

	return nil
}

func MakeFileExecutable(path string) error {
	data, err := ExecuteCommandAndGetOutput("chmod", []string{"+x", path}...)
	if err != nil {
		color.Red(data)
		return PrintAndReturnError(err.Error())
	}

	return nil
}

func GetPermissions(filename string) (string, error) {
	permissions := ""
	info, err := os.Stat(filename)
	if err != nil {
		return "", err
	}

	mode := info.Mode()

	permissions += "Owner: "
	for i := 1; i < 4; i++ {
		permissions += string(mode.String()[i])
	}

	permissions += " Group: "
	for i := 4; i < 7; i++ {
		permissions += string(mode.String()[i])
	}

	permissions += " Other: "
	for i := 7; i < 10; i++ {
		permissions += string(mode.String()[i])
	}

	return permissions, nil
}
