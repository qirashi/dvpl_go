// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2026 Qirashi
// Project: dvpl_go

package dvpl_c

/*
#cgo windows CFLAGS: -I${SRCDIR}/lz4_win64_dev/include -O3 -DNDEBUG -fomit-frame-pointer -mtune=native
#cgo windows LDFLAGS: -L${SRCDIR}/lz4_win64_dev/static -lliblz4 -Wl,-O1 -Wl,--as-needed

#cgo linux CFLAGS: -I${SRCDIR}/lz4_linux_dev/include -O3 -DNDEBUG -fomit-frame-pointer
#cgo linux LDFLAGS: -L${SRCDIR}/lz4_linux_dev/static -llz4 -Wl,-O1 -Wl,--as-needed

#include <lz4.h>
#include <lz4hc.h>
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/klauspost/crc32"
)

var dvplMarker = [4]byte{'D', 'V', 'P', 'L'}

func compressLZ4(data []byte, level int) ([]byte, error) {
	lenData := len(data)

	bound := int(C.LZ4_compressBound(C.int(lenData)))
	if bound == 0 {
		return nil, fmt.Errorf("LZ4_compressBound failed")
	}

	compressed := make([]byte, bound)
	src := (*C.char)(unsafe.Pointer(&data[0]))
	dst := (*C.char)(unsafe.Pointer(&compressed[0]))

	var compressedSize int
	if level > 0 {
		compressedSize = int(C.LZ4_compress_HC(src, dst, C.int(lenData), C.int(bound), C.int(level))) // LZ4 HC
	} else {
		compressedSize = int(C.LZ4_compress_default(src, dst, C.int(lenData), C.int(bound))) // LZ4
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

func Pack(data []byte, compressType int, forcedCompress bool) ([]byte, uint32, error) {
	lenData := len(data)

	if lenData == 0 {
		compressType = 0
	}

	var result []byte
	var ptype uint32

	switch compressType {
	case 0: // NONE
		result = data
		ptype = 0
	case 1: // LZ4 HC
		compressed, err := compressLZ4(data, 9)
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
		compressed, err := compressLZ4(data, 0)
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
		Marker:   dvplMarker,
	}

	const footerSize = 20 // Unpacked(4) + Packed(4) + CRC(4) + PType(4) + Marker(4)
	footerBytes := make([]byte, footerSize)
	binary.LittleEndian.PutUint32(footerBytes[0:4], footer.Unpacked)
	binary.LittleEndian.PutUint32(footerBytes[4:8], footer.Packed)
	binary.LittleEndian.PutUint32(footerBytes[8:12], footer.CRC)
	binary.LittleEndian.PutUint32(footerBytes[12:16], footer.PType)
	copy(footerBytes[16:20], footer.Marker[:])
	result = append(result, footerBytes...)

	return result, ptype, nil
}

func Unpack(data []byte, trustData bool) ([]byte, uint32, error) {
	lenData := len(data)

	const footerSize = 20 // Unpacked(4) + Packed(4) + CRC(4) + PType(4) + Marker(4)
	if lenData < footerSize {
		return nil, 0, fmt.Errorf("invalid DVPL data: size %d is less than footer size %d", lenData, footerSize)
	}

	payloadSize := lenData - footerSize
	footerBytes := data[payloadSize:]
	data = data[:payloadSize]

	unpacked := binary.LittleEndian.Uint32(footerBytes[0:4])
	packed := binary.LittleEndian.Uint32(footerBytes[4:8])
	crc := binary.LittleEndian.Uint32(footerBytes[8:12])
	ptype := binary.LittleEndian.Uint32(footerBytes[12:16])

	if footerBytes[16] != dvplMarker[0] || footerBytes[17] != dvplMarker[1] || footerBytes[18] != dvplMarker[2] || footerBytes[19] != dvplMarker[3] {
		return nil, ptype, fmt.Errorf("invalid marker")
	}

	if lenData == footerSize {
		return []byte{}, 0, nil
	}

	if uint32(payloadSize) != packed {
		return nil, ptype, fmt.Errorf("packed size mismatch: got %d, expected %d", lenData, packed)
	}

	if !trustData && unpacked > (1<<30) { // > 1 GB
		return nil, ptype, fmt.Errorf("unpacked size too large: %d. Use -trust-data for unpacking.", unpacked)
	}

	if !trustData && crc32.ChecksumIEEE(data) != crc {
		return nil, ptype, fmt.Errorf("CRC32 mismatch")
	}

	var result []byte
	var err error

	switch ptype {
	case 0: // NONE
		result = data
	case 1, 2: // LZ4
		result, err = decompressLZ4(data, int(unpacked))
		if err != nil {
			return nil, ptype, fmt.Errorf("failed to decompress LZ4 data: %v", err)
		}
	default:
		return nil, ptype, fmt.Errorf("unsupported compression type: %d", ptype)
	}

	return result, ptype, nil
}
