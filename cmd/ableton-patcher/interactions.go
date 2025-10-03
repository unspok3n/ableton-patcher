package main

import (
	"crypto/dsa"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unspok3n/ableton-patcher/internal/ableton"
)

func (app *application) mainMenu() {
	for {
		fmt.Print("1. Patch\n2. Unpatch\n3. Deauthorize\n4. Generate license\n5. Generate DSA key pair")
		fmt.Print("\nSelect option: ")
		option := GetLine()
		ClearScreen()
		switch option {
		case "1":
			app.patcher(false)
		case "2":
			app.patcher(true)
		case "3":
			app.deauthorize()
		case "4":
			app.licenseGenerator()
		case "5":
			app.keyGenerator()
		default:
			fmt.Print("invalid option")
		}
		Pause()
	}
}

func (app *application) deauthorize() {
	selectedInstsData := installationDataSelector()

	for i, data := range selectedInstsData {
		unlockFile := filepath.Join(data.Path, "/Unlock/Unlock.json")
		os.Remove(unlockFile)
		fmt.Printf("Deauthorized (%s)", data.Name)
		if i != len(selectedInstsData)-1 {
			fmt.Print("\n")
		}
	}
}

func (app *application) licenseGenerator() {
	fmt.Print("Enter HWID: ")
	hwid := GetLine()

	fmt.Print("Enter major version: ")
	version := GetLine()
	versionInt, err := strconv.Atoi(version)
	if err != nil {
		LogFatalError("parse version", err)
	}

	var edition int
	var editionName string
	fmt.Print("1. Suite\n2. Standard\n3. Intro\n4. Lite")
	for {
		fmt.Print("\nSelect edition: ")
		editionNumber := GetLine()
		switch editionNumber {
		case "1":
			edition = 2
			editionName = "Suite"
		case "2":
			edition = 0
			editionName = "Standard"
		case "3":
			edition = 3
			editionName = "Intro"
		case "4":
			edition = 4
			editionName = "Lite"
		default:
			fmt.Println("Invalid edition")
			continue
		}
		break
	}

	license, err := ableton.GenerateLicense(*app.key, hwid, edition, versionInt)
	if err != nil {
		LogFatalError("generate license", err)
	}

	filename := fmt.Sprintf("Authorize_%s_%s.auz", editionName, hwid)
	path, err := ExecutableDirFilePath(filename)
	if err != nil {
		LogFatalError("get executable path", err)
	}

	if err := ableton.WriteAuthorizationFile(license, path); err != nil {
		LogFatalError("write authorization file", err)
	}

	fmt.Print("License file generated")
}

func (app *application) keyGenerator() {
	// There's gotta be a better way of doing this
	var priv dsa.PrivateKey
	for {
		var params dsa.Parameters
		if err := dsa.GenerateParameters(&params, rand.Reader, dsa.L1024N160); err != nil {
			LogFatalError("generate parameters", err)
		}

		priv.Parameters = params
		if err := dsa.GenerateKey(&priv, rand.Reader); err != nil {
			LogFatalError("generate key", err)
		}

		if priv.Y.BitLen() == 1024 && priv.G.BitLen() < 1024 {
			break
		}
	}

	keyString, err := ableton.PrivateDSAToHex(&priv)
	if err != nil {
		LogFatalError("key to hex", err)
	}

	fmt.Printf("\n%s\n\n", keyString)

	fmt.Print("Update config? (y/n): ")
	answer := GetLine()
	if answer == "y" {
		app.config.PrivateKey = keyString
		app.key = &priv
		err := app.config.Save(app.configPath)
		if err != nil {
			LogFatalError("save config", err)
		}
		fmt.Print("Config saved")
	}
}

func (app *application) patcher(reverse bool) {
	public, err := ableton.PublicDSAToHex(&app.key.PublicKey)
	if err != nil {
		LogFatalError("key to hex", err)
	}

	if len(public) != len(app.config.OriginalPublicKey) {
		LogFatalError("compare keys", fmt.Errorf("key length mismatch: %d vs %d", len(public), len(app.config.OriginalPublicKey)))
	}

	selectedInsts := installationSelector()

	for i, installation := range selectedInsts {
		fileBytes, err := os.ReadFile(installation.Path)
		if err != nil {
			LogFatalError("open file", err)
		}

		var newContents string
		var replacementsCounter int
		if reverse {
			newContents, replacementsCounter = ReplaceAndCount(string(fileBytes), public, app.config.OriginalPublicKey)
		} else {
			newContents, replacementsCounter = ReplaceAndCount(string(fileBytes), app.config.OriginalPublicKey, public)
		}
		if replacementsCounter == 0 {
			fmt.Printf("Key not found (%s)", installation.Path)
		} else {
			err = os.WriteFile(installation.Path, []byte(newContents), 0)
			if err != nil {
				LogFatalError("write file", err)
			}

			fmt.Printf("Patch/Unpatch applied succesfully (%s)", installation.Path)
		}

		if i != len(selectedInsts)-1 {
			fmt.Print("\n")
		}
	}
}

func installationSelector() []ableton.Installation {
	installations, err := ableton.FindInstallations()
	if err != nil {
		LogFatalError("find installations", err)
	}

	installationsLen := len(installations)

	for i, installation := range installations {
		if installation.Path != "" {
			fmt.Printf("%d. %s (%s)\n", i+1, installation.Name, installation.Path)
		} else {
			fmt.Printf("%d. %s\n", i+1, installation.Name)
		}
	}

	var selectedInsts []ableton.Installation
	for {
		fmt.Print("Choose the number of installation(if multipule, separate by space): ")
		input := GetLine()
		requestedInsts := strings.Split(input, " ")
		for _, result := range requestedInsts {
			resultInt, err := strconv.Atoi(result)
			if err != nil {
				fmt.Printf("invalid installation number: %s", result)
				continue
			}

			if resultInt > installationsLen || resultInt == 0 {
				fmt.Printf("invalid installation number: %d", resultInt)
				continue
			}

			if installations[resultInt-1].Path == "" {
				fmt.Print("Enter path to the Live executable: ")
				installations[resultInt-1].Path = GetLine()
				if _, err := os.Stat(installations[resultInt-1].Path); err != nil {
					fmt.Printf("invalid executable path")
					continue
				}
			}

			selectedInsts = append(selectedInsts, installations[resultInt-1])
		}
		break
	}

	return selectedInsts
}

func installationDataSelector() []ableton.InstallationData {
	data, err := ableton.FindInstallationData()
	if err != nil {
		LogFatalError("find installation data", err)
	}

	dataLen := len(data)

	for i, installationData := range data {
		if installationData.Path != "" {
			fmt.Printf("%d. %s (%s)\n", i+1, installationData.Name, installationData.Path)
		} else {
			fmt.Printf("%d. %s\n", i+1, installationData.Name)
		}
	}

	var selectedInstsData []ableton.InstallationData
	for {
		fmt.Print("Enter the installation data number(s): ")
		input := GetLine()
		requestedInstsData := strings.Split(input, " ")
		for _, result := range requestedInstsData {
			resultInt, err := strconv.Atoi(result)
			if err != nil {
				fmt.Printf("invalid installation data number: %s", result)
				continue
			}

			if resultInt > dataLen || resultInt == 0 {
				fmt.Printf("invalid installation data number: %d", resultInt)
				continue
			}

			if data[resultInt-1].Path == "" {
				fmt.Print("Enter path to the Live data folder: ")
				data[resultInt-1].Path = GetLine()
				if _, err := os.Stat(data[resultInt-1].Path); err != nil {
					fmt.Printf("invalid data folder path")
					continue
				}
			}

			selectedInstsData = append(selectedInstsData, data[resultInt-1])
		}
		break
	}

	return selectedInstsData
}
