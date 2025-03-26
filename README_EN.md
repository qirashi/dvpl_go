# dvpl_go [RU](README.md) | [EN]
Dvpl converter to Go Lang. (Using the "C" code)

 > [!NOTE]
 > In this converter, the [lz4](https://github.com/lz4/lz4) library was used to speed up compression and improve its quality.

 > [!WARNING]
 > There will be no instructions on how to build using `dvpl_c`, if anyone is interested, figure it out on your own.

## Description

### Supported compression types

- `0` - `none` - There is no compression at all.
- `1` - `LZ4HC` - Stronger and slower than LZ4.
- `2` - `LZ4` - Less strong and faster than LZ4HC.
- `3` - `RFC1951` - Compression is not used in the game. (His reading is broken. I only added unpacking.)

### Console Output

```cmd
[debug] Config file found: .dvpl_go.yml
Usage: dvpl_go [options]
Options:
-c | -d
        Compress|Decompress .dvpl files
-i string
        Input path (file or directory)
-o string
        Output path (file or directory)
-keep-original
        Keep original files
-compress int
        Compression type 0 (none), 1 (lz4hc), 2 (lz4) | (default 1)
-ignore string
        Comma-separated list of file patterns to ignore

Examples:
  Compress   : dvpl_go -c -i ./input_dir -o ./output_dir
  Decompress : dvpl_go -d -i ./input_dir -o ./output_dir
  Ignore     : dvpl_go -c -i ./input_dir -ignore "*.exe,*.dll"
  Compression: dvpl_go -c -i ./input_dir -compress 2
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
- `-ignore` - A comma-separated list of file patterns to ignore.
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
        compress: 5
        keepOriginal: false
        inputPath: "./input_dir"
        outputPath: "./output_dir"
        compressFlag: false
        decompressFlag: false
        ignorePatterns:
        - "*.exe"
        - "*.dll"
        - "*.pdb"
        - "*.pak"
        - "temp*"

## Building the Project

### Rebuild `.mod` and `.sum`
1. Open a terminal in the project directory and execute the commands below.
2. `go mod init dvpl_go`
3. `go mod tidy`

---

### Build with Icon
1. Install `GO` on your PC.
2. Open `_build_.bat` in a text editor.
3. Edit the path to `ResourceHacker` or comment it out. (It is only needed for the icon.)
4. Run `_build_.bat`.
