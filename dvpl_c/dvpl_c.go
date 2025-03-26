// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Qirashi
// Project: dvpl_go

package dvpl

/*
#cgo CFLAGS: -I${SRCDIR}/lz4_win64/include
#cgo LDFLAGS: -L${SRCDIR}/lz4_win64/static -lliblz4_static

#include <lz4.h>
#include <lz4hc.h>
#include <stdlib.h>
*/
import "C"

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

const (
	FooterSize = 20 // 4 * 4 + 4
	Marker     = "DVPL"
)

type Block struct {
	SeekPosInDvpl     uint64
	CompressedCRC32   uint32
	CompressedSize    uint32
	UncompressedCRC32 uint32
	UncompressedSize  uint32
}

type LZ4StreamHeader struct {
	NumOfParts uint32
	Parts      []Block
}

// compressLZ4 сжимает данные с помощью LZ4 (C-версия)
func compressLZ4(data []byte, level int) ([]byte, error) {
	bound := int(C.LZ4_compressBound(C.int(len(data))))
	if bound == 0 {
		return nil, fmt.Errorf("LZ4_compressBound failed")
	}

	compressed := make([]byte, bound)
	src := (*C.char)(unsafe.Pointer(&data[0]))
	dst := (*C.char)(unsafe.Pointer(&compressed[0]))

	var compressedSize int
	if level > 0 {
		// Используем LZ4 HC
		compressedSize = int(C.LZ4_compress_HC(
			src,
			dst,
			C.int(len(data)),
			C.int(bound),
			C.int(level),
		))
	} else {
		// Используем обычный LZ4
		compressedSize = int(C.LZ4_compress_default(
			src,
			dst,
			C.int(len(data)),
			C.int(bound),
		))
	}

	if compressedSize <= 0 {
		return nil, fmt.Errorf("LZ4 compression failed")
	}

	return compressed[:compressedSize], nil
}

// decompressLZ4 распаковывает данные с помощью LZ4 (C-версия)
func decompressLZ4(compressed []byte, uncompressedSize int) ([]byte, error) {
	uncompressed := make([]byte, uncompressedSize)
	src := (*C.char)(unsafe.Pointer(&compressed[0]))
	dst := (*C.char)(unsafe.Pointer(&uncompressed[0]))

	decompressedSize := int(C.LZ4_decompress_safe(
		src,
		dst,
		C.int(len(compressed)),
		C.int(uncompressedSize),
	))

	if decompressedSize < 0 {
		return nil, fmt.Errorf("LZ4 decompression failed")
	}

	return uncompressed[:decompressedSize], nil
}

// Pack сжимает файл и добавляет DVPL-футер
func Pack(inputPath, outputPath string, compressType int) error {
	fmt.Printf("Pack: %s\n", inputPath)

	// Читаем данные из файла
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("[error] Failed to read input file: %v", err)
	}

	if len(data) == 0 {
		compressType = 0
	}

	var result []byte
	var ptype uint32

	switch compressType {
	case 0: // Без сжатия
		result = data
		ptype = 0
	case 1: // LZ4 HC
		compressed, err := compressLZ4(data, 9) // Уровень 9 для HC
		if err != nil {
			return fmt.Errorf("[error] Failed to compress data: %v", err)
		}

		// Если сжатые данные больше или равны исходным, не сжимаем
		if len(compressed) == 0 || len(compressed) >= len(data) {
			result = data
			ptype = 0 // Без сжатия
		} else {
			result = compressed
			ptype = 1 // LZ4 HC
		}
	case 2: // LZ4
		compressed, err := compressLZ4(data, 0) // Уровень 0 для обычного LZ4
		if err != nil {
			return fmt.Errorf("[error] Failed to compress data: %v", err)
		}

		// Если сжатые данные больше или равны исходным, не сжимаем
		if len(compressed) == 0 || len(compressed) >= len(data) {
			result = data
			ptype = 0 // Без сжатия
		} else {
			result = compressed
			ptype = 2 // LZ4
		}
	default:
		return fmt.Errorf("[error] Unsupported compression type: %d", compressType)
	}

	// Вычисляем CRC32 для результата
	crc := crc32.ChecksumIEEE(result)

	// Создаем футер
	unpacked := uint32(len(data))
	packed := uint32(len(result))

	footer := struct {
		Unpacked uint32
		Packed   uint32
		CRC      uint32
		PType    uint32
		Marker   [4]byte
	}{
		Unpacked: unpacked,
		Packed:   packed,
		CRC:      crc,
		PType:    ptype,
		Marker:   [4]byte{'D', 'V', 'P', 'L'},
	}

	// Записываем результат и футер в выходной файл
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("[error] Failed to create output file: %v", err)
	}
	defer outFile.Close()

	// Записываем данные
	if _, err := outFile.Write(result); err != nil {
		return fmt.Errorf("[error] Failed to write data: %v", err)
	}

	// Записываем футер
	if err := binary.Write(outFile, binary.LittleEndian, footer); err != nil {
		return fmt.Errorf("[error] Failed to write footer: %v", err)
	}

	return nil
}

// Unpack распаковывает DVPL-файл
func Unpack(inputPath, outputPath string, _ int) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("[error] Failed to open input file: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("[error] Failed to get file info: %v", err)
	}

	fileSize := fileInfo.Size()
	if fileSize < FooterSize {
		return fmt.Errorf("[error] File too small to contain footer")
	}

	// Читаем футер
	_, err = file.Seek(fileSize-FooterSize, io.SeekStart)
	if err != nil {
		return fmt.Errorf("[error] Failed to seek to footer: %v", err)
	}

	footerData := make([]byte, FooterSize)
	if _, err := io.ReadFull(file, footerData); err != nil {
		return fmt.Errorf("[error] Failed to read footer: %v", err)
	}

	footer := struct {
		Unpacked uint32
		Packed   uint32
		CRC      uint32
		PType    uint32
		Marker   [4]byte
	}{}

	if err := binary.Read(bytes.NewReader(footerData), binary.LittleEndian, &footer); err != nil {
		return fmt.Errorf("[error] Failed to parse footer: %v", err)
	}

	if string(footer.Marker[:]) != Marker {
		return fmt.Errorf("[error] Invalid marker: %s", footer.Marker)
	}

	// Выводим тип сжатия
	var compressionType string
	switch footer.PType {
	case 0:
		compressionType = "none"
	case 1:
		compressionType = "lz4hc"
	case 2:
		compressionType = "lz4"
	case 3:
		compressionType = "rfc1951"
	default:
		compressionType = "unknown"
	}
	fmt.Printf("Unpack %s: %s\n", compressionType, inputPath)

	// Читаем сжатые данные
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("[error] Failed to seek to start: %v", err)
	}

	compressedData := make([]byte, footer.Packed)
	if _, err := io.ReadFull(file, compressedData); err != nil {
		return fmt.Errorf("[error] Failed to read compressed data: %v", err)
	}

	// Проверяем CRC32
	if crc32.ChecksumIEEE(compressedData) != footer.CRC {
		return fmt.Errorf("[error] CRC32 mismatch")
	}

	// Распаковываем данные
	var result []byte
	switch footer.PType {
	case 0: // Без сжатия
		result = compressedData
	case 1, 2: // lz4
		result, err = decompressLZ4(compressedData, int(footer.Unpacked))
		if err != nil {
			return fmt.Errorf("[error] Failed to decompress LZ4 data: %v", err)
		}
	case 3: // rfc1951
		reader, err := zlib.NewReader(bytes.NewReader(compressedData))
		if err != nil {
			return fmt.Errorf("[error] Failed to create zlib reader: %v", err)
		}
		defer reader.Close()
		result, err = io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("[error] Failed to decompress zlib data: %v", err)
		}
	default:
		return fmt.Errorf("[error] Unsupported compression type: %d", footer.PType)
	}

	// Убираем расширение .dvpl из имени файла
	if strings.HasSuffix(outputPath, ".dvpl") {
		outputPath = strings.TrimSuffix(outputPath, ".dvpl")
	}

	// Создаем директорию, если она не существует
	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return fmt.Errorf("[error] Failed to create output directory: %v", err)
	}

	// Записываем распакованные данные в выходной файл
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("[error] Failed to create output file: %v", err)
	}
	defer outFile.Close()

	if _, err := outFile.Write(result); err != nil {
		return fmt.Errorf("[error] Failed to write decompressed data: %v", err)
	}

	return nil
}