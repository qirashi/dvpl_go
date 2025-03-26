# dvpl_go [RU] | [EN](README_EN.md)
 Dvpl конвертер на Go Lang. (С использованием кода на "С")

 > [!NOTE]
 > В данном конвертере для ускорения сжатия и улучшения его качества была использована библиотека [lz4](https://github.com/lz4/lz4).

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
  -c    Compress .dvpl files
  -compress int
        Compression type: 0 (none), 1 (lz4hc), 2 (lz4) | (default 1)
  -d    Decompress .dvpl files
  -i string
        Input path (file or directory)
  -ignore string
        Comma-separated list of file patterns to ignore
  -keep-original
        Keep original files
  -m int
        Maximum number of parallel workers. When used, 2 are recommended, with a maximum of 6. (default 1)
  -o string
        Output path (file or directory)

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

- `-m` - Максимальное количество параллельных обработчиков (workers).
    - По умолчанию: 1 (однопоточный режим)
    - Оптимальное значение: 2-4 (зависит от CPU)
    - При указании значений > максимума автоматически корректируется.
    - Максимальное кол-во зависит от ядер и потоков процессора.
    > **Возможны проблемы работы многопоточного режима на энергоэффективных ядрах от Intel.**
