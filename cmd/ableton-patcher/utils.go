package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

func LogFatalError(caller string, err error) {
	message := fmt.Sprintf("%s: %s", caller, err.Error())
	fmt.Println(message)
	debug.PrintStack()
	Shutdown()
}

func ExecutableDirFilePath(fileName string) (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %v", err)
	}
	execDir := filepath.Dir(execPath)
	filePathExec := filepath.Join(execDir, fileName)
	return filePathExec, nil
}

func WorkingDirFilePath(fileName string) (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %v", err)
	}

	filePathCurrent := filepath.Join(workingDir, fileName)
	return filePathCurrent, nil
}

func FindFile(fileName string) (string, error) {
	filePathExec, err := ExecutableDirFilePath(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %v", err)
	}

	_, err = os.Stat(filePathExec)
	if err == nil {
		return filePathExec, nil
	}

	filePathCurrent, err := WorkingDirFilePath(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %v", err)
	}

	_, err = os.Stat(filePathCurrent)
	if err == nil {
		return filePathCurrent, nil
	}

	return "", fmt.Errorf("%s not found", fileName)
}

func GetLine() string {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input string: %v\n", err)
		os.Exit(1)
	}
	input = strings.TrimSuffix(input, "\n")
	input = strings.TrimSuffix(input, "\r")
	return input
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

func Pause() {
	fmt.Println("\n\nPress enter to continue (or Ctrl+C to exit)")
	fmt.Scanln()
	ClearScreen()
}

func Shutdown() {
	fmt.Println("\nPress enter to exit")
	fmt.Scanln()
	os.Exit(0)
}

func ReplaceAndCount(s, old, new string) (string, int) {
	count := strings.Count(s, old)
	replaced := strings.ReplaceAll(s, old, new)
	return replaced, count
}
