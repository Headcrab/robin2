# Спецификация проекта Robin

Этот документ описывает конфигурацию, API и структуры данных для проекта Robin.

## 1. Конфигурация (`config/Robin.json`)

Основной конфигурационный файл приложения.

```json
{
    "port": 8008,
    "round": 2,
    "date_formats": [
        "2006-01-02 15:04:05",
        "..."
    ],
    "curr_db": "hs0",
    "db": [
        {
            "name": "hs0",
            "type": "mssql",
            "host": "hs0",
            "port": "1433",
            "user": "sa",
            "password": "...",
            "database": "Runtime",
            "timeout": 30,
            "connection_string": "...",
            "query": {
                "get_tag_date": "...",
                "get_tag_from_to": "...",
                "..."
            }
        }
    ],
    "curr_cache": "redis.localhost",
    "cache": [
        {
            "name": "redis.localhost",
            "type": "redis",
            "ttl": 1,
            "active": "false",
            "host": "localhost",
            "port": "6379",
            "..."
        }
    ]
}
```

### Основные параметры:

*   `port`: Порт, на котором запускается веб-сервер.
*   `round`: Количество знаков после запятой для округления числовых значений по умолчанию.
*   `date_formats`: Массив форматов дат, которые приложение пытается распознать.
*   `curr_db`: Имя текущей базы данных для использования из списка `db`.
*   `db`: Массив объектов, описывающих подключения к базам данных (поддерживаются `mssql`, `mysql`, `clickhouse`).
    *   Каждый объект содержит параметры подключения и маппинг `query` с именованными SQL-запросами.
*   `curr_cache`: Имя текущего кэша для использования из списка `cache`.
*   `cache`: Массив объектов, описывающих конфигурации кэша (поддерживаются `redis`, `memory`).

---

## 2. API Эндпоинты

### 2.1. Системные

#### `GET /api/info/`
**@Summary**: Получить информацию о приложении.
**@Description**: Возвращает имя, версию и время работы приложения.
**@Tags**: System

#### `GET /api/reload/`
**@Summary**: Перезагрузить конфигурацию.
**@Description**: Перечитывает конфигурационный файл `config/Robin.json`.
**@Tags**: System

#### `GET /api/log/`
**@Summary**: Получить лог.
**@Description**: Возвращает логи приложения.
**@Tags**: System
**@Parameters**:
*   `format` (query, string, optional): Формат вывода (`text`, `json`). По умолчанию `text`.

#### `GET /api/log/clear/`
**@Summary**: Очистить лог.
**@Description**: Очищает файл логов.
**@Tags**: System

#### `GET /api/status/`
**@Summary**: Получить статус сервера.
**@Description**: Возвращает статус сервера и текущей базы данных.
**@Tags**: System

### 2.2. Данные тегов

#### `GET /get/tag/`
**@Summary**: Получить значение тега.
**@Description**: Универсальный эндпоинт для получения данных по тегам.
**@Tags**: Tag
**@Parameters**:
*   `tag` (query, string, required): Имя тега. Можно несколько через запятую.
*   `date` (query, string, optional): Дата для получения значения (`YYYY-MM-DD HH:MM:SS`).
*   `from` (query, string, optional): Начало периода.
*   `to` (query, string, optional): Конец периода.
*   `group` (query, string, optional): Функция агрегации (`avg`, `sum`, `count`, `min`, `max`).
*   `count` (query, string, optional): Количество значений (для интерполяции).
*   `round` (query, int, optional): Округление, знаков после запятой.
*   `format` (query, string, optional): Формат вывода (`text`, `json`, `grafana`).

#### `GET /get/tag/list/`
**@Summary**: Получить список тегов.
**@Description**: Возвращает список тегов по маске.
**@Tags**: Tag
**@Parameters**:
*   `like` (query, string, optional): Маска для поиска тегов (например, `TAG*`).
*   `format` (query, string, optional): Формат вывода (`text`, `json`).

#### `GET /get/tag/up/`
**@Summary**: Получить даты включения оборудования.
**@Description**: Возвращает дату и время, когда значение тега изменилось на `1`.
**@Tags**: Tag
**@Parameters**:
*   `tag` (query, string, required): Имя тега.
*   `from` (query, string, required): Начало периода.
*   `to` (query, string, required): Конец периода.
*   `count` (query, int, optional): Порядковый номер события.

#### `GET /get/tag/down/`
**@Summary**: Получить даты выключения оборудования.
**@Description**: Возвращает дату и время, когда значение тега изменилось на `0`.
**@Tags**: Tag
**@Parameters**:
*   `tag` (query, string, required): Имя тега.
*   `from` (query, string, required): Начало периода.
*   `to` (query, string, required): Конец периода.
*   `count` (query, int, optional): Порядковый номер события.

### 2.3. Прочее

#### `GET /tag/decode/`
**@Summary**: Декодировать тег
**@Description**: Разбирает сложное имя тега на составляющие.
**@Tags**: Tag
**@Parameters**:
*   `tag` (query, string, required): Имя тега для разбора.

---

## 3. Структуры данных

Основные структуры данных, используемые в ответах API.

### `Tag`

Представляет одну точку данных для тега.

```json
{
  "name": "TEN_1.Value",
  "date": "2023-10-01T12:00:00Z",
  "value": 123.45
}
```

### `Tags`

Массив объектов `Tag`.

### `TimePoint` (пользовательский формат)

Группировка метрик по временным точкам.

```json
[
    {
        "time": "2023-10-01T12:00:00Z",
        "data": [
            {
                "name": "TEN_1.Value",
                "value": 123.45
            },
            {
                "name": "TEN_2.Value",
                "value": 234.56
            }
        ]
    }
]
```

### Grafana Time Series

Формат для интеграции с Grafana.

```json
[
  {
    "target": "TEN_1.Value",
    "datapoints": [
      [123.45, 1696161600000],
      [123.50, 1696161660000]
    ]
  }
]
``` 