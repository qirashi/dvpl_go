// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Qirashi
// Project: dvpl_go

package main

import (
	dvpl "dvpl_go/dvpl_c"

	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/eiannone/keyboard"
	"gopkg.in/yaml.v2"
)

const (
	DvplExt = ".dvpl"
	DpvlInf = "dvpl_go 1.3.5 x64 | Copyright (c) 2026 Qirashi"
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
	SkipCRC        bool     `yaml:"skipCRC"`
}

func main() {
	maxCPU := runtime.NumCPU()

	compressFlag := flag.Bool("c", false, "Compress .dvpl files")
	decompressFlag := flag.Bool("d", false, "Decompress .dvpl files")
	inputPath := flag.String("i", "", "Input path (file or directory)")
	outputPath := flag.String("o", "", "Output path (file or directory)")
	keepOriginal := flag.Bool("keep-original", false, "Keep original files")
	compressType := flag.Int("compress", 1, "Compression type: 0 (none), 1 (lz4hc), 2 (lz4) |")
	ignorePatterns := flag.String("ignore", "", "Comma-separated list of file patterns to ignore")
	filterPatterns := flag.String("filter", "", "Comma-separated list of file patterns to include (e.g. \"*.sc2,*.scg\")")
	ignoreCompress := flag.String("ignore-compress", "", "Comma-separated list of file patterns to force no compression (type 0)")
	forcedCompress := flag.Bool("forced-compress", false, "Forced compression, even if the result is larger than the original")
	maxWorkers := flag.Int("m", 1, fmt.Sprintf("Maximum number of parallel workers (%d). Minimum 1, recommended 2.", maxCPU))
	skipCRC := flag.Bool("skip-crc", false, "When unpacking, the crc will be ignored.")

	flag.Usage = func() {
		fmt.Printf("\n%s\n\nUsage: dvpl [options]\n[Options]:\n", DpvlInf)
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
		if config.SkipCRC {
			*skipCRC = true
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
		processFiles(*inputPath, *outputPath, Pack, DvplExt, *keepOriginal, *compressFlag, *compressType, ignoreList, ignoreCompressList, filterList, *maxWorkers, *forcedCompress, *skipCRC)
	} else if *decompressFlag {
		processFiles(*inputPath, *outputPath, Unpack, "", *keepOriginal, *compressFlag, *compressType, ignoreList, ignoreCompressList, filterList, *maxWorkers, *forcedCompress, *skipCRC)
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

func Pack(inputPath, outputPath string, compressType int, forcedCompress bool, skipCRC bool) error {
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	fileSize := fileInfo.Size()

	if fileSize > 1<<32-1 {
		return fmt.Errorf("input file too large: %d bytes (max %d)", fileSize, 1<<32-1)
	}

	if (compressType == 1 || compressType == 2) && fileSize > 0x7E000000 {
		return fmt.Errorf("input file too large for raw LZ4 compression: %d bytes (max %d)", fileSize, 0x7E000000)
	}

	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}

	dvplData, compressionType, err := dvpl.Pack(data, compressType, forcedCompress, skipCRC)
	if err != nil {
		return fmt.Errorf("failed to pack data: %v", err)
	}

	fmt.Printf("Pack [%s]: %s\n", getCompressionTypeString(compressionType), inputPath)

	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	if err := os.WriteFile(outputPath, dvplData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	return nil
}

func Unpack(inputPath, outputPath string, _ int, _ bool, skipCRC bool) error {

	dvplData, err := os.ReadFile(inputPath)

	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}

	data, compressionType, err := dvpl.Unpack(dvplData, skipCRC)
	if err != nil {
		return fmt.Errorf("failed to unpack data: %v", err)
	}

	fmt.Printf("Unpack [%s]: %s\n", getCompressionTypeString(compressionType), inputPath)

	outputPath = strings.TrimSuffix(outputPath, DvplExt)

	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
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
		return nil
	}

	fmt.Println("[debug] Configuration loaded: .dvpl_go.yml")

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

func nameForFilter(baseName string, compress bool) string {
	if compress {
		return baseName
	}
	return strings.TrimSuffix(baseName, DvplExt)
}

func matchesAnyPattern(name string, patterns []string) bool {
	for _, p := range patterns {
		if matched, _ := filepath.Match(strings.TrimSpace(p), name); matched {
			return true
		}
	}
	return false
}

func matchesFilter(nameToMatch string, filterPatterns []string) bool {
	if len(filterPatterns) == 0 {
		return true
	}
	return matchesAnyPattern(nameToMatch, filterPatterns)
}

func shouldProcessFile(path string, info os.FileInfo, exeFileName string, compressFlag bool, ignorePatterns, filterPatterns []string) bool {
	name := info.Name()

	if compressFlag && (name == ".dvpl_go.yml" || name == exeFileName) {
		fmt.Printf("Excluding file: %s\n", path)
		return false
	}
	if compressFlag && strings.HasSuffix(name, DvplExt) {
		return false
	}
	if !compressFlag && !strings.HasSuffix(name, DvplExt) {
		return false
	}
	if matchesAnyPattern(name, ignorePatterns) {
		fmt.Printf("Ignoring file: %s\n", path)
		return false
	}
	if !matchesFilter(nameForFilter(name, compressFlag), filterPatterns) {
		fmt.Printf("Filter skip: %s\n", path)
		return false
	}
	return true
}

func processFiles(inputPath, outputPath string,
	processor func(string, string, int, bool, bool) error,
	newExt string,
	keepOriginal bool,
	compressFlag bool,
	compressType int,
	ignorePatterns []string,
	ignoreCompressPatterns []string,
	filterPatterns []string,
	maxWorkers int,
	forcedCompress bool,
	skipCRC bool) {

	info, err := os.Stat(inputPath)
	if err != nil {
		fmt.Printf("[error] Error accessing input path: %v\n", err)
		return
	}

	exeFileName := filepath.Base(os.Args[0])
	tasks := make(chan task, maxWorkers*2)
	errorsCh := make(chan error, maxWorkers*2)
	var wg sync.WaitGroup

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go worker(tasks, errorsCh, &wg)
	}

	var errList []error
	var errMu sync.Mutex
	var collectWg sync.WaitGroup
	collectWg.Add(1)
	go func() {
		defer collectWg.Done()
		for err := range errorsCh {
			errMu.Lock()
			errList = append(errList, err)
			errMu.Unlock()
		}
	}()

	effectiveCompress := func(name string) int {
		if matchesAnyPattern(name, ignoreCompressPatterns) {
			return 0
		}
		return compressType
	}

	addTask := func(path string, info os.FileInfo) {
		relativePath, _ := filepath.Rel(inputPath, path)
		outPath := filepath.Join(outputPath, relativePath)
		if newExt != "" {
			outPath += newExt
		}
		tasks <- task{
			path:           path,
			outPath:        outPath,
			processor:      processor,
			compressType:   effectiveCompress(info.Name()),
			keepOriginal:   keepOriginal,
			forcedCompress: forcedCompress,
			skipCRC:        skipCRC,
		}
	}

	finishAndReturn := func() {
		close(tasks)
		wg.Wait()
		close(errorsCh)
		collectWg.Wait()
		fmt.Println("\nOperation completed!")
		errMu.Lock()
		for _, e := range errList {
			fmt.Printf("[error] %v\n", e)
		}
		errMu.Unlock()
	}

	if info.IsDir() {
		filepath.WalkDir(inputPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("[error] Error accessing path %s: %v\n", path, err)
				return nil
			}
			if d.IsDir() {
				return nil
			}
			fileInfo, _ := d.Info()
			if shouldProcessFile(path, fileInfo, exeFileName, compressFlag, ignorePatterns, filterPatterns) {
				addTask(path, fileInfo)
			}
			return nil
		})
	} else {
		if !shouldProcessFile(inputPath, info, exeFileName, compressFlag, ignorePatterns, filterPatterns) {
			finishAndReturn()
			return
		}
		addTask(inputPath, info)
	}

	finishAndReturn()
}

type task struct {
	path           string
	outPath        string
	processor      func(string, string, int, bool, bool) error
	compressType   int
	keepOriginal   bool
	forcedCompress bool
	skipCRC        bool
}

func worker(tasks <-chan task, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		if err := task.processor(task.path, task.outPath, task.compressType, task.forcedCompress, task.skipCRC); err != nil {
			errors <- fmt.Errorf("processing file %s: %v", task.path, err)
		} else if !task.keepOriginal {
			if err := os.Remove(task.path); err != nil {
				errors <- fmt.Errorf("removing original file %s: %v", task.path, err)
			}
		}
	}
}

func drawMenu(options []string, selected int) {
	for i, option := range options {
		if i == selected {
			fmt.Printf("> %s\n", option)
		} else {
			fmt.Printf("  %s\n", option)
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

	fmt.Print("\033[?25l")       // hide cursor
	defer fmt.Print("\033[?25h") // show cursor
	fmt.Print("\033[2J\033[H")   // clear once

	options := []string{
		"Compress",
		"Decompress",
		"Help",
	}
	selectedIndex := 0

	for {
		fmt.Printf("\033[H%s\n\nUsage: dvpl_go [-h] - To get help.\nPress Ctrl+C or Esc to exit.\n\n", DpvlInf)

		drawMenu(options, selectedIndex)

		event := <-keysEvents
		if event.Err != nil {
			fmt.Printf("[error] Keyboard error: %v\n", event.Err)
			return
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
			fmt.Print("\033[2J\033[H")
			switch selectedIndex {
			case 0:
				compressInteractive()
			case 1:
				decompressInteractive()
			case 2:
				flag.Usage()
				<-keysEvents
			}
			return
		case keyboard.KeyEsc:
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

	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")
	fmt.Print("\033[2J\033[H")

	options := []string{
		"[0] none",
		"[1] lz4hc",
		"[2] lz4",
	}
	compressionTypes := []int{0, 1, 2}
	selectedIndex := 1

	for {
		fmt.Printf("\033[H%s\n\nSelect compression type.\nPress Ctrl+C or Esc to exit.\n\n", DpvlInf)

		drawMenu(options, selectedIndex)

		event := <-keysEvents
		if event.Err != nil {
			fmt.Printf("[error] Keyboard error: %v\n", event.Err)
			return
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
			fmt.Print("\033[2J\033[H")
			selectedCompressionType := compressionTypes[selectedIndex]
			fmt.Println("Start compressing...")
			processFiles(".", ".", Pack, DvplExt, false, true, selectedCompressionType, nil, nil, nil, 1, false, false)
			return
		case keyboard.KeyEsc:
			return
		}
	}
}

func decompressInteractive() {
	fmt.Println("Start decompressing...")
	processFiles(".", ".", Unpack, "", false, false, 0, nil, nil, nil, 1, false, true)
}
