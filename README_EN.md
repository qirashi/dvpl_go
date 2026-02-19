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
  > | Type |   Title  |            Description                   |
  > |------|----------|------------------------------------------|
  > |  0   |   none   | There is no compression at all.          |
  > |  1   |   lz4hc  | Stronger and slower than lz4.            |
  > |  2   |   lz4    | Less strong and faster than lz4hc.       |

## Environment Variables
  Environment variables can store two converter settings, `DVPL_MAX_WORKERS` and `DVPL_COMPRESS_TYPE`, to specify the number of parallel running processes and the compression type, respectively.

- `DVPL_MAX_WORKERS` — Maximum number of parallel workers. (If the number is too large, it will be limited.)
- `DVPL_COMPRESS_TYPE` — Specifies the compression level from 0 to 2. (If the type does not exist, an error will occur.)

How to set variables:
1. **Create manually**
    - Press `Win+R`.
    - Run `SystemPropertiesAdvanced`.
    - Open `Environment Variables` and create the appropriate variables.

2. **Via the command line**
    - `Win` → `cmd` → `right-click` → `Run as administrator`.
    - Insert one of the commands:
      *   For a single user:
            ```cmd
            setx DVPL_MAX_WORKERS 4
            setx DVPL_COMPRESS_TYPE 1
            ```

      *   For all users (with admin):
            ```cmd
            setx DVPL_MAX_WORKERS 4 /M
            setx DVPL_COMPRESS_TYPE 1 /M
            ```

## CMD

```cmd
R:\Github\dvpl_go\out>dvpl.exe -h

dvpl_go 2.0.0 x64 | Copyright (c) 2026 Qirashi

Usage: dvpl [options]
[Options]:
  -c    Compress .dvpl files.
  -compress int
        Compression type: 0 (none), 1 (lz4hc), 2 (lz4) | (default 1)
  -d    Decompress .dvpl files.
  -filter string
        List of file patterns to include. ("*.sc2,*.scg")
  -forced-compress
        Force compression even if the result is larger than the original.
  -i string
        Input path. (file or directory)
  -ignore string
        List of file patterns to ignore. ("*.exe,*.dll")
  -ignore-compress string
        List of file patterns for which compression should be disabled. ("*.webp")
  -keep-original
        Keep original files.
  -m int
        Maximum number of parallel workers (12). Minimum 1, recommended 2 | (default 2)
  -o string
        Output path. (file or directory)
  -skip-crc
        CRC can be ignored when unpacking or packing.

Examples:
  Compress   : dvpl -c -i ./in_dir -o ./out_dir
  Decompress : dvpl -d -i ./in_dir -o ./out_dir
  Ignore     : dvpl -c -i ./in_dir -ignore "*.exe,*.dll"
  Filter     : dvpl -d -i ./in_dir -o ./out_dir -filter "*.sc2,*.scg"
  No compress: dvpl -c -i ./in_dir -ignore-compress "*.webp"
  Compression: dvpl -c -i ./in_dir -compress 2
```

### Command Descriptions
- `-c` — Compress to `.dvpl`.
- `-d` — Decompress `.dvpl`.
- `-i` — Input directory or file.
- `-o` — Output directory or file.
- `-keep-original` — Keep the original file during decompression or compression.
- `-compress` — Specifies the compression level from 0 to 2.
    - `0` — `none`
    - `1` — `lz4hc`
    - `2` — `lz4`
- `-ignore` — List of file patterns to ignore. (Files and extensions will not be processed)
- `-ignore-compress` — A list of file templates that will be forcibly compressed to type 0. (For example, `*.webp`)
- `-filter` — A list of template files to be processed. (Only files and extensions that will be processed are the reverse of `-ignore`)
    - For example, you need to unpack only `*.webp` and `*.txt` into a separate folder.
    - It will look like this: `dvpl -d -i ./in -o ./out -filter "*.webp,*.txt" -keep-original -m 4`
    #### Wildcards for filters:
    - `*` — Any number of characters (except `/`).
    - `?` — One character.
    - `[abc]` — One of the specified characters.

    #### Examples:
    - `*.exe` — Ignore all `.exe` files.
    - `file?.log` — Ignore files like `file1.log`, `file2.log`.
    - `folder/*.txt` — Ignore all `.txt` files in the `folder` directory.
    - `data[1-3].csv` — Ignore files `data1.csv`, `data2.csv`, `data3.csv`.
    - `image_[xyz].png` — Ignore files `image_x.png`, `image_y.png`, `image_z.png`.

- `-m` is the maximum number of parallel handlers (workers).
    - Default: 2 (single-threaded mode)
    - Optimal value: 2-4 (depends on CPU)
    - When values > maximum are specified, it is automatically adjusted.
    - The maximum number depends on the cores and threads of the processor.

- `-skip-crc` - When unpacking, the CRC will be ignored.

## Comparison of operating speed

### This converter is for GoLang with multithreading (2 workers)
```
Start:     16:4:43.85
The end:   16:5:2.78
-----------------
Total: 0 h 0 min 18.93 sec

Weight: 1.15 GB (1,244,843,076 bytes)
```

### Another NodeJS converter
```
Start:     15:59:13.41
The end:   16:0:10.19
-----------------
Total: 0 h 0 min 56.78 sec

Weight: 1.15 GB (1,243,007,962 bytes)
```

### Another converter for GoLang
```
Start:     16:18:37.28
The end:   16:18:43.51
-----------------
Total: 0 h 0 min 6.23 sec

Weight: 2.81 GB (3,020,488,406 bytes)
```

## Results
  This converter offers optimal compression and speed. It allows you to choose the compression level, which affects speed. In the `lz4hc` compression mode, it outperforms its classmates in speed and is equally high-quality. Another converter in Go used `lz4`, which offers worse compression but is faster. This converter is fast and supports all the main available methods.
