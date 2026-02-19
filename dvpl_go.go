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
	"strconv"
	"strings"
	"sync"

	"github.com/eiannone/keyboard"
)

const (
	DvplExt = ".dvpl"
	DpvlInf = "dvpl_go 2.0.0 x64 | Copyright (c) 2026 Qirashi"
)

func main() {
	compressFlag := flag.Bool("c", false, "Compress .dvpl files.")
	decompressFlag := flag.Bool("d", false, "Decompress .dvpl files.")
	inputPath := flag.String("i", "", "Input path. (file or directory)")
	outputPath := flag.String("o", "", "Output path. (file or directory)")
	keepOriginal := flag.Bool("keep-original", false, "Keep original files.")
	compressType := flag.Int("compress", 1, "Compression type: 0 (none), 1 (lz4hc), 2 (lz4) |")
	ignorePatterns := flag.String("ignore", "", "List of file patterns to ignore. (\"*.exe,*.dll\")")
	filterPatterns := flag.String("filter", "", "List of file patterns to include. (\"*.sc2,*.scg\")")
	ignoreCompressPatterns := flag.String("ignore-compress", "", "List of file patterns for which compression should be disabled. (\"*.webp\")")
	forcedCompress := flag.Bool("forced-compress", false, "Force compression even if the result is larger than the original.")
	maxWorkers := flag.Int("m", 2, fmt.Sprintf("Maximum number of parallel workers (%d). Minimum 1, recommended 2 |", runtime.NumCPU()))
	skipCRC := flag.Bool("skip-crc", false, "CRC can be ignored when unpacking or packing.")

	flag.Usage = func() {
		fmt.Printf("\n%s\n\nUsage: dvpl [options]\n[Options]:\n", DpvlInf)
		flag.PrintDefaults()
		fmt.Println(`
Examples:
  Compress   : dvpl -c -i ./in_dir -o ./out_dir
  Decompress : dvpl -d -i ./in_dir -o ./out_dir
  Ignore     : dvpl -c -i ./in_dir -ignore "*.exe,*.dll"
  Filter     : dvpl -d -i ./in_dir -o ./out_dir -filter "*.sc2,*.scg"
  No compress: dvpl -c -i ./in_dir -ignore-compress "*.webp"
  Compression: dvpl -c -i ./in_dir -compress 2`)
	}

	if envMaxWorkers := os.Getenv("DVPL_MAX_WORKERS"); envMaxWorkers != "" {
		if val, err := strconv.Atoi(envMaxWorkers); err == nil {
			fmt.Println("[info] DVPL_MAX_WORKERS:", val)
			*maxWorkers = val
		}
	}

	if envCompress := os.Getenv("DVPL_COMPRESS_TYPE"); envCompress != "" {
		if val, err := strconv.Atoi(envCompress); err == nil {
			fmt.Println("[info] DVPL_COMPRESS_TYPE:", val)
			*compressType = val
		}
	}

	if len(os.Args) == 1 {
		interactiveMode(*maxWorkers)
		return
	}

	if len(os.Args) > 1 {
		if !strings.HasPrefix(os.Args[1], "-") {
			dragAndDropMode(os.Args[1:], *maxWorkers, *compressType)
			return
		}
	}

	flag.Parse()

	if (*compressFlag && *decompressFlag) || (!*compressFlag && !*decompressFlag) {
		fmt.Println("[debug] Specify either compression (-c) or decompression (-d)")
		flag.Usage()
		return
	}

	if *inputPath == "" {
		flag.Usage()
		return
	}

	if *outputPath == "" {
		*outputPath = *inputPath
	}

	var ignoreList []string
	if *ignorePatterns != "" {
		ignoreList = strings.Split(*ignorePatterns, ",")
	}

	var ignoreCompressList []string
	if *ignoreCompressPatterns != "" {
		ignoreCompressList = strings.Split(*ignoreCompressPatterns, ",")
	}

	var filterList []string
	if *filterPatterns != "" {
		filterList = strings.Split(*filterPatterns, ",")
		for i := range filterList {
			filterList[i] = strings.TrimSpace(filterList[i])
		}
	}

	if *compressFlag {
		processFiles(*inputPath, *outputPath, Pack, *keepOriginal, *compressFlag, *compressType, ignoreList, ignoreCompressList, filterList, *maxWorkers, *forcedCompress, *skipCRC)
	} else if *decompressFlag {
		processFiles(*inputPath, *outputPath, Unpack, *keepOriginal, *compressFlag, *compressType, ignoreList, ignoreCompressList, filterList, *maxWorkers, *forcedCompress, *skipCRC)
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
		return fmt.Errorf("input file too large for LZ4 compression: %d bytes (max %d)", fileSize, 0x7E000000)
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

	if err := os.WriteFile(outputPath+DvplExt, dvplData, 0644); err != nil {
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

func matchesAnyPattern(name string, patterns []string) bool {
	for _, p := range patterns {
		if p = strings.TrimSpace(p); p == "" {
			continue
		}

		if ok, err := filepath.Match(p, name); err != nil {
			fmt.Printf("[info] invalid pattern %q: %v\n", p, err)
		} else if ok {
			return true
		}
	}
	return false
}

func shouldProcessFile(path string, info os.FileInfo, exeFileName string, compressFlag bool, ignorePatterns, filterPatterns []string) bool {
	name := info.Name()

	if compressFlag && name == exeFileName {
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

	if len(filterPatterns) > 0 {
		filterName := name
		if !compressFlag {
			filterName = strings.TrimSuffix(name, DvplExt)
		}

		if !matchesAnyPattern(filterName, filterPatterns) {
			fmt.Printf("Filter skip: %s\n", path)
			return false
		}
	}

	return true
}

func processFiles(inputPath, outputPath string,
	processor func(string, string, int, bool, bool) error,
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

	maxCPU := runtime.NumCPU()
	if maxWorkers < 1 || maxWorkers > maxCPU {
		fmt.Printf("[info] maxWorkers value has been changed from %d to %d\n", maxWorkers, maxCPU)
		maxWorkers = maxCPU
	}

	tasks := make(chan task, maxWorkers*2)
	errorsCh := make(chan error, maxWorkers*2)
	var wg sync.WaitGroup

	for range maxWorkers {
		wg.Add(1)
		go worker(tasks, errorsCh, &wg)
	}

	var errList []error
	var errMu sync.Mutex
	var collectWg sync.WaitGroup
	collectWg.Go(func() {
		for err := range errorsCh {
			errMu.Lock()
			errList = append(errList, err)
			errMu.Unlock()
		}
	})

	effectiveCompress := func(name string) int {
		if matchesAnyPattern(name, ignoreCompressPatterns) {
			return 0
		}
		return compressType
	}

	addTask := func(path string, info os.FileInfo) {
		relativePath, _ := filepath.Rel(inputPath, path)
		outPath := filepath.Join(outputPath, relativePath)

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
	for tsk := range tasks {
		if err := tsk.processor(tsk.path, tsk.outPath, tsk.compressType, tsk.forcedCompress, tsk.skipCRC); err != nil {
			errors <- fmt.Errorf("processing file %s: %v", tsk.path, err)
		} else if !tsk.keepOriginal {
			if err := os.Remove(tsk.path); err != nil {
				errors <- fmt.Errorf("removing original file %s: %v", tsk.path, err)
			}
		}
	}
}

func detectDirMode(root string) bool {
	unpack := false

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		if strings.HasSuffix(d.Name(), DvplExt) {
			unpack = true
			return fmt.Errorf("stop walk")
		}
		return nil
	})

	return unpack
}

func dragAndDropMode(paths []string, maxWorkers int, compressType int) {
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		if info.IsDir() {
			if detectDirMode(path) {
				processFiles(path, path, Unpack, false, false, 0, nil, nil, nil, maxWorkers, false, true)
			} else {
				processFiles(path, path, Pack, false, true, compressType, nil, nil, nil, maxWorkers, false, true)
			}
			continue
		}

		if strings.HasSuffix(info.Name(), DvplExt) {
			processFiles(path, path, Unpack, false, false, 0, nil, nil, nil, maxWorkers, false, true)
		} else {
			processFiles(path, path, Pack, false, true, compressType, nil, nil, nil, maxWorkers, false, true)
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

func interactiveMode(maxWorkers int) {
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
				compressInteractive(maxWorkers)
			case 1:
				decompressInteractive(maxWorkers)
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

func compressInteractive(maxWorkers int) {
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
			processFiles(".", ".", Pack, false, true, selectedCompressionType, nil, nil, nil, maxWorkers, false, true)
			return
		case keyboard.KeyEsc:
			return
		}
	}
}

func decompressInteractive(maxWorkers int) {
	fmt.Println("Start decompressing...")
	processFiles(".", ".", Unpack, false, false, 0, nil, nil, nil, maxWorkers, false, true)
}
