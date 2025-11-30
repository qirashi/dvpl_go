// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Qirashi
// Project: dvpl_go

package main

import (
	dvpl "dvpl_go/dvpl_c"
	"io/fs"

	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/eiannone/keyboard"
	"gopkg.in/yaml.v3"
)

type Config struct {
	CompressFlag   bool     `yaml:"compressFlag"`
	DecompressFlag bool     `yaml:"decompressFlag"`
	InputPath      string   `yaml:"inputPath"`
	OutputPath     string   `yaml:"outputPath"`
	KeepOriginal   bool     `yaml:"keepOriginal"`
	CompressType   int      `yaml:"compress"`
	IgnorePatterns []string `yaml:"ignorePatterns"`
	FilterPatterns []string `yaml:"filterPatterns"`
	IgnoreCompress []string `yaml:"ignoreCompress"`
	ForcedCompress bool     `yaml:"forcedCompress"`
	MaxWorkers     int      `yaml:"maxWorkers"`
}

func main() {
	maxCPU := runtime.NumCPU()
	
	compressFlag   := flag.Bool("c", false, "Compress .dvpl files")
	decompressFlag := flag.Bool("d", false, "Decompress .dvpl files")
	inputPath      := flag.String("i", "", "Input path (file or directory)")
	outputPath     := flag.String("o", "", "Output path (file or directory)")
	keepOriginal   := flag.Bool("keep-original", false, "Keep original files")
	compressType   := flag.Int("compress", 1, "Compression type: 0 (none), 1 (lz4hc), 2 (lz4) |")
	ignorePatterns := flag.String("ignore", "", "Comma-separated list of file patterns to ignore")
	filterPatterns := flag.String("filter", "", "Comma-separated list of file patterns to include (e.g. \"*.sc2,*.scg\")")
	ignoreCompress := flag.String("ignore-compress", "", "Comma-separated list of file patterns to force no compression (type 0)")
	forcedCompress := flag.Bool("forced-compress", false, "Forced compression, even if the result is larger than the original")
	maxWorkers     := flag.Int("m", 1, fmt.Sprintf("Maximum number of parallel workers (%d). Minimum 1, recommended 2.", maxCPU))

	flag.Usage = func() {
		fmt.Println(`Usage: dvpl [options]
Options:`)
		flag.PrintDefaults()
		fmt.Println(`
Examples:
  Compress   : dvpl -c -i ./input_dir -o ./output_dir
  Decompress : dvpl -d -i ./input_dir -o ./output_dir
  Ignore     : dvpl -c -i ./input_dir -ignore "*.exe,*.dll"
  Filter     : dvpl -d -i ./input_dir -o ./output_dir -filter "*.sc2,*.scg"
  No compress: dvpl -c -i ./input_dir -ignore-compress "*.webp"
  Compression: dvpl -c -i ./input_dir -compress 2`)
	}

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
		if config.CompressType != 0 {
			*compressType = config.CompressType
		}
		if config.ForcedCompress {
			*forcedCompress = true
		}
		if config.MaxWorkers != 0 {
			*maxWorkers = config.MaxWorkers
		}
		if len(config.IgnorePatterns) > 0 {
			*ignorePatterns = strings.Join(config.IgnorePatterns, ",")
		}
		if len(config.FilterPatterns) > 0 {
			*filterPatterns = strings.Join(config.FilterPatterns, ",")
		}
		if len(config.IgnoreCompress) > 0 {
			*ignoreCompress = strings.Join(config.IgnoreCompress, ",")
		}
	}

	flag.Parse()

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

	var ignoreCompressList []string
	if *ignoreCompress != "" {
		ignoreCompressList = strings.Split(*ignoreCompress, ",")
	}

	var filterList []string
	if *filterPatterns != "" {
		filterList = strings.Split(*filterPatterns, ",")
		for i := range filterList {
			filterList[i] = strings.TrimSpace(filterList[i])
		}
	}

	if *maxWorkers < 1 {
		*maxWorkers = 1
	} else if *maxWorkers > maxCPU {
		fmt.Printf("[info]  maxWorkers value %d is too high, using maximum %d\n", *maxWorkers, maxCPU)
		*maxWorkers = maxCPU
	}

	debugPrintFlags(config, *compressFlag, *inputPath, *outputPath, *keepOriginal, *compressType, ignoreList, ignoreCompressList, filterList, *maxWorkers)

	if *compressFlag {
		processFiles(*inputPath, *outputPath, Pack, ".dvpl", *keepOriginal, *compressFlag, *compressType, ignoreList, ignoreCompressList, filterList, *maxWorkers, *forcedCompress)
	} else if *decompressFlag {
		processFiles(*inputPath, *outputPath, Unpack, "", *keepOriginal, *compressFlag, *compressType, ignoreList, ignoreCompressList, filterList, *maxWorkers, *forcedCompress)
	}
}

func getCompressionTypeString(compressionType uint32) string {
	switch compressionType {
	case 0:
		return "none"
	case 1:
		return "lz4hc"
	case 2:
		return "lz4"
	case 3:
		return "rfc1951"
	default:
		return fmt.Sprintf("unknown(%d)", compressionType)
	}
}

func Pack(inputPath, outputPath string, compressType int, forcedCompress bool) error {

	data, err := os.ReadFile(inputPath)

	if err != nil {
		return fmt.Errorf("[error] Failed to read input file: %v", err)
	}

	dvplData, compressionType, err := dvpl.Pack(data, compressType, forcedCompress)
	if err != nil {
		return fmt.Errorf("[error] Failed to pack data: %v", err)
	}

	fmt.Printf("Pack [%s]: %s\n", getCompressionTypeString(compressionType), inputPath)

	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return fmt.Errorf("[error] Failed to create output directory: %v", err)
	}

	if err := os.WriteFile(outputPath, dvplData, 0644); err != nil {
		return fmt.Errorf("[error] Failed to write output file: %v", err)
	}

	return nil
}

func Unpack(inputPath, outputPath string, _ int, _ bool) error {

	dvplData, err := os.ReadFile(inputPath)

	if err != nil {
		return fmt.Errorf("[error] Failed to read input file: %v", err)
	}

	data, compressionType, err := dvpl.Unpack(dvplData)
	if err != nil {
		return fmt.Errorf("[error] Failed to unpack data: %v", err)
	}

	fmt.Printf("Unpack [%s]: %s\n", getCompressionTypeString(compressionType), inputPath)

	outputPath = strings.TrimSuffix(outputPath, ".dvpl")

	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return fmt.Errorf("[error] Failed to create output directory: %v", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("[error] Failed to write output file: %v", err)
	}

	return nil
}

func debugPrintFlags(c *Config, compressFlag bool, inputPath, outputPath string, 
	keepOriginal bool, compressType int, ignorePatterns, ignoreCompressPatterns, filterPatterns []string, maxWorkers int) {

	var flags []string

	if compressFlag {
		flags = append(flags, "-c")
	} else {
		flags = append(flags, "-d")
	}

	flags = append(flags, fmt.Sprintf("-i \"%s\"", inputPath))
	if outputPath != inputPath {
		flags = append(flags, fmt.Sprintf("-o \"%s\"", outputPath))
	}
	if keepOriginal {
		flags = append(flags, "-keep-original")
	}
	flags = append(flags, fmt.Sprintf("-compress %d", compressType))
	if len(ignorePatterns) > 0 {
		flags = append(flags, fmt.Sprintf("-ignore \"%s\"", strings.Join(ignorePatterns, ",")))
	}
	if len(ignoreCompressPatterns) > 0 {
		flags = append(flags, fmt.Sprintf("-ignore-compress \"%s\"", strings.Join(ignoreCompressPatterns, ",")))
	}
	if len(filterPatterns) > 0 {
		flags = append(flags, fmt.Sprintf("-filter \"%s\"", strings.Join(filterPatterns, ",")))
	}
	flags = append(flags, fmt.Sprintf("-m %d", maxWorkers))

	source := "cmd"
	if c != nil {
		source = "cfg"
	}
	fmt.Printf("[debug] [%s] [%s]\n", strings.Join(flags, " "), source)
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

func shouldProcessFile(path string, info os.FileInfo, exeFileName string, compressFlag bool, ignorePatterns, filterPatterns []string) bool {
	if compressFlag && (info.Name() == ".dvpl_go.yml" || info.Name() == exeFileName) {
		fmt.Printf("Excluding file: %s\n", path)
		return false
	}

	if compressFlag && strings.HasSuffix(info.Name(), ".dvpl") {
		fmt.Printf("Skip .dvpl file: %s\n", path)
		return false
	}

	if !compressFlag && !strings.HasSuffix(info.Name(), ".dvpl") {
		fmt.Printf("Skip file: %s\n", path)
		return false
	}

	for _, pattern := range ignorePatterns {
		if matched, _ := filepath.Match(pattern, info.Name()); matched {
			fmt.Printf("Ignoring file: %s\n", path)
			return false
		}
	}

	if len(filterPatterns) > 0 {
		nameToMatch := info.Name()
		if !compressFlag {
			nameToMatch = strings.TrimSuffix(nameToMatch, ".dvpl")
		}
		matchedAny := false
		for _, pattern := range filterPatterns {
			if matched, _ := filepath.Match(strings.TrimSpace(pattern), nameToMatch); matched {
				matchedAny = true
				break
			}
		}
		if !matchedAny {
			fmt.Printf("Filter skip: %s\n", path)
			return false
		}
	}

	return true
}

func processFiles(inputPath, outputPath string,
	processor func(string, string, int, bool) error,
	newExt string,
	keepOriginal bool,
	compressFlag bool,
	compressType int,
	ignorePatterns []string,
	ignoreCompressPatterns []string,
	filterPatterns []string,
	maxWorkers int,
	forcedCompress bool) {

	info, err := os.Stat(inputPath)
	if err != nil {
		fmt.Printf("[error] Error accessing input path: %v\n", err)
		return
	}

	exeFileName := filepath.Base(os.Args[0])

	if info.IsDir() {
		if maxWorkers <= 1 {
			// Однопоточный режим
			filepath.WalkDir(inputPath, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					fmt.Printf("[error] Error accessing path %s: %v\n", path, err)
					return nil
				}

				if !d.IsDir() {
					info, _ := d.Info() // получаем os.FileInfo только при необходимости
					if !shouldProcessFile(path, info, exeFileName, compressFlag, ignorePatterns, filterPatterns) {
						return nil
					}

					relativePath, _ := filepath.Rel(inputPath, path)
					outPath := filepath.Join(outputPath, relativePath)
					if newExt != "" {
						outPath += newExt
					}

					actualCompressType := compressType
					for _, pattern := range ignoreCompressPatterns {
						if matched, _ := filepath.Match(pattern, info.Name()); matched {
							actualCompressType = 0
							break
						}
					}

					if err := processor(path, outPath, actualCompressType, forcedCompress); err != nil {
						fmt.Printf("[error] Error processing file %s: %v\n", path, err)
					} else if !keepOriginal {
						os.Remove(path)
					}
				}
				return nil
			})
		} else {
			// Многопоточный режим
			tasks := make(chan task, maxWorkers*2)
			errors := make(chan error, maxWorkers*2)
			var wg sync.WaitGroup

			for i := 0; i < maxWorkers; i++ {
				wg.Add(1)
				go worker(tasks, errors, &wg)
			}

			go func() {
				for err := range errors {
					fmt.Printf("[error] %v\n", err)
				}
			}()

			filepath.WalkDir(inputPath, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					fmt.Printf("[error] Error accessing path %s: %v\n", path, err)
					return nil
				}

				if !d.IsDir() {
					info, _ := d.Info()
					if !shouldProcessFile(path, info, exeFileName, compressFlag, ignorePatterns, filterPatterns) {
						return nil
					}

					relativePath, _ := filepath.Rel(inputPath, path)
					outPath := filepath.Join(outputPath, relativePath)
					if newExt != "" {
						outPath += newExt
					}

					ignoreCompress := false
					for _, pattern := range ignoreCompressPatterns {
						if matched, _ := filepath.Match(pattern, info.Name()); matched {
							ignoreCompress = true
							break
						}
					}

					tasks <- task{
						path:           path,
						outPath:        outPath,
						processor:      processor,
						compressType:   compressType,
						ignoreCompress: ignoreCompress,
						keepOriginal:   keepOriginal,
						forcedCompress: forcedCompress,
					}
				}
				return nil
			})

			close(tasks)
			wg.Wait()
			close(errors)
		}
	} else {
		// Обработка одиночного файла
		if compressFlag && strings.HasSuffix(inputPath, ".dvpl") {
			fmt.Printf("Skipping .dvpl file: %s\n", inputPath)
			return
		}

		outPath := outputPath
		if newExt != "" {
			outPath += newExt
		}

		if len(filterPatterns) > 0 {
			baseName := filepath.Base(inputPath)
			nameToMatch := baseName
			if !compressFlag {
				nameToMatch = strings.TrimSuffix(baseName, ".dvpl")
			}
			matchedAny := false
			for _, pattern := range filterPatterns {
				if matched, _ := filepath.Match(strings.TrimSpace(pattern), nameToMatch); matched {
					matchedAny = true
					break
				}
			}
			if !matchedAny {
				fmt.Printf("Filter skip: %s\n", inputPath)
				return
			}
		}

		actualCompressType := compressType
		info, _ := os.Stat(inputPath)
		for _, pattern := range ignoreCompressPatterns {
			if matched, _ := filepath.Match(pattern, info.Name()); matched {
				actualCompressType = 0
				break
			}
		}

		if err := processor(inputPath, outPath, actualCompressType, forcedCompress); err != nil {
			fmt.Printf("[error] Error processing file %s: %v\n", inputPath, err)
		} else if !keepOriginal {
			os.Remove(inputPath)
		}
	}

	fmt.Println("\nOperation completed!")
}

type task struct {
	path           string
	outPath        string
	processor      func(string, string, int, bool) error
	compressType   int
	ignoreCompress bool
	keepOriginal   bool
	forcedCompress bool
}

func worker(tasks <-chan task, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		actualCompressType := task.compressType
		if task.ignoreCompress {
			actualCompressType = 0
		}

		if err := task.processor(task.path, task.outPath, actualCompressType, task.forcedCompress); err != nil {
			errors <- fmt.Errorf("processing file %s: %v", task.path, err)
		} else if !task.keepOriginal {
			if err := os.Remove(task.path); err != nil {
				errors <- fmt.Errorf("removing original file %s: %v", task.path, err)
			}
		}
	}
}

func interactiveMode() {
	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		fmt.Printf("[error] Failed to initialize keyboard: %v\n", err)
		return
	}
	defer keyboard.Close()

	options := []string{"Compress", "Decompress", "Help", "Exit"}
	selectedIndex := 0

	for {
		// Очистка экрана
		fmt.Print("\033[H\033[2J")
		fmt.Println("Usage: dvpl_go [-h] - To get help.")

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
			fmt.Println("\nYou selected:", options[selectedIndex])
			switch selectedIndex {
			case 0: // Compress
				compressInteractive()
				return
			case 1: // Decompress
				decompressInteractive()
				return
			case 2: // Help
				flag.Usage()
				<-keysEvents
				return
			case 3: // Exit
				fmt.Println("Exiting...")
				return
			}
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

	options := []string{"0. none", "1. lz4hc", "2. lz4"}
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
			processFiles(".", ".", Pack, ".dvpl", false, true, selectedCompressionType, nil, nil, nil, 1, false)
			return
		}
	}
}

func decompressInteractive() {
	fmt.Println("Decompressing files in current directory...")
	processFiles(".", ".", Unpack, "", false, false, 0, nil, nil, nil, 1, false)
}
