<div align="center">
  <h1>dvpl_go [RU] | <a href="README_EN.md">EN</a></h1>
  <br>
  <img src="https://img.shields.io/github/downloads/qirashi/dvpl_go/total?logo=github&label=Downloads&style=for-the-badge" alt="Downloads">
  <img src="https://img.shields.io/github/license/qirashi/dvpl_go?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/github/last-commit/qirashi/dvpl_go?style=for-the-badge" alt="Last Commit">
  <img src="https://img.shields.io/github/languages/code-size/qirashi/dvpl_go?style=for-the-badge&label=Code%20Size" alt="Code Size">
  <a href="https://github.com/qirashi/dvpl_go/stargazers">
    <img src="https://img.shields.io/github/stars/qirashi/dvpl_go?style=for-the-badge&logo=github&color=ffca28&labelColor=24292e" alt="Stars">
  </a>
</div>
<br>

  > [!NOTE]
  > Конвертер использует библиотеку [lz4](https://github.com/lz4/lz4) для повышения скорости и качества сжатия.  
  > Формат имеет ограничения по размеру сжимаемых данных!  

## Как использовать?
  > [!TIP]  
  > [Гайд по использованию конвертера на Русском](.readme/how_to_use.md)  
  > [A guide to using the converter in English](.readme/how_to_use_en.md)  

## Поддерживаемые типы сжатия
  > [!NOTE]  
  > | Тип  | Название |                 Описание                 |
  > |------|----------|------------------------------------------|
  > |  0   |   none   | Сжатие полностью отсутствует.            |
  > |  1   |   lz4hc  | Более сильное и медленное чем lz4.       |
  > |  2   |   lz4    | Менее сильное и более быстрое чем lz4hc. |

## Переменные среды
  В переменных среды могут храниться 2 настройки конвертера `DVPL_MAX_WORKERS` и `DVPL_COMPRESS_TYPE` для указания количества параллельно работающих процессов и тип сжатия соответственно.

- `DVPL_MAX_WORKERS` — Максимальное количество параллельных обработчиков. (Автоматически ограничивается при превышении допустимого значения.)
- `DVPL_COMPRESS_TYPE` — Указывает уровень сжатия от 0 до 2. (В случае несуществующего типа, будет ошибка.)

Как задать:
1. **Создать вручную**
    - Нажать `Win+R`.
    - Выполнить `SystemPropertiesAdvanced`.
    - Открыть `Переменные среды...` и создать соответствующие переменные.

2. **Через командную строку**
    - `Пуск` → `cmd` → `правой кнопкой` → `Запуск от имени администратора`
    - Вставьте одну из команд:
      *   Для одного пользователя:
            ```cmd
            setx DVPL_MAX_WORKERS 4
            ```
            ```cmd
            setx DVPL_COMPRESS_TYPE 1
            ```

## CMD
```cmd
R:\Github\dvpl_go\out>dvpl.exe -h

DvplGo 2.1.5 x64 | Copyright (c) 2026 Qirashi

Usage: dvpl (-c|-d) -i <path> [options]
[main options]
  -c               Compress files into .dvpl format.
  -d               Decompress .dvpl files.
  -i <path>        Input file or directory.
  -o <path>        Output file or directory (default: same as -i).

[general options]
  -filter <masks>  Process only files matching given patterns, e.g. "*.sc2,*.scg".
  -ignore <masks>  Skip files matching given patterns, e.g. "*.exe,*.dll".
  -keep-original   Do not delete original files after processing.
  -m <number>      Max parallel workers (default 2, max 12).
  -trust-data      Skip CRC and some integrity checks.

[compress options]
  -compress <type>  Compression type: 0=none, 1=lz4hc, 2=lz4, (default 1).
  -forced-compress  Force compression even if result is larger than original.
  -ignore-compress <masks>
                    Disable compression for files matching these patterns, e.g. "*.webp".

[examples]
  Compress   : dvpl -c -i ./input -compress 1
  Decompress : dvpl -d -i ./input -o ./output
  Filter     : dvpl -d -i ./input -o ./output -filter "*.sc2,*.scg"
  Ignore     : dvpl -c -i ./input -ignore "*.exe,*.dll"
```

### Описание команд
- `-c` — Сжатие в `.dvpl`.
- `-d` — Распаковка `.dvpl`.
- `-i` — Входная директория или файл.
- `-o` — Выходная директория или файл.
- `-keep-original` — Сохранять оригинальный файл при распаковке или сжатии.
- `-compress` — Указывает уровень сжатия от 0 до 2.
    - `0` — `none`
    - `1` — `lz4hc`
    - `2` — `lz4`
- `-ignore` — Список шаблонов файлов, которые стоит игнорировать. (Файлы и расширения не будут обработаны)
- `-ignore-compress` — Список шаблонов файлов, которые принудительно будут сжаты в 0 тип. (Например `*.webp`)
- `-filter` — Список файлов шаблонов, которые будут обработаны. (Только файлы и расширения, которые будут обработаны, обратный от `-ignore`)
    - Например вам нужно распаковать в отдельную папку только `*.webp` и `*.txt`.
    - Это будет выглядеть так: `dvpl -d -i ./in -o ./out -filter "*.webp,*.txt" -keep-original -m 4`
    #### Символы подстановки для фильтров:
    - `*` — любое количество символов (кроме `/`).
    - `?` — один символ.
    - `[abc]` — один из указанных символов.

    #### Примеры:
    - `*.exe` — игнорировать все `.exe` файлы.
    - `file?.log` — игнорировать файлы вида `file1.log`, `file2.log`.
    - `data[1-3].csv` — игнорировать файлы `data1.csv`, `data2.csv`, `data3.csv`.
    - `image_[xyz].png` — игнорировать файлы `image_x.png`, `image_y.png`, `image_z.png`.

- `-m` - Максимальное количество параллельных обработчиков (workers).
    - По умолчанию: 2 (однопоточный режим)
    - Оптимальное значение: 2-4 (зависит от CPU)
    - При указании значений > максимума автоматически корректируется.
    - Максимальное кол-во зависит от ядер и потоков процессора.

- `-trust-data` - CRC и некоторые проверки игнорируются (Используется для ускорения распаковки при работе с большим количеством данных или при неудачной попытке распаковать файл).

## Сравнение скорости работы

### Этот конвертер на GoLang с многопотоком (2 workers) (lz4hc)
```
Начало:   16:4:43.85
Конец:    16:5:2.78
-----------------
Всего:    0 ч 0 мин 18.93 сек

Вес: 1,15 ГБ (1 244 843 076 байт)
```

### Другой конвертер на NodeJS (lz4hc)
```
Начало:   15:59:13.41
Конец:    16:0:10.19
-----------------
Всего:    0 ч 0 мин 56.78 сек

Вес: 1,15 ГБ (1 243 007 962 байт)
```

### Другой конвертер на GoLang (lz4)
```
Начало:   16:18:37.28
Конец:    16:18:43.51
-----------------
Всего:    0 ч 0 мин 6.23 сек

Вес: 2,81 ГБ (3 020 488 406 байт)
```

## Итоги
  Данный конвертер является оптимальным вариантом сжатия и скорости. Он позволяет выбрать уровень сжатия, что сказывается на скорости. В режиме сжатия `lz4hc` он обгоняет свои аналоги по скорости и не уступает в качестве сжатия. Другой конвертер на Go использовал `lz4`, что сжимает хуже, но быстрее. Этот же конвертер работает быстро и поддерживает все основные доступные методы.
