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
	dvplExt = ".dvpl"
	dvplInf = esc_color_green + "DvplGo 2.1.4 x64" + esc_color_reset + " | " + esc_color_cyan + "Copyright (c) 2026 Qirashi" + esc_color_reset
)

const (
	esc_cursorHide   = "\033[?25l"
	esc_cursorShow   = "\033[?25h"
	esc_screenClear  = "\033[2J"
	esc_cursorHome   = "\033[H"
	esc_clearAndHome = esc_screenClear + esc_cursorHome

	esc_color_reset  = "\033[0m"
	esc_color_cyan   = "\033[36m"
	esc_color_yellow = "\033[33m"
	esc_color_red    = "\033[31m"
	esc_color_green  = "\033[32m"
)

func main() {
	compressFlag := flag.Bool("c", false, "")
	decompressFlag := flag.Bool("d", false, "")
	inputPath := flag.String("i", "", "")
	outputPath := flag.String("o", "", "")
	keepOriginal := flag.Bool("keep-original", false, "")
	compressType := flag.Int("compress", 1, "")
	ignorePatterns := flag.String("ignore", "", "")
	filterPatterns := flag.String("filter", "", "")
	ignoreCompressPatterns := flag.String("ignore-compress", "", "")
	forcedCompress := flag.Bool("forced-compress", false, "")
	maxWorkers := flag.Int("m", 2, "")
	trustData := flag.Bool("trust-data", false, "")

	flag.Usage = func() {
		fmt.Printf(`
%s

Usage: %sdvpl%s %s(-c|-d) -i%s <path> %s[options]%s
%s[main options]%s
  %s-c%s                  Compress files into .dvpl format.
  %s-d%s                  Decompress .dvpl files.
  %s-i%s <path>           Input file or directory.
  %s-o%s <path>           Output file or directory (default: same as -i).

%s[general options]%s
  %s-filter%s <masks>     Process only files matching given patterns, e.g. "*.sc2,*.scg".
  %s-ignore%s <masks>     Skip files matching given patterns, e.g. "*.exe,*.dll".
  %s-keep-original%s      Do not delete original files after processing.
  %s-m%s <number>         Max parallel workers (default 2, max %d).
  %s-trust-data%s         Skip CRC and some integrity checks.

%s[compress options]%s
  %s-compress%s <type>    Compression type: 0=none, 1=lz4hc, 2=lz4, (default 1).
  %s-forced-compress%s    Force compression even if result is larger than original.
  %s-ignore-compress%s <masks>
                      Disable compression for files matching these patterns, e.g. "*.webp".

%s[examples]%s
  Compress   : dvpl -c -i ./input -compress 1
  Decompress : dvpl -d -i ./input -o ./output
  Filter     : dvpl -d -i ./input -o ./output -filter "*.sc2,*.scg"
  Ignore     : dvpl -c -i ./input -ignore "*.exe,*.dll"
`,
			dvplInf,
			esc_color_green, esc_color_reset,
			esc_color_yellow, esc_color_reset,
			esc_color_cyan, esc_color_reset,
			esc_color_cyan, esc_color_reset,
			esc_color_green, esc_color_reset,
			esc_color_green, esc_color_reset,
			esc_color_green, esc_color_reset,
			esc_color_green, esc_color_reset,
			esc_color_cyan, esc_color_reset,
			esc_color_green, esc_color_reset,
			esc_color_green, esc_color_reset,
			esc_color_green, esc_color_reset,
			esc_color_green, esc_color_reset, runtime.NumCPU(),
			esc_color_green, esc_color_reset,
			esc_color_cyan, esc_color_reset,
			esc_color_green, esc_color_reset,
			esc_color_green, esc_color_reset,
			esc_color_green, esc_color_reset,
			esc_color_cyan, esc_color_reset,
		)
	}

	if envMaxWorkers := os.Getenv("DVPL_MAX_WORKERS"); envMaxWorkers != "" {
		if val, err := strconv.Atoi(envMaxWorkers); err == nil {
			*maxWorkers = val
		}
	}

	if envCompress := os.Getenv("DVPL_COMPRESS_TYPE"); envCompress != "" {
		if val, err := strconv.Atoi(envCompress); err == nil {
			*compressType = val
			if *compressType < 0 || *compressType > 5 {
				printWarn("Invalid compression type: %d. Using valid: 1.", *compressType)
				*compressType = 1
			}
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

	if *compressType < 0 || *compressType > 5 {
		printWarn("Invalid compression type: %d. Using valid: 1.", *compressType)
		*compressType = 5
	}

	if (*compressFlag && *decompressFlag) || (!*compressFlag && !*decompressFlag) {
		printError("Specify either compression (-c) or decompression (-d)")
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
		for i := range ignoreList {
			ignoreList[i] = strings.TrimSpace(ignoreList[i])
		}
	}

	var ignoreCompressList []string
	if *ignoreCompressPatterns != "" {
		ignoreCompressList = strings.Split(*ignoreCompressPatterns, ",")
		for i := range ignoreCompressList {
			ignoreCompressList[i] = strings.TrimSpace(ignoreCompressList[i])
		}
	}

	var filterList []string
	if *filterPatterns != "" {
		filterList = strings.Split(*filterPatterns, ",")
		for i := range filterList {
			filterList[i] = strings.TrimSpace(filterList[i])
		}
	}

	if *compressFlag {
		processFiles(*inputPath, *outputPath, Pack, *keepOriginal, *compressFlag, *compressType, ignoreList, ignoreCompressList, filterList, *maxWorkers, *forcedCompress, *trustData)
	} else if *decompressFlag {
		processFiles(*inputPath, *outputPath, Unpack, *keepOriginal, *compressFlag, *compressType, ignoreList, ignoreCompressList, filterList, *maxWorkers, *forcedCompress, *trustData)
	}
}

func printInfo(format string, a ...any) {
	fmt.Printf(esc_color_cyan+"[info]"+esc_color_reset+" "+format+"\n", a...)
}

func printWarn(format string, a ...any) {
	fmt.Printf(esc_color_yellow+"[warn]"+esc_color_reset+" "+format+"\n", a...)
}

func printError(format string, a ...any) {
	fmt.Printf(esc_color_red+"[error]"+esc_color_reset+" "+format+"\n", a...)
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

func Pack(inputPath, outputPath string, compressType int, forcedCompress bool, trustData bool) error {
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

	dvplData, compressionType, err := dvpl.Pack(data, compressType, forcedCompress)
	if err != nil {
		return fmt.Errorf("failed to pack data: %v", err)
	}

	fmt.Printf("Pack %s[%s]%s: %s\n", esc_color_cyan, getCompressionTypeString(compressionType), esc_color_reset, inputPath)

	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	if err := os.WriteFile(outputPath+dvplExt, dvplData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	return nil
}

func Unpack(inputPath, outputPath string, _ int, _ bool, trustData bool) error {

	dvplData, err := os.ReadFile(inputPath)

	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}

	data, compressionType, err := dvpl.Unpack(dvplData, trustData)
	if err != nil {
		return fmt.Errorf("failed to unpack data: %v", err)
	}

	fmt.Printf("Unpack %s[%s]%s: %s\n", esc_color_cyan, getCompressionTypeString(compressionType), esc_color_reset, inputPath)

	outputPath = strings.TrimSuffix(outputPath, dvplExt)

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
			printWarn("Invalid pattern %q: %v", p, err)
		} else if ok {
			return true
		}
	}
	return false
}

func shouldProcessFile(path string, name string, exeFileName string, compressFlag bool, ignorePatterns, filterPatterns []string) bool {
	if compressFlag && name == exeFileName {
		printInfo("Excluding file: %s", path)
		return false
	}

	if compressFlag && strings.HasSuffix(name, dvplExt) {
		return false
	}

	if !compressFlag && !strings.HasSuffix(name, dvplExt) {
		return false
	}

	if matchesAnyPattern(name, ignorePatterns) {
		printInfo("Ignoring file: %s", path)
		return false
	}

	if len(filterPatterns) > 0 {
		filterName := name
		if !compressFlag {
			filterName = strings.TrimSuffix(name, dvplExt)
		}

		if !matchesAnyPattern(filterName, filterPatterns) {
			printInfo("Filter skip: %s", path)
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
	trustData bool) {

	info, err := os.Stat(inputPath)
	if err != nil {
		printError("Error accessing input path: %v", err)
		return
	}

	exeFileName := filepath.Base(os.Args[0])

	maxCPU := runtime.NumCPU()
	if maxWorkers < 1 || maxWorkers > maxCPU {
		printWarn("maxWorkers value has been changed from %d to %d", maxWorkers, maxCPU)
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

	done := make(chan struct{})
	go func() {
		for err := range errorsCh {
			errList = append(errList, err)
		}
		close(done)
	}()

	effectiveCompress := func(name string) int {
		if matchesAnyPattern(name, ignoreCompressPatterns) {
			return 0
		}
		return compressType
	}

	var totalTasks int

	addTask := func(path string, name string) {
		totalTasks++
		rel := strings.TrimPrefix(path, inputPath)
		rel = strings.TrimPrefix(rel, string(filepath.Separator))
		outPath := filepath.Join(outputPath, rel)

		tasks <- task{
			path:           path,
			outPath:        outPath,
			processor:      processor,
			compressType:   effectiveCompress(name),
			keepOriginal:   keepOriginal,
			forcedCompress: forcedCompress,
			trustData:      trustData,
		}
	}

	finishAndReturn := func() {
		close(tasks)
		wg.Wait()
		close(errorsCh)
		<-done
		successCount := totalTasks - len(errList)
		printInfo("Operation completed! %d files, %d success, %d errors", totalTasks, successCount, len(errList))
		for _, e := range errList {
			printError("%v", e)
		}

		if len(errList) > 0 {
			waitForKey()
		}
	}

	if info.IsDir() {
		filepath.WalkDir(inputPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				printError("Error accessing path %s: %v", path, err)
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if shouldProcessFile(path, d.Name(), exeFileName, compressFlag, ignorePatterns, filterPatterns) {
				addTask(path, d.Name())
			}
			return nil
		})
	} else {
		if !shouldProcessFile(inputPath, info.Name(), exeFileName, compressFlag, ignorePatterns, filterPatterns) {
			finishAndReturn()
			return
		}
		addTask(inputPath, info.Name())
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
	trustData      bool
}

func worker(tasks <-chan task, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for tsk := range tasks {
		if err := tsk.processor(tsk.path, tsk.outPath, tsk.compressType, tsk.forcedCompress, tsk.trustData); err != nil {
			errors <- fmt.Errorf("processing file %s: %v", tsk.path, err)
		} else if !tsk.keepOriginal {
			if err := os.Remove(tsk.path); err != nil {
				errors <- fmt.Errorf("removing original file %s: %v", tsk.path, err)
			}
		}
	}
}

func waitForKey() {
	fmt.Print("\n\nPress Enter to continue...")

	var b [1]byte

	for {
		_, err := os.Stdin.Read(b[:])
		if err != nil {
			return
		}

		if b[0] == '\n' {
			break
		}
	}
}

func detectDirMode(root string) bool {
	unpack := false

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		if strings.HasSuffix(d.Name(), dvplExt) {
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

		if strings.HasSuffix(info.Name(), dvplExt) {
			processFiles(path, path, Unpack, false, false, 0, nil, nil, nil, maxWorkers, false, true)
		} else {
			processFiles(path, path, Pack, false, true, compressType, nil, nil, nil, maxWorkers, false, true)
		}
	}
}

func drawMenu(options []string, selected int) {
	for i, option := range options {
		if i == selected {
			fmt.Printf(esc_color_cyan+">  %s"+esc_color_reset+"\n", option)
		} else {
			fmt.Printf("  %s \n", option)
		}
	}
}

func interactiveMode(maxWorkers int) {
	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		printError("Failed to initialize keyboard: %v", err)
		return
	}
	defer keyboard.Close()

	fmt.Print(esc_cursorHide)
	defer fmt.Print(esc_cursorShow)
	fmt.Print(esc_clearAndHome)

	options := []string{
		"Compress",
		"Decompress",
		"Help",
	}
	selectedIndex := 0

	for {
		fmt.Printf(esc_cursorHome+"%s\n\nUsage: %sdvpl%s %s[-h]%s - To get help.\nPress Ctrl+C or Esc to exit.\n\n", dvplInf, esc_color_green, esc_color_reset, esc_color_cyan, esc_color_reset)

		drawMenu(options, selectedIndex)

		event := <-keysEvents
		if event.Err != nil {
			printError("Keyboard error: %v", event.Err)
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
			fmt.Print(esc_clearAndHome)
			switch selectedIndex {
			case 0:
				compressInteractive(keysEvents, maxWorkers)
			case 1:
				decompressInteractive(maxWorkers)
			case 2:
				keyboard.Close()
				flag.Usage()
				waitForKey()
			}
			return
		case keyboard.KeyEsc:
			return
		}
	}
}

func compressInteractive(keysEvents <-chan keyboard.KeyEvent, maxWorkers int) {
	fmt.Print(esc_cursorHide)
	defer fmt.Print(esc_cursorShow)
	fmt.Print(esc_clearAndHome)

	options := []string{
		"[0] none",
		"[1] lz4hc",
		"[2] lz4",
	}
	compressionTypes := []int{0, 1, 2}
	selectedIndex := 1

	for {
		fmt.Printf(esc_cursorHome+"%s\n\nSelect compression type.\nPress Ctrl+C or Esc to exit.\n\n", dvplInf)

		drawMenu(options, selectedIndex)

		event := <-keysEvents
		if event.Err != nil {
			printError("Keyboard error: %v", event.Err)
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
			fmt.Print(esc_clearAndHome)
			selectedCompressionType := compressionTypes[selectedIndex]
			processFiles(".", ".", Pack, false, true, selectedCompressionType, nil, nil, nil, maxWorkers, false, true)
			return
		case keyboard.KeyEsc:
			return
		}
	}
}

func decompressInteractive(maxWorkers int) {
	processFiles(".", ".", Unpack, false, false, 0, nil, nil, nil, maxWorkers, false, true)
}
