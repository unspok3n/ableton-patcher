package ableton

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charlievieth/fastwalk"
)

const (
	defaultInstallDirWindows = "C:\\ProgramData\\Ableton"
	defaultInstallDirDarwin  = "/Applications"
)

type Installation struct {
	Path string
	Name string
}

type InstallationData struct {
	Path string
	Name string
}

var fwConfig = &fastwalk.Config{}

// promptForInstallDir prompts the user for the installation directory of Ableton Live.
func promptForInstallDir(defaultDir string) (string, error) {
	if _, statErr := os.Stat(defaultDir); !os.IsNotExist(statErr) {
		return defaultDir, nil
	}

	// defaultDir not found, prompt user for input
	println("Default path \""+defaultDir+"\" not found, please enter the path to your Ableton Live installation directory:")
	reader := bufio.NewReader(os.Stdin)
	line, readErr := reader.ReadString('\n')
	if readErr != nil {
		return "", readErr
	}

	inputDir := strings.TrimSpace(line)
	if _, statErr := os.Stat(inputDir); os.IsNotExist(statErr) {
		return "", os.ErrNotExist
	}

	return inputDir, nil
}

func FindInstallations() ([]Installation, error) {
	var installations []Installation

	var err error
	switch runtime.GOOS {
	case "windows":
		installDir, err := promptForInstallDir(defaultInstallDirWindows)
		if err != nil {
			return installations, err
		}

		err = fastwalk.Walk(fwConfig,
			installDir, func(path string, info fs.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() && strings.Contains(info.Name(), "Live") {
					binDir := filepath.Join(path, "/Program")
					instName := info.Name()

					err := filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return nil
						}

						if !info.IsDir() && strings.Contains(info.Name(), "Live") && strings.Contains(info.Name(), ".exe") {
							installations = append(installations, Installation{path, instName})
						}
						return nil
					})
					if err != nil {
						return err
					}

					return fastwalk.SkipDir
				}
				return nil
			})
	case "darwin":
		installDir, err := promptForInstallDir(defaultInstallDirDarwin)
		if err != nil {
			return installations, err
		}

		err = fastwalk.Walk(fwConfig, installDir, func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			entryName := info.Name()
			if info.IsDir() && strings.Contains(entryName, "Ableton Live") {
				binPath := filepath.Join(path, "/Contents/MacOS/Live")
				instName := strings.TrimPrefix(entryName, "Ableton ")
				instName = strings.TrimSuffix(instName, ".app")

				if _, err := os.Stat(binPath); err == nil {
					installations = append(installations, Installation{binPath, instName})
				}

				return fastwalk.SkipDir
			}
			return nil
		})
	}

	if err != nil {
		return installations, err
	}

	return installations, err
}

func FindInstallationData() ([]InstallationData, error) {
	var data []InstallationData

	appData, err := os.UserConfigDir()
	if err != nil {
		return data, err
	}
	defaultDataLocation := filepath.Join(appData, "/Ableton")
	err = filepath.Walk(defaultDataLocation, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && strings.Contains(info.Name(), "Live") {
			dataPath := path
			dataName := info.Name()
			err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}

				if info.IsDir() && strings.Contains(info.Name(), "Unlock") {
					data = append(data, InstallationData{dataPath, dataName})
				}
				return nil
			})
			if err != nil {
				return err
			}

			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return data, err
	}

	return data, err
}
