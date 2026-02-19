<div align="center">
	
[![GitHub license](https://img.shields.io/github/license/qirashi/dvpl_go?logo=apache&label=License&style=flat  )](https://github.com/qirashi/dvpl_go/blob/main/LICENSE  )
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/qirashi/dvpl_go/total?logo=github&label=Downloads&style=flat  )
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/qirashi/dvpl_go?style=flat&label=Code%20Size  )

</div>

# dvpl_go [RU](README.md) | [EN]

  > [!NOTE]
  > The converter uses the [lz4](https://github.com/lz4/lz4) library to improve compression speed and quality.  
  > The format has limitations on the size of compressed data!  

## How to use?
  > [!TIP]  
  > [Гайд по использованию конвертера на Русском](.readme/how_to_use.md)  
  > [A guide to using the converter in English](.readme/how_to_use_en.md)  

## Supported compression types

  > [!NOTE]  
  > `0` - `none` - There is no compression at all.  
  > `1` - `lz4hc` - Stronger and slower than lz4.  
  > `2` - `lz4` - Less strong and faster than lz4hc.  

## CMD

```cmd
R:\Github\dvpl_go>dvpl.exe -h
[debug] Configuration loaded: .dvpl_go.yml

dvpl_go 1.4.1 x64 | Copyright (c) 2026 Qirashi

Usage: dvpl [options]
[Options]:
  -c    Compress .dvpl files
  -compress int
        Compression type: 0 (none), 1 (lz4hc), 2 (lz4) | (default 1)
  -d    Decompress .dvpl files
  -filter string
        Comma-separated list of file patterns to include (e.g. "*.sc2,*.scg")
  -forced-compress
        Forced compression, even if the result is larger than the original
  -i string
        Input path (file or directory)
  -ignore string
        Comma-separated list of file patterns to ignore
  -ignore-compress string
        Comma-separated list of file patterns to force no compression (type 0)
  -keep-original
        Keep original files
  -m int
        Maximum number of parallel workers (12). Minimum 1, recommended 2. (default 1)
  -o string
        Output path (file or directory)
  -skip-crc
        When unpacking, the crc will be ignored.

Examples:
  Compress   : dvpl -c -i ./input_dir -o ./output_dir
  Decompress : dvpl -d -i ./input_dir -o ./output_dir
  Ignore     : dvpl -c -i ./input_dir -ignore "*.exe,*.dll"
  Filter     : dvpl -d -i ./input_dir -o ./output_dir -filter "*.sc2,*.scg"
  No compress: dvpl -c -i ./input_dir -ignore-compress "*.webp"
  Compression: dvpl -c -i ./input_dir -compress 2
```

## Command Descriptions
- `-c` - Compress files into `.dvpl`.
- `-d` - Decompress `.dvpl` files.
- `-i` - Input directory or file.
- `-o` - Output directory or file.
- `-keep-original` - Keep the original file during decompression or compression.
- `-compress` - Specifies the compression level from 0 to 2.
    - `0` - `none`
    - `1` - `lz4hc`
    - `2` - `lz4`
- `-ignore` is a list of file templates that should be ignored. (Files and extensions will not be processed)
- `-ignore-compress` - A list of file templates that will be forcibly compressed to type 0. (For example, `*.webp')
- `-filter` - A list of template files to be processed. (Only files and extensions that will be processed are the reverse of `-ignore`)
    - For example, you need to unpack only `*.webp` and `*.txt` into a separate folder.
    - It will look like this: `dvpl -d -i ./in -o ./out -filter "*.webp,*.txt" -keep-original -m 4`
    #### Supported wildcard characters:
    - `*` — Any number of characters (except `/`).
    - `?` — One character.
    - `[abc]` — One of the specified characters.

    #### Usage Examples:
    - `*.exe` — Ignore all `.exe` files.
    - `file?.log` — Ignore files like `file1.log`, `file2.log`.
    - `folder/*.txt` — Ignore all `.txt` files in the `folder` directory.
    - `data[1-3].csv` — Ignore files `data1.csv`, `data2.csv`, `data3.csv`.
    - `image_[xyz].png` — Ignore files `image_x.png`, `image_y.png`, `image_z.png`.

    #### The contents of the .dvpl_go.yml:
```yaml
compress: 1
keepOriginal: false
inputPath: "./input_dir"
outputPath: "./output_dir"
compressFlag: false
decompressFlag: false
forcedCompress: false
maxWorkers: 2
ignorePatterns:
- "*.exe"
- "*.dll"
- "*.pdb"
- "*.pak"
- "temp*"
filterPatterns:
- "*.sc2"
- "*.scg"
ignoreCompress:
- "*.webp"
skipCRC: false
```

- `-m` is the maximum number of parallel handlers (workers).
    - Default: 1 (single-threaded mode)
    - Optimal value: 2-4 (depends on CPU)
    - When values > maximum are specified, it is automatically adjusted.
    - The maximum number depends on the cores and threads of the processor.
> [!WARNING]  
> **There may be problems with multithreaded mode running on energy-efficient cores from Intel.**

- `-skip-crc` - When unpacking, the CRC will be ignored.

## Comparison of operation speed and compression

### This converter is for GoLang with multithreading (2 workers)
```cmd
Start:     16:4:43.85
The end:   16:5:2.78
-----------------
Total: 0 h 0 min 18.93 sec

Weight: 1.15 GB (1,244,843,076 bytes)
```

### Another NodeJS converter
```cmd
Start:     15:59:13.41
The end:   16:0:10.19
-----------------
Total: 0 h 0 min 56.78 sec

Weight: 1.15 GB (1,243,007,962 bytes)
```

### Another converter for GoLang
```cmd
Start:     16:18:37.28
The end:   16:18:43.51
-----------------
Total: 0 h 0 min 6.23 sec

Weight: 2.81 GB (3,020,488,406 bytes)
```

## Results
This converter offers the best compression and speed. It allows you to choose the compression level, which directly impacts file compression speed, and supports multithreaded mode. In lz4hc compression mode, it outperforms its peers in speed and maintains the same compression quality. Another GoLang converter used lz4, which compresses significantly worse, but is faster (it was impossible to change the compression mode). This converter works quickly and supports all the main available methods.
