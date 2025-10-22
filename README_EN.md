# dvpl_go [RU](README.md) | [EN]

 > [!NOTE]
 > In this converter, the [lz4](https://github.com/lz4/lz4) library was used to speed up compression and improve its quality.

## How to use the converter?
* [Гайд на Русском](.readme/how_to_use.md)
* [Guide in English](.readme/how_to_use_en.md)

## Description

### Supported compression types

- `0` - `none` - There is no compression at all.
- `1` - `LZ4HC` - Stronger and slower than LZ4.
- `2` - `LZ4` - Less strong and faster than LZ4HC.
- `3` - `RFC1951` - Compression is not used in the game. (The format is cut out.)

### Console Output

```cmd
R:\Github\dvpl_go>dvpl.exe -h
[debug] Config file found: .dvpl_go.yml
Usage: dvpl [options]
Options:
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

Examples:
  Compress   : dvpl -c -i ./input_dir -o ./output_dir
  Decompress : dvpl -d -i ./input_dir -o ./output_dir
  Ignore     : dvpl -c -i ./input_dir -ignore "*.exe,*.dll"
  Filter     : dvpl -d -i ./input_dir -o ./output_dir -filter "*.sc2,*.scg"
  No compress: dvpl -c -i ./input_dir -ignore-compress "*.webp"
  Compression: dvpl -c -i ./input_dir -compress 2
```

### Command Descriptions
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

- `-m` is the maximum number of parallel handlers (workers).
    - Default: 1 (single-threaded mode)
    - Optimal value: 2-4 (depends on CPU)
    - When values > maximum are specified, it is automatically adjusted.
    - The maximum number depends on the cores and threads of the processor.
    > **There may be problems with multithreaded mode running on energy-efficient cores from Intel.**

## Comparison of operation speed and compression

### This converter is on GoLang
```cmd
Start:     16:2:41.15
The end:   16:3:17.60
-----------------
Total: 0 h 0 min 36.45 sec

Weight: 1.15 GB (1,244,843,076 bytes)
```

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

### Results
- This converter is the best option at the moment. It allows you to select the compression level, which directly affects the speed of file compression and supports multithreaded mode. In the lz4hc compression mode, it surpasses its classmates in speed and is not inferior in compression quality. Another GoLang converter used `lz4`, which compresses worse, but faster (it was impossible to change the compression mode). The same converter is fast and supports all available methods.
