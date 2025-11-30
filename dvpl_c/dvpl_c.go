// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Qirashi
// Project: dvpl_go

package dvpl_c

/*
#cgo CFLAGS: -I${SRCDIR}/lz4_win64/include -O3 -DNDEBUG
#cgo LDFLAGS: -L${SRCDIR}/lz4_win64/static -lliblz4_static -Wl,-O1 -Wl,--as-needed

#include <lz4.h>
#include <lz4hc.h>
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/klauspost/crc32" // Замена "hash/crc32"
	// "compress/zlib"
	// "io"
)

const (
	FooterSize = 20 // 4 * 4 + 4
	Marker     = "DVPL"
)

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
		compressedSize = int(C.LZ4_compress_HC(src, dst, C.int(len(data)), C.int(bound), C.int(level)))
	} else {
		// Используем обычный LZ4
		compressedSize = int(C.LZ4_compress_default(src, dst, C.int(len(data)), C.int(bound)))
	}

	if compressedSize <= 0 {
		return nil, fmt.Errorf("LZ4 compression failed")
	}

	return compressed[:compressedSize], nil
}

// CGO
func decompressLZ4(compressed []byte, uncompressedSize int) ([]byte, error) {
	uncompressed := make([]byte, uncompressedSize)
	src := (*C.char)(unsafe.Pointer(&compressed[0]))
	dst := (*C.char)(unsafe.Pointer(&uncompressed[0]))

	decompressedSize := int(C.LZ4_decompress_safe(src, dst, C.int(len(compressed)), C.int(uncompressedSize)))

	if decompressedSize < 0 {
		return nil, fmt.Errorf("LZ4 decompression failed")
	}

	return uncompressed[:decompressedSize], nil
}
func Pack(data []byte, compressType int, forcedCompress bool) ([]byte, uint32, error) {
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
			return result, ptype, fmt.Errorf("[error] Failed to compress data: %v", err)
		}

		if !forcedCompress && len(compressed) >= len(data) {
			result = data
			ptype = 0 // Без сжатия
		} else {
			result = compressed
			ptype = 1 // LZ4 HC
		}
	case 2: // LZ4
		compressed, err := compressLZ4(data, 0) // Уровень 0 для обычного LZ4
		if err != nil {
			return result, ptype, fmt.Errorf("[error] Failed to compress data: %v", err)
		}

		if !forcedCompress && len(compressed) >= len(data) {
			result = data
			ptype = 0 // Без сжатия
		} else {
			result = compressed
			ptype = 2 // LZ4
		}
	// case 3: // rfc1951 (zlib)
		// var compressed bytes.Buffer
		// writer := zlib.NewWriter(&compressed)
		// if _, err := writer.Write(data); err != nil {
			// return result, ptype, fmt.Errorf("[error] Failed to compress data with zlib: %v", err)
		// }
		// if err := writer.Close(); err != nil {
			// return result, ptype, fmt.Errorf("[error] Failed to close zlib writer: %v", err)
		// }

		// if !forcedCompress && compressed.Len() >= len(data) {
			// result = data
			// ptype = 0 // Без сжатия
		// } else {
			// result = compressed.Bytes()
			// ptype = 3 // rfc1951
		// }
	default:
		return result, ptype, fmt.Errorf("[error] Unsupported compression type: %d", compressType)
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

	// Добавляем футер к результату
	var footerBuf bytes.Buffer
	if err := binary.Write(&footerBuf, binary.LittleEndian, footer); err != nil {
		return result, ptype, fmt.Errorf("[error] Failed to write footer: %v", err)
	}
	result = append(result, footerBuf.Bytes()...)

	return result, ptype, nil
}

func Unpack(data []byte) ([]byte, uint32, error) {
	if len(data) < FooterSize {
		return nil, 0, fmt.Errorf("invalid DVPL data: size %d is less than footer size %d", len(data), FooterSize)
	}

	// Читаем футер (последние FooterSize байт)
	footerData := data[len(data)-FooterSize:]
	data = data[:len(data)-FooterSize] // Отрезаем футер от сжатых данных

	footer := struct {
		Unpacked uint32
		Packed   uint32
		CRC      uint32
		PType    uint32
		Marker   [4]byte
	}{}

	if err := binary.Read(bytes.NewReader(footerData), binary.LittleEndian, &footer); err != nil {
		return nil, footer.PType, fmt.Errorf("[error] Failed to parse footer: %v", err)
	}

	if string(footer.Marker[:]) != Marker {
		return nil, footer.PType, fmt.Errorf("[error] Invalid marker: %s", footer.Marker)
	}

	// Проверяем CRC32
	if crc32.ChecksumIEEE(data) != footer.CRC {
		return nil, footer.PType, fmt.Errorf("[error] CRC32 mismatch")
	}

	var result []byte
	var err error // Объявляем error здесь, чтобы использовать во всех case

	switch footer.PType {
	case 0: // Без сжатия
		result = data
	case 1, 2: // lz4
		result, err = decompressLZ4(data, int(footer.Unpacked)) // Используем = вместо :=
		if err != nil {
			return nil, footer.PType, fmt.Errorf("[error] Failed to decompress LZ4 data: %v", err)
		}
	// case 3: // rfc1951
		// var reader io.ReadCloser
		// reader, err = zlib.NewReader(bytes.NewReader(data))
		// if err != nil {
			// return nil, footer.PType, fmt.Errorf("[error] Failed to create zlib reader: %v", err)
		// }
		// defer reader.Close()
		// result, err = io.ReadAll(reader)
		// if err != nil {
			// return nil, footer.PType, fmt.Errorf("[error] Failed to decompress zlib data: %v", err)
		// }
	default:
		return nil, footer.PType, fmt.Errorf("[error] Unsupported compression type: %d", footer.PType)
	}

	return result, footer.PType, nil
}