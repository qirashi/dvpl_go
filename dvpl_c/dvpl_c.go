// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Qirashi
// Project: dvpl_go

package dvpl_c

/*
#cgo CFLAGS: -I${SRCDIR}/lz4_win64_dev/include -O3 -DNDEBUG -fomit-frame-pointer -mtune=native
#cgo LDFLAGS: -L${SRCDIR}/lz4_win64_dev/static -lliblz4_static -Wl,-O1 -Wl,--as-needed

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

var Marker = [4]byte{'D', 'V', 'P', 'L'}

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
		compressedSize = int(C.LZ4_compress_HC(src, dst, C.int(len(data)), C.int(bound), C.int(level))) // LZ4 HC
	} else {
		compressedSize = int(C.LZ4_compress_default(src, dst, C.int(len(data)), C.int(bound))) // LZ4
	}

	if compressedSize <= 0 {
		return nil, fmt.Errorf("LZ4 compression failed")
	}

	return compressed[:compressedSize], nil
}

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

func Pack(data []byte, compressType int, forcedCompress bool, skipCRC bool) ([]byte, uint32, error) {
	lenData := len(data)

	if lenData == 0 {
		compressType = 0
	}

	var result []byte
	var ptype uint32

	switch compressType {
	case 0: // none
		result = data
		ptype = 0
	case 1: // LZ4 HC
		compressed, err := compressLZ4(data, 9) // Уровень 9 для HC
		if err != nil {
			return result, ptype, fmt.Errorf("failed to compress data: %v", err)
		}

		if !forcedCompress && len(compressed) >= lenData {
			result = data
			ptype = 0
		} else {
			result = compressed
			ptype = 1
		}
	case 2: // LZ4
		compressed, err := compressLZ4(data, 0) // Уровень 0 для обычного LZ4
		if err != nil {
			return result, ptype, fmt.Errorf("failed to compress data: %v", err)
		}

		if !forcedCompress && len(compressed) >= lenData {
			result = data
			ptype = 0
		} else {
			result = compressed
			ptype = 2
		}
	// case 3: // rfc1951 (zlib)
	// var compressed bytes.Buffer
	// writer := zlib.NewWriter(&compressed)
	// if _, err := writer.Write(data); err != nil {
	// return result, ptype, fmt.Errorf("failed to compress data with zlib: %v", err)
	// }
	// if err := writer.Close(); err != nil {
	// return result, ptype, fmt.Errorf("failed to close zlib writer: %v", err)
	// }

	// if !forcedCompress && compressed.Len() >= lenData {
	// result = data
	// ptype = 0 // Без сжатия
	// } else {
	// result = compressed.Bytes()
	// ptype = 3 // rfc1951
	// }
	default:
		return result, ptype, fmt.Errorf("unsupported compression type: %d", compressType)
	}

	packed := uint32(len(result))
	crc := crc32.ChecksumIEEE(result)

	footer := struct {
		Unpacked uint32
		Packed   uint32
		CRC      uint32
		PType    uint32
		Marker   [4]byte
	}{
		Unpacked: uint32(lenData),
		Packed:   packed,
		CRC:      crc,
		PType:    ptype,
		Marker:   Marker,
	}

	// Добавляем футер к результату
	var footerBuf bytes.Buffer
	if err := binary.Write(&footerBuf, binary.LittleEndian, footer); err != nil {
		return result, ptype, fmt.Errorf("failed to write footer: %v", err)
	}
	result = append(result, footerBuf.Bytes()...)

	return result, ptype, nil
}

func Unpack(data []byte, skipCRC bool) ([]byte, uint32, error) {
	lenData := len(data)
	const footerSize = 20 // Unpacked(4) + Packed(4) + CRC(4) + PType(4) + Marker(4)
	if lenData < footerSize {
		return nil, 0, fmt.Errorf("invalid DVPL data: size %d is less than footer size %d", lenData, footerSize)
	}

	footerBytes := data[lenData-footerSize:]
	data = data[:lenData-footerSize]

	unpacked := binary.LittleEndian.Uint32(footerBytes[0:4])
	crc := binary.LittleEndian.Uint32(footerBytes[8:12])
	ptype := binary.LittleEndian.Uint32(footerBytes[12:16])
	var marker [4]byte
	copy(marker[:], footerBytes[16:20])

	if marker != Marker {
		return nil, ptype, fmt.Errorf("invalid marker: %v", marker)
	}

	if !skipCRC && crc32.ChecksumIEEE(data) != crc {
		return nil, ptype, fmt.Errorf("CRC32 mismatch")
	}

	var result []byte
	var err error

	switch ptype {
	case 0: // none
		result = data
	case 1, 2: // LZ4
		result, err = decompressLZ4(data, int(unpacked))
		if err != nil {
			return nil, ptype, fmt.Errorf("failed to decompress LZ4 data: %v", err)
		}
	// case 3: // rfc1951
	// var reader io.ReadCloser
	// reader, err = zlib.NewReader(bytes.NewReader(data))
	// if err != nil {
	// return nil, footer.PType, fmt.Errorf("failed to create zlib reader: %v", err)
	// }
	// defer reader.Close()
	// result, err = io.ReadAll(reader)
	// if err != nil {
	// return nil, footer.PType, fmt.Errorf("[failed to decompress zlib data: %v", err)
	// }

	default:
		return nil, ptype, fmt.Errorf("unsupported compression type: %d", ptype)
	}

	return result, ptype, nil
}
