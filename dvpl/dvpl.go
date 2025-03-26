// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Qirashi
// Project: dvpl_go

package dvpl

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
	"github.com/pierrec/lz4/v4"
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

func Pack(inputPath, outputPath string, compressType int) error {
	fmt.Printf("Pack: %s\n", inputPath)

	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("[error] Failed to read input file: %v", err)
	}

	var result []byte
	var ptype uint32

	switch compressType {
	case 0:
		result = data
		ptype = 0
	case 1:
		hashTable := make([]int, 1<<18)
		chainTable := make([]int, 1<<18)

		compressed := make([]byte, lz4.CompressBlockBound(len(data)))
		n, err := lz4.CompressBlockHC(data, compressed, lz4.Level5, hashTable, chainTable)
		if err != nil {
			return fmt.Errorf("[error] Failed to compress data: %v", err)
		}
		compressed = compressed[:n]

		if n == 0 || n >= len(data) {
			result = data
			ptype = 0
		} else {
			result = compressed
			ptype = 1
		}
	case 2:
		compressed := make([]byte, lz4.CompressBlockBound(len(data)))
		n, err := lz4.CompressBlock(data, compressed, nil)
		if err != nil {
			return fmt.Errorf("[error] Failed to compress data: %v", err)
		}
		compressed = compressed[:n]

		if n == 0 || n >= len(data) {
			result = data
			ptype = 0
		} else {
			result = compressed
			ptype = 2
		}
	default:
		return fmt.Errorf("[error] Unsupported compression type: %d", compressType)
	}

	crc := crc32.ChecksumIEEE(result)

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

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("[error] Failed to create output file: %v", err)
	}
	defer outFile.Close()

	if _, err := outFile.Write(result); err != nil {
		return fmt.Errorf("[error] Failed to write data: %v", err)
	}

	if err := binary.Write(outFile, binary.LittleEndian, footer); err != nil {
		return fmt.Errorf("[error] Failed to write footer: %v", err)
	}

	return nil
}


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

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("[error] Failed to seek to start: %v", err)
	}

	compressedData := make([]byte, footer.Packed)
	if _, err := io.ReadFull(file, compressedData); err != nil {
		return fmt.Errorf("[error] Failed to read compressed data: %v", err)
	}

	if crc32.ChecksumIEEE(compressedData) != footer.CRC {
		return fmt.Errorf("[error] CRC32 mismatch")
	}

	var result []byte
	switch footer.PType {
	case 0:
		result = compressedData
	case 1, 2:
		result = make([]byte, footer.Unpacked)
		_, err := lz4.UncompressBlock(compressedData, result)
		if err != nil {
			return fmt.Errorf("[error] Failed to decompress LZ4 data: %v", err)
		}
	case 3:
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

	if strings.HasSuffix(outputPath, ".dvpl") {
		outputPath = strings.TrimSuffix(outputPath, ".dvpl")
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return fmt.Errorf("[error] Failed to create output directory: %v", err)
	}

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
