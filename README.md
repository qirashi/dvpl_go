# dvpl_go [RU] | [EN](README_EN.md)
 Dvpl конвертер на Go Lang. (С использованием кода на "С")

 > [!NOTE]
 > В данном конвертере для ускорения сжатия и улучшения его качества была использована библиотека [lz4](https://github.com/lz4/lz4).

 > [!WARNING]
 > Инструкции по сборке с использованием `dvpl_c` не будет, кому интересно, разбирайтесь с этим самостоятельно.

## Описание

### Поддерживаемые типы сжатия

- `0` - `none` - Сжатие полностью отсутствует.
- `1` - `lz4hc` - Более сильное и медленное чем lz4.
- `2` - `lz4` - Менее сильное и более быстрое чем lz4hc.
- `3` - `rfc1951` - Сжатие не используется в игре. (Его чтение сломано. Добавил только распаковку.)

### Вывод в консоль

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

### Описание команд
- `-c` - Сжатие в файлов в `.dvpl`.
- `-d` - Распаковка `.dvpl` файлов.
- `-i` - Входная директория или файл.
- `-o` - Выходная директория или файл.
- `-keep-original` - Сохранять оригинальный файл при распаковке или сжатии.
- `-compress` - Указывает уровень сжатия от 0 до 2.
    - `0` - `none`
    - `1` - `lz4hc`
    - `2` - `lz4`
- `-ignore` - Список шаблонов файлов, которые следует игнорировать, разделенный запятыми.
    #### Поддерживаются следующие символы подстановки:
    - `*` — любое количество символов (кроме `/`).
    - `?` — один символ.
    - `[abc]` — один из указанных символов.

    #### Примеры использования:
    - `*.exe` — игнорировать все `.exe` файлы.
    - `file?.log` — игнорировать файлы вида `file1.log`, `file2.log`.
    - `folder/*.txt` — игнорировать все `.txt` файлы в папке `folder`.
    - `data[1-3].csv` — игнорировать файлы `data1.csv`, `data2.csv`, `data3.csv`.
    - `image_[xyz].png` — игнорировать файлы `image_x.png`, `image_y.png`, `image_z.png`.

    #### Содержимое .dvpl_go.yml:
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
        

## Сборка проекта

### Пересборка .mod и .sum
1. Откройте консоль в проекте и выполните команды ниже.
2. `go mod init dvpl_go`
3. `go mod tidy`

---

### Сборка с иконкой
1. Установите на ПК `GO`.
2. Откройте `_build_.bat` в текстовом редакторе.
3. Отредактируйте путь до `ResourceHacker` либо закомментируйте его. (Он нужен только для иконки)
4. Запустите `_build_.bat`.
