// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Qirashi
// Project: dvpl_go

package main

import (
	// dvpl "dvpl_go/dvpl"
	dvpl "dvpl_go/dvpl_c"

	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	compressType   int      `yaml:"compress"`
	IgnorePatterns []string `yaml:"ignorePatterns"`
	maxWorkers     int      `yaml:"maxWorkers"`
}

func main() {
	compressFlag   := flag.Bool("c", false, "Compress .dvpl files")
	decompressFlag := flag.Bool("d", false, "Decompress .dvpl files")
	inputPath      := flag.String("i", "", "Input path (file or directory)")
	outputPath     := flag.String("o", "", "Output path (file or directory)")
	keepOriginal   := flag.Bool("keep-original", false, "Keep original files")
	compressType   := flag.Int("compress", 1, "Compression type: 0 (none), 1 (lz4hc), 2 (lz4) |")
	ignorePatterns := flag.String("ignore", "", "Comma-separated list of file patterns to ignore")
	maxWorkers     := flag.Int("m", 1, "Maximum number of parallel workers. When used, 2 are recommended, with a maximum of 6.")

	flag.Usage = func() {
		fmt.Println(`Usage: dvpl_go [options]
Options:`)
		flag.PrintDefaults()
		fmt.Println(`
Examples:
  Compress   : dvpl_go -c -i ./input_dir -o ./output_dir
  Decompress : dvpl_go -d -i ./input_dir -o ./output_dir
  Ignore     : dvpl_go -c -i ./input_dir -ignore "*.exe,*.dll"
  Compression: dvpl_go -c -i ./input_dir -compress 2`)
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
		if config.compressType != 0 {
			*compressType = config.compressType
		}
		if config.maxWorkers != 0 {
			*maxWorkers = config.maxWorkers
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

	debugPrintFlags(config, *compressFlag, *decompressFlag, *inputPath, *outputPath, *keepOriginal, *compressType, ignoreList, *maxWorkers)

	if *compressFlag {
		processFiles(*inputPath, *outputPath, dvpl.Pack, ".dvpl", *keepOriginal, *compressFlag, *compressType, ignoreList, *maxWorkers)
	} else if *decompressFlag {
		processFiles(*inputPath, *outputPath, dvpl.Unpack, "", *keepOriginal, *compressFlag, *compressType, ignoreList, *maxWorkers)
	}
}

func debugPrintFlags(c *Config, compressFlag, decompressFlag bool, inputPath, outputPath string, 
	keepOriginal bool, compressType int, ignorePatterns []string, maxWorkers int) {

	var flags []string

	// Всегда показывать режим
	if compressFlag {
		flags = append(flags, "-c")
	} else {
		flags = append(flags, "-d")
	}

	// Показывать все остальные параметры, даже если они из конфига
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
	flags = append(flags, fmt.Sprintf("-m %d", maxWorkers))

	// Добавить пометку о конфиге
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

func processFiles(inputPath, outputPath string, processor func(string, string, int) error, newExt string, keepOriginal bool, compressFlag bool, compressType int, ignorePatterns []string, maxWorkers int) {
	info, err := os.Stat(inputPath)
	if err != nil {
		fmt.Printf("[error] Error accessing input path: %v\n", err)
		return
	}

	exeFileName := filepath.Base(os.Args[0])

	if info.IsDir() {
		if maxWorkers <= 1 {
			// Однопоточный режим (оригинальная логика)
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
			// Многопоточный режим
			tasks := make(chan task, maxWorkers*2)
			errors := make(chan error, maxWorkers*2)
			var wg sync.WaitGroup

			// Запускаем worker'ов
			for i := 0; i < maxWorkers; i++ {
				wg.Add(1)
				go worker(tasks, errors, &wg)
			}

			// Собираем ошибки
			go func() {
				for err := range errors {
					fmt.Printf("[error] %v\n", err)
				}
			}()

			// Обход файлов и отправка задач в канал
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

					tasks <- task{
						path:         path,
						outPath:      outPath,
						processor:    processor,
						compressType: compressType,
						keepOriginal: keepOriginal,
					}
				}
				return nil
			})

			close(tasks)
			wg.Wait()
			close(errors)
		}
	} else {
		// Обработка одиночного файла (без изменений)
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

type task struct {
	path         string
	outPath      string
	processor    func(string, string, int) error
	compressType int
	keepOriginal bool
}

func worker(tasks <-chan task, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		if err := task.processor(task.path, task.outPath, task.compressType); err != nil {
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
			processFiles(".", ".", dvpl.Pack, ".dvpl", false, true, selectedCompressionType, nil, 1)
			return
		}
	}
}

func decompressInteractive() {
	fmt.Println("Decompressing files in current directory...")
	processFiles(".", ".", dvpl.Unpack, "", false, false, 0, nil, 1)
}
