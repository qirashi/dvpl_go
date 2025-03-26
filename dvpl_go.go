// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Qirashi
// Project: dvpl_go

package main

import (
	"dvpl_go/dvpl"
	// "dvpl_go/dvpl_c"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eiannone/keyboard"
	"gopkg.in/yaml.v3"
)

type Config struct {
	CompressFlag   bool     `yaml:"compressFlag"`
	DecompressFlag bool     `yaml:"decompressFlag"`
	InputPath      string   `yaml:"inputPath"`
	OutputPath     string   `yaml:"outputPath"`
	KeepOriginal   bool     `yaml:"keepOriginal"`
	Compress       int      `yaml:"compress"`
	IgnorePatterns []string `yaml:"ignorePatterns"`
}

func main() {
	compressFlag   := flag.Bool("c", false, "")
	decompressFlag := flag.Bool("d", false, "")
	inputPath      := flag.String("i", "", "")
	outputPath     := flag.String("o", "", "")
	keepOriginal   := flag.Bool("keep-original", false, "")
	compressType   := flag.Int("compress", 1, "")
	ignorePatterns := flag.String("ignore", "", "Comma-separated list of file patterns to ignore")

	flag.Usage = func() {
		fmt.Println("Usage: dvpl_go [options]")
		fmt.Println("Options:")
		fmt.Println("-c | -d")
		fmt.Println("	Compress|Decompress .dvpl files")
		fmt.Println("-i string")
		fmt.Println("	Input path (file or directory)")
		fmt.Println("-o string")
		fmt.Println("	Output path (file or directory)")
		fmt.Println("-keep-original")
		fmt.Println("	Keep original files")
		fmt.Println("-compress int")
		fmt.Println("	Compression type 0 (none), 1 (lz4hc), 2 (lz4) | (default 1)")
		fmt.Println("-ignore string")
		fmt.Println("	Comma-separated list of file patterns to ignore")
		fmt.Println("\nExamples:")
		fmt.Println("  Compress   : dvpl_go -c -i ./input_dir -o ./output_dir")
		fmt.Println("  Decompress : dvpl_go -d -i ./input_dir -o ./output_dir")
		fmt.Println("  Ignore     : dvpl_go -c -i ./input_dir -ignore \"*.exe,*.dll\"")
		fmt.Println("  Compression: dvpl_go -c -i ./input_dir -compress 2")
	}

	flag.Parse()

	config := readConfig()
	if config != nil {
		if config.CompressFlag {
			*compressFlag = true
		}
		if config.DecompressFlag {
			*decompressFlag = true
		}
		if config.InputPath != "" {
			*inputPath = config.InputPath
		}
		if config.OutputPath != "" {
			*outputPath = config.OutputPath
		}
		if config.KeepOriginal {
			*keepOriginal = true
		}
		if config.Compress != 0 {
			*compressType = config.Compress
		}
		if len(config.IgnorePatterns) > 0 {
			*ignorePatterns = strings.Join(config.IgnorePatterns, ",")
		}
	}

	if len(os.Args) == 1 {
		interactiveMode()
		return
	}

	if *inputPath == "" {
		flag.Usage()
		return
	}

	if *outputPath == "" {
		*outputPath = *inputPath
	}

	if (*compressFlag && *decompressFlag) || (!*compressFlag && !*decompressFlag) {
		fmt.Println("[debug] Please specify either -c (compress) or -d (decompress)")
		flag.Usage()
		return
	}

	var ignoreList []string
	if *ignorePatterns != "" {
		ignoreList = strings.Split(*ignorePatterns, ",")
	}

	if *compressFlag {
		processFiles(*inputPath, *outputPath, dvpl.Pack, ".dvpl", *keepOriginal, *compressFlag, *compressType, ignoreList)
	} else if *decompressFlag {
		processFiles(*inputPath, *outputPath, dvpl.Unpack, "", *keepOriginal, *compressFlag, *compressType, ignoreList)
	}
}

func readConfig() *Config {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("[error] Error getting executable path: %v\n", err)
		return nil
	}

	configPath := filepath.Join(filepath.Dir(exePath), ".dvpl_go.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("[debug] Config file not found: .dvpl_go.yml")
		return nil
	}

	fmt.Println("[debug] Config file found: .dvpl_go.yml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("[error] Error reading config file: %v\n", err)
		return nil
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		fmt.Printf("[error] Error parsing config file: %v\n", err)
		return nil
	}

	return &config
}

func processFiles(inputPath, outputPath string, processor func(string, string, int) error, newExt string, keepOriginal bool, compressFlag bool, compressType int, ignorePatterns []string) {
	info, err := os.Stat(inputPath)
	if err != nil {
		fmt.Printf("[error] Error accessing input path: %v\n", err)
		return
	}

	exeFileName := filepath.Base(os.Args[0])

	if info.IsDir() {
		filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("[error] Error accessing path %s: %v\n", path, err)
				return nil
			}

			if !info.IsDir() {
				if compressFlag && (info.Name() == ".dvpl_go.yml" || info.Name() == exeFileName) {
					fmt.Printf("Excluding file: %s\n", path)
					return nil
				}

				if compressFlag && strings.HasSuffix(info.Name(), ".dvpl") {
					fmt.Printf("Skip .dvpl file: %s\n", path)
					return nil
				}

				if !compressFlag && !strings.HasSuffix(info.Name(), ".dvpl") {
					fmt.Printf("Skip file: %s\n", path)
					return nil
				}

				for _, pattern := range ignorePatterns {
					if matched, _ := filepath.Match(pattern, info.Name()); matched {
						fmt.Printf("Ignoring file: %s\n", path)
						return nil
					}
				}

				relativePath, _ := filepath.Rel(inputPath, path)
				outPath := filepath.Join(outputPath, relativePath)

				if newExt != "" {
					outPath += newExt
				}

				if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
					fmt.Printf("[error] Error creating output directory: %v\n", err)
					return nil
				}

				if err := processor(path, outPath, compressType); err != nil {
					fmt.Printf("[error] Error processing file %s: %v\n", path, err)
				} else if !keepOriginal {
					os.Remove(path)
				}
			}
			return nil
		})
	} else {
		if compressFlag && strings.HasSuffix(inputPath, ".dvpl") {
			fmt.Printf("Skipping .dvpl file: %s\n", inputPath)
			return
		}

		outPath := outputPath
		if newExt != "" {
			outPath += newExt
		}

		if err := processor(inputPath, outPath, compressType); err != nil {
			fmt.Printf("[error] Error processing file %s: %v\n", inputPath, err)
		} else if !keepOriginal {
			os.Remove(inputPath)
		}
	}

	fmt.Println("\nOperation completed!")
}

func interactiveMode() {
	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		fmt.Printf("[error] Failed to initialize keyboard: %v\n", err)
		return
	}
	defer keyboard.Close()

	options := []string{"Compress", "Decompress"}
	selectedIndex := 0

	for {
		// Очистка экрана
		fmt.Print("\033[H\033[2J")

		// Отображение меню
		for i, option := range options {
			if i == selectedIndex {
				fmt.Printf("> %s\n", option)
			} else {
				fmt.Printf("  %s\n", option)
			}
		}

		// Чтение нажатия клавиши
		event := <-keysEvents
		if event.Err != nil {
			fmt.Printf("[error] Keyboard error: %v\n", event.Err)
			break
		}

		switch event.Key {
		case keyboard.KeyArrowUp:
			selectedIndex--
			if selectedIndex < 0 {
				selectedIndex = len(options) - 1
			}
		case keyboard.KeyArrowDown:
			selectedIndex++
			if selectedIndex >= len(options) {
				selectedIndex = 0
			}
		case keyboard.KeyEnter:
			fmt.Println("\nYou selected:", options[selectedIndex])
			if selectedIndex == 0 {
				compressInteractive()
			} else {
				decompressInteractive()
			}
			return
		}
	}
}

func compressInteractive() {
	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		fmt.Printf("[error] Failed to initialize keyboard: %v\n", err)
		return
	}
	defer keyboard.Close()

	options := []string{"1. None (0)", "2. LZ4HC (1)", "3. LZ4 (2)"}
	compressionTypes := []int{0, 1, 2}
	selectedIndex := 1

	for {

		fmt.Print("\033[H\033[2J")

		for i, option := range options {
			if i == selectedIndex {
				fmt.Printf("> %s\n", option)
			} else {
				fmt.Printf("  %s\n", option)
			}
		}

		event := <-keysEvents
		if event.Err != nil {
			fmt.Printf("[error] Keyboard error: %v\n", event.Err)
			break
		}

		switch event.Key {
		case keyboard.KeyArrowUp:
			selectedIndex--
			if selectedIndex < 0 {
				selectedIndex = len(options) - 1
			}
		case keyboard.KeyArrowDown:
			selectedIndex++
			if selectedIndex >= len(options) {
				selectedIndex = 0
			}
		case keyboard.KeyEnter:
			selectedCompressionType := compressionTypes[selectedIndex]
			fmt.Println("\nYou selected compression type:", options[selectedIndex])
			processFiles(".", ".", dvpl.Pack, ".dvpl", false, true, selectedCompressionType, nil)
			return
		}
	}
}

func decompressInteractive() {
	fmt.Println("Decompressing files in current directory...")
	processFiles(".", ".", dvpl.Unpack, "", false, false, 0, nil)
}
