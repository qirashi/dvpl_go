# dvpl_go [RU] | [EN](README_EN.md)

 > [!NOTE]
 > В данном конвертере для ускорения сжатия и улучшения его качества была использована библиотека [lz4](https://github.com/lz4/lz4).

## Как пользоваться конвертером?
* [Гайд на Русском](.readme/how_to_use.md)
* [Guide in English](.readme/how_to_use_en.md)

## Описание

### Поддерживаемые типы сжатия

- `0` - `none` - Сжатие полностью отсутствует.
- `1` - `lz4hc` - Более сильное и медленное чем lz4.
- `2` - `lz4` - Менее сильное и более быстрое чем lz4hc.
- `3` - `rfc1951` - Сжатие не используется в игре. (Его чтение вырезано)

### Вывод в консоль

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
- `-ignore` - Список шаблонов файлов, которые стоит игнорировать. (Файлы и расширения не будут обработаны)
- `-ignore-compress` - Список шаблонов файлов, которые принудительно будут сжаты в 0 тип. (Например `*.webp`)
- `-filter` - Список файлов шаблонов, которые будут обработаны. (Только файлы и расширения, которые будут обработаны, обратный от `-ignore`)
    - Например вам нужно распаковыать в отдельную папку только `*.webp` и `*.txt`.
    - Это будет выглядеть так: `dvpl -d -i ./in -o ./out -filter "*.webp,*.txt" -keep-original -m 4`
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

- `-m` - Максимальное количество параллельных обработчиков (workers).
    - По умолчанию: 1 (однопоточный режим)
    - Оптимальное значение: 2-4 (зависит от CPU)
    - При указании значений > максимума автоматически корректируется.
    - Максимальное кол-во зависит от ядер и потоков процессора.
    > **Возможны проблемы работы многопоточного режима на энергоэффективных ядрах от Intel.**

## Сравнение скорости работы и сжатия

### Этот конвертер на GoLang
```cmd
Начало:   16:2:41.15
Конец:    16:3:17.60
-----------------
Всего:    0 ч 0 мин 36.45 сек

Вес: 1,15 ГБ (1 244 843 076 байт)
```

### Этот конвертер на GoLang с многопотоком (2 workers)
```cmd
Начало:   16:4:43.85
Конец:    16:5:2.78
-----------------
Всего:    0 ч 0 мин 18.93 сек

Вес: 1,15 ГБ (1 244 843 076 байт)
```

### Другой конвертер на NodeJS
```cmd
Начало:   15:59:13.41
Конец:    16:0:10.19
-----------------
Всего:    0 ч 0 мин 56.78 сек

Вес: 1,15 ГБ (1 243 007 962 байт)
```

### Другой конвертер на GoLang
```cmd
Начало:   16:18:37.28
Конец:    16:18:43.51
-----------------
Всего:    0 ч 0 мин 6.23 сек

Вес: 2,81 ГБ (3 020 488 406 байт)
```

### Итоги
- Данный конвертер является оптимальным вариантом на данный момент. Он позволяет выбрать уровень сжатия, что напрямую сказывается на скорости сжатия файлов и поддерживает многопоточный режим. В режиме сжатия `lz4hc` он обгоняет своих одноклассников по скорости и не уступает в качестве сжатия. Другой конвертер на GoLang использовал `lz4`, что сжимает хуже, но быстрее (сменить режим сжатия было нельзя). Этот же конвертер работает быстро и поддерживает все доступные методы.
