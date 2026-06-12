// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Qirashi
// Project: installer

package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eiannone/keyboard"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// EMBED FILES ================================================================

//go:embed assets/dvpl_go.ico
var iconData []byte

//go:embed assets/dvpl.exe
var exeData []byte

// CONFIG =====================================================================

const (
	AppName           = "DvplGo 2.1.5 x64 | Copyright (c) 2026 Qirashi"
	DefaultInstallDir = `C:\Tools\DvplGO`
	RegRoot           = `Software\XInstaller\DVPLGO`
)

var InstallDir = DefaultInstallDir

// MAIN =======================================================================

func main() {
	enableVirtualTerminal()
	err := runUI()

	if err != nil {
		panic(err)
	}
}

func enableVirtualTerminal() {
	var mode uint32
	handle := windows.Stdout
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return
	}
	mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	windows.SetConsoleMode(handle, mode)
}

// UI ========================================================================

func runUI() error {
	if err := keyboard.Open(); err != nil {
		return err
	}
	defer keyboard.Close()

	reader := bufio.NewReader(os.Stdin)
	selected := 0

	for {
		clearScreen()
		fmt.Printf("\n%s\n", AppName)
		fmt.Println("Use the arrow keys to navigate: ↓ ↑")
		fmt.Println()

		installed := isInstalled()
		var options []string
		var actions []func()

		if installed {
			options = append(options, "Reinstall DvplGo")
			actions = append(actions, func() {
				if err := reinstall(); err != nil {
					printError(err)
				} else {
					printSuccess("DvplGo successfully reinstalled!")
				}
				pause(reader)
			})

			options = append(options, "Uninstall DvplGo")
			actions = append(actions, func() {
				if err := uninstall(); err != nil {
					printError(err)
				} else {
					printSuccess("DvplGo successfully uninstalled!")
				}
				pause(reader)
			})
		} else {
			options = append(options, "Install DvplGo")
			actions = append(actions, func() {
				if err := selectAndInstall(reader); err != nil {
					printError(err)
				} else {
					printSuccess("DvplGo successfully installed!")
				}
				pause(reader)
			})
		}

		options = append(options, "")
		actions = append(actions, nil)

		// Configure MAX_WORKERS with status
		maxWorkersStatus := "Configure MAX_WORKERS"
		if mw, _ := getEnvVar("DVPL_MAX_WORKERS"); mw != "" {
			maxWorkersStatus = fmt.Sprintf("Configure MAX_WORKERS (%s)", mw)
		}
		options = append(options, maxWorkersStatus)
		actions = append(actions, func() {
			keyboard.Close()
			if err := configureMaxWorkers(reader); err != nil {
				printError(err)
			}
			pause(reader)
			keyboard.Open()
		})

		// Configure COMPRESS_TYPE with status
		compressTypeStatus := "Configure COMPRESS_TYPE"
		if ct, _ := getEnvVar("DVPL_COMPRESS_TYPE"); ct != "" {
			compressTypeStatus = fmt.Sprintf("Configure COMPRESS_TYPE (%s)", ct)
		}
		options = append(options, compressTypeStatus)
		actions = append(actions, func() {
			keyboard.Close()
			if err := configureCompressType(reader); err != nil {
				printError(err)
			}
			pause(reader)
			keyboard.Open()
		})

		// PATH status
		pathStatus := "Add to PATH"
		if isInPath() {
			pathStatus = "Remove from PATH (Installed)"
		} else {
			pathStatus = "Add to PATH (Not installed)"
		}
		options = append(options, pathStatus)
		actions = append(actions, func() {
			if isInPath() {
				if err := removeFromPath(); err != nil {
					printError(err)
				} else {
					printSuccess("Removed from PATH")
				}
			} else {
				if err := addToPath(); err != nil {
					printError(err)
				} else {
					printSuccess("Added to PATH")
				}
			}
			pause(reader)
		})

		options = append(options, "")
		actions = append(actions, nil)

		options = append(options, "Exit")
		actions = append(actions, func() {})

		// Display menu
		for i, opt := range options {
			if opt == "" {
				fmt.Println()
				continue
			}
			marker := "  "
			if i == selected {
				marker = ">  "
			}
			fmt.Printf("%s%s\n", marker, opt)
		}

		// Read keyboard input
		_, key, err := keyboard.GetKey()
		if err != nil {
			return err
		}

		switch key {
		case keyboard.KeyArrowUp:
			selected--
			for selected >= 0 && options[selected] == "" {
				selected--
			}
			if selected < 0 {
				selected = len(options) - 1
				for selected >= 0 && options[selected] == "" {
					selected--
				}
			}
		case keyboard.KeyArrowDown:
			selected++
			for selected < len(options) && options[selected] == "" {
				selected++
			}
			if selected >= len(options) {
				selected = 0
				for selected < len(options) && options[selected] == "" {
					selected++
				}
			}
		case keyboard.KeyEnter:
			if selected == len(options)-1 {
				return nil
			}
			actions[selected]()
		case keyboard.KeyEsc:
			return nil
		}
	}
}

// CONSOLE HELPERS ============================================================

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func pause(reader *bufio.Reader) {
	fmt.Print("\nPress Enter to continue...")
	reader.ReadString('\n')
}

func printError(err error) {
	fmt.Printf("\n\033[31mERROR: %v\033[0m\n", err)
}

func printSuccess(msg string) {
	fmt.Printf("\n\033[32mSUCCESS: %s\033[0m\n", msg)
}

func readInput(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// USER ACTIONS ===============================================================

func selectAndInstall(reader *bufio.Reader) error {
	if err := selectInstallDir(reader); err != nil {
		return err
	}
	return install()
}

func selectInstallDir(reader *bufio.Reader) error {
	selected := 0
	options := []string{
		DefaultInstallDir,
		"Custom Path",
	}

	for {
		clearScreen()
		fmt.Println("\nSelect Installation Path")
		fmt.Println("Use arrow keys to navigate: ↓ ↑")
		fmt.Println()

		for i, opt := range options {
			marker := "  "
			if i == selected {
				marker = ">  "
			}
			fmt.Printf("%s%s\n", marker, opt)
		}

		_, key, err := keyboard.GetKey()
		if err != nil {
			return err
		}

		switch key {
		case keyboard.KeyArrowUp:
			selected--
			if selected < 0 {
				selected = len(options) - 1
			}
		case keyboard.KeyArrowDown:
			selected++
			if selected >= len(options) {
				selected = 0
			}
		case keyboard.KeyEnter:
			switch selected {
			case 0:
				InstallDir = DefaultInstallDir
				fmt.Printf("\nUsing: %s\n", InstallDir)
				return nil
			case 1:
				keyboard.Close()
				if err := inputCustomPath(reader); err != nil {
					keyboard.Open()
					return err
				}
				keyboard.Open()
				fmt.Printf("\nUsing: %s\n", InstallDir)
				return nil
			}
		case keyboard.KeyEsc:
			return fmt.Errorf("installation cancelled")
		}
	}
}

func inputCustomPath(reader *bufio.Reader) error {
	for {
		path := readInput(reader, "Enter custom installation path: ")

		if path == "" {
			printError(fmt.Errorf("path cannot be empty"))
			continue
		}

		// Check if path is absolute
		if !filepath.IsAbs(path) {
			printError(fmt.Errorf("path must be absolute (e.g., C:\\Tools\\DvplGO or C:\\Program Files\\...)"))
			continue
		}

		// Check for invalid characters
		if strings.ContainsAny(path, `<>"|?*`) {
			printError(fmt.Errorf("path contains invalid characters: < > \" | ? *"))
			continue
		}

		// Optional: check if path is on a valid drive
		if len(path) < 2 || path[1] != ':' {
			printError(fmt.Errorf("path must start with a drive letter (e.g., C:, D:)"))
			continue
		}

		InstallDir = path
		printSuccess(fmt.Sprintf("Path accepted: %s", InstallDir))
		return nil
	}
}

func configureMaxWorkers(reader *bufio.Reader) error {
	current, _ := getEnvVar("DVPL_MAX_WORKERS")

	if current != "" {
		fmt.Printf("Current DVPL_MAX_WORKERS: %s\n", current)
	}

	input := readInput(reader, "Enter DVPL_MAX_WORKERS (1-99) or press Enter to skip: ")

	if input == "" {
		deleteEnvVar("DVPL_MAX_WORKERS")
		return nil
	}

	val, err := strconv.Atoi(input)
	if err != nil || val < 1 || val > 99 {
		return fmt.Errorf("invalid value. Use 1-99")
	}

	return setEnvVar("DVPL_MAX_WORKERS", input)
}

func configureCompressType(reader *bufio.Reader) error {
	current, _ := getEnvVar("DVPL_COMPRESS_TYPE")

	if current != "" {
		fmt.Printf("Current DVPL_COMPRESS_TYPE: %s\n", current)
	}

	input := readInput(reader, "Enter DVPL_COMPRESS_TYPE (0,1,2) or press Enter to skip: ")

	if input == "" {
		deleteEnvVar("DVPL_COMPRESS_TYPE")
		return nil
	}

	val, err := strconv.Atoi(input)
	if err != nil || val < 0 || val > 2 {
		return fmt.Errorf("invalid value. Use 0,1,2")
	}

	return setEnvVar("DVPL_COMPRESS_TYPE", input)
}

// INSTALL ====================================================================

func install() error {
	err := os.MkdirAll(InstallDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create installation directory: %v", err)
	}

	exePath := filepath.Join(InstallDir, "dvpl.exe")
	iconPath := filepath.Join(InstallDir, "dvpl_go.ico")

	err = os.WriteFile(exePath, exeData, 0755)
	if err != nil {
		return fmt.Errorf("failed to extract dvpl.exe: %v", err)
	}

	err = os.WriteFile(iconPath, iconData, 0644)
	if err != nil {
		return fmt.Errorf("failed to extract dvpl_go.ico: %v", err)
	}

	err = writeInstallRegistry()
	if err != nil {
		return fmt.Errorf("failed to write installation registry: %v", err)
	}

	err = addToPath()
	if err != nil {
		return fmt.Errorf("failed to add to PATH: %v", err)
	}

	err = createContextMenus()
	if err != nil {
		return fmt.Errorf("failed to create context menus: %v", err)
	}
	return nil
}

func reinstall() error {
	removeContextMenus()

	err := removeFromPath()
	if err != nil {
		return fmt.Errorf("failed to remove from PATH: %v", err)
	}

	deleteInstallRegistry()

	err = os.RemoveAll(InstallDir)
	if err != nil {
		return fmt.Errorf("failed to remove old files: %v", err)
	}

	err = install()
	if err != nil {
		return err
	}

	return nil
}

func uninstall() error {
	removeContextMenus()

	err := removeFromPath()
	if err != nil {
		return fmt.Errorf("failed to remove from PATH: %v", err)
	}

	deleteInstallRegistry()

	err = os.RemoveAll(InstallDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove files: %v", err)
	}

	return nil
}

// INSTALL STATE =============================================================

func isInstalled() bool {

	_, err := registry.OpenKey(
		registry.CURRENT_USER,
		RegRoot,
		registry.READ,
	)

	return err == nil
}

// PATH ======================================================================

func addToPath() error {

	key, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`Environment`,
		registry.ALL_ACCESS,
	)

	if err != nil {
		return err
	}

	defer key.Close()

	path, _, _ := key.GetStringValue("Path")

	parts := splitPath(path)

	for _, p := range parts {

		if strings.EqualFold(p, InstallDir) {
			return nil
		}
	}

	parts = append(parts, InstallDir)

	newPath := strings.Join(parts, ";")

	err = key.SetExpandStringValue("Path", newPath)

	if err != nil {
		return err
	}

	return nil
}

func removeFromPath() error {

	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Environment`,
		registry.ALL_ACCESS,
	)

	if err != nil {
		return err
	}

	defer key.Close()

	path, _, _ := key.GetStringValue("Path")

	parts := splitPath(path)

	var filtered []string

	for _, p := range parts {

		if !strings.EqualFold(p, InstallDir) {
			filtered = append(filtered, p)
		}
	}

	newPath := strings.Join(filtered, ";")

	return key.SetExpandStringValue("Path", newPath)
}

func splitPath(path string) []string {

	raw := strings.Split(path, ";")

	var out []string

	for _, p := range raw {

		p = strings.TrimSpace(p)

		if p != "" {
			out = append(out, p)
		}
	}

	return out
}

func isInPath() bool {

	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Environment`,
		registry.READ,
	)

	if err != nil {
		return false
	}

	defer key.Close()

	path, _, _ := key.GetStringValue("Path")
	parts := splitPath(path)

	for _, p := range parts {
		if strings.EqualFold(p, InstallDir) {
			return true
		}
	}

	return false
}

// ENVIRONMENT ================================================================

func getEnvVar(name string) (string, error) {

	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Environment`,
		registry.READ,
	)

	if err != nil {
		return "", err
	}

	defer key.Close()

	value, _, err := key.GetStringValue(name)
	if err != nil {
		return "", nil // Return empty string if not found, not an error
	}

	return value, nil
}

func setEnvVar(name, value string) error {

	key, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`Environment`,
		registry.ALL_ACCESS,
	)

	if err != nil {
		return err
	}

	defer key.Close()

	err = key.SetStringValue(name, value)
	if err != nil {
		return err
	}

	return nil
}

func deleteEnvVar(name string) error {

	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Environment`,
		registry.ALL_ACCESS,
	)

	if err != nil {
		return err
	}

	defer key.Close()

	err = key.DeleteValue(name)
	if err != nil {
		return err
	}

	return nil
}

// INSTALL REGISTRY ===========================================================

func writeInstallRegistry() error {

	key, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		RegRoot,
		registry.ALL_ACCESS,
	)

	if err != nil {
		return err
	}

	defer key.Close()

	err = key.SetStringValue("InstallLocation", InstallDir)

	if err != nil {
		return err
	}

	return nil
}

func deleteInstallRegistry() {

	registry.DeleteKey(registry.CURRENT_USER, RegRoot)
}

// CONTEXT MENU ===============================================================

func createContextMenus() error {

	err := createFileContextMenu()

	if err != nil {
		return err
	}

	err = createDirectoryContextMenu()

	if err != nil {
		return err
	}

	err = createBackgroundContextMenu()

	if err != nil {
		return err
	}

	return nil
}

func removeContextMenus() {

	deleteRegistryTree(registry.CLASSES_ROOT, `*\shell\DvplTools`)
	deleteRegistryTree(registry.CLASSES_ROOT, `Directory\shell\DvplTools`)
	deleteRegistryTree(registry.CLASSES_ROOT, `Directory\Background\shell\DvplTools`)
}

func createFileContextMenu() error {

	base := `*\shell\DvplTools`

	err := createMenuRoot(base)

	if err != nil {
		return err
	}

	err = createCommand(
		base+`\shell\01Compress`,
		`dvpl -c "Compress"`,
		fmt.Sprintf(`"%s\dvpl.exe" -c -trust-data -i "%%1"`, InstallDir),
	)

	if err != nil {
		return err
	}

	err = createCommand(
		base+`\shell\02Decompress`,
		`dvpl -d "Decompress"`,
		fmt.Sprintf(`"%s\dvpl.exe" -d -trust-data -i "%%1"`, InstallDir),
	)

	if err != nil {
		return err
	}

	return nil
}

func createDirectoryContextMenu() error {

	base := `Directory\shell\DvplTools`

	err := createMenuRoot(base)

	if err != nil {
		return err
	}

	err = createCommand(
		base+`\shell\01Compress`,
		`dvpl -c "Compress"`,
		fmt.Sprintf(`"%s\dvpl.exe" -c -trust-data -i "%%1"`, InstallDir),
	)

	if err != nil {
		return err
	}

	err = createCommand(
		base+`\shell\02Decompress`,
		`dvpl -d "Decompress"`,
		fmt.Sprintf(`"%s\dvpl.exe" -d -trust-data -i "%%1"`, InstallDir),
	)

	if err != nil {
		return err
	}

	return nil
}

func createBackgroundContextMenu() error {

	base := `Directory\Background\shell\DvplTools`

	err := createMenuRoot(base)

	if err != nil {
		return err
	}

	err = createCommand(
		base+`\shell\01Compress`,
		`dvpl -c "Compress"`,
		fmt.Sprintf(`"%s\dvpl.exe" -c -trust-data -i "%%V"`, InstallDir),
	)

	if err != nil {
		return err
	}

	err = createCommand(
		base+`\shell\02Decompress`,
		`dvpl -d "Decompress"`,
		fmt.Sprintf(`"%s\dvpl.exe" -d -trust-data -i "%%V"`, InstallDir),
	)

	if err != nil {
		return err
	}

	return nil
}

func createMenuRoot(path string) error {

	key, _, err := registry.CreateKey(
		registry.CLASSES_ROOT,
		path,
		registry.ALL_ACCESS,
	)

	if err != nil {
		return err
	}

	defer key.Close()

	err = key.SetStringValue("MUIVerb", "Dvpl Tools Private")

	if err != nil {
		return err
	}

	err = key.SetStringValue("SubCommands", "")

	if err != nil {
		return err
	}

	err = key.SetStringValue(
		"Icon",
		filepath.Join(InstallDir, "dvpl_go.ico"),
	)

	if err != nil {
		return err
	}

	err = key.SetStringValue("Position", "Top")

	if err != nil {
		return err
	}

	return nil
}

func createCommand(path string, title string, command string) error {

	key, _, err := registry.CreateKey(
		registry.CLASSES_ROOT,
		path,
		registry.ALL_ACCESS,
	)

	if err != nil {
		return err
	}

	err = key.SetStringValue("MUIVerb", title)

	if err != nil {
		key.Close()
		return err
	}

	err = key.SetStringValue(
		"Icon",
		filepath.Join(InstallDir, "dvpl_go.ico"),
	)

	if err != nil {
		key.Close()
		return err
	}

	key.Close()

	cmdKey, _, err := registry.CreateKey(
		registry.CLASSES_ROOT,
		path+`\command`,
		registry.ALL_ACCESS,
	)

	if err != nil {
		return err
	}

	defer cmdKey.Close()

	err = cmdKey.SetStringValue("", command)

	if err != nil {
		return err
	}

	return nil
}

func deleteRegistryTree(root registry.Key, path string) {

	sub, err := registry.OpenKey(
		root,
		path,
		registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE,
	)

	if err == nil {

		names, _ := sub.ReadSubKeyNames(-1)

		sub.Close()

		for _, name := range names {
			deleteRegistryTree(root, path+`\`+name)
		}
	}

	registry.DeleteKey(root, path)
}
