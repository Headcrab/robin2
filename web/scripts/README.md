# Модульная структура JavaScript

Код разбит на логические модули для лучшей организации и поддержки.

## Структура модулей

### `core.js`
Основные функции приложения:
- `loadPage()` - загрузка страниц
- `initialize()` - инициализация приложения
- `saveParams()`, `restoreParams()` - работа с параметрами

### `ui.js`
Компоненты пользовательского интерфейса:
- `showLoader()`, `hideLoader()` - индикатор загрузки
- `initializeMobileMenu()` - мобильное меню
- `showErrorNotification()`, `showSuccessNotification()` - уведомления
- `setViewMode()` - переключение режимов просмотра

### `status.js`
Работа со статусом системы:
- `fetchStatus()` - получение статуса
- `updateSystemStatus()` - обновление индикаторов
- `loadStatistics()` - загрузка статистики
- `loadRecentActivity()` - последняя активность

### `data.js`
Работа с данными:
- `getTagOnDate()` - поиск данных по тегу
- `updateDataTable()` - обновление таблицы
- `loadHomePageData()` - данные главной страницы
- `parseDataString()` - парсинг данных

### `navigation.js`
Навигация:
- `updateBreadcrumb()` - обновление хлебных крошек

### `utils.js`
Утилитарные функции:
- `initializeRefreshButton()` - кнопка обновления
- `copyToClipboard()` - копирование в буфер

### `export.js`
Функции экспорта:
- `exportData()`, `exportTags()`, `exportLogs()` - экспорт данных
- `clearLogs()` - очистка логов

### `global.js`
Главный модуль, экспортирует все функции в `window` для совместимости с HTML

### `script.js`
Точка входа, импортирует `global.js`

## Использование

В HTML подключается только `script.js` как модуль:
```html
<script type="module" src="../scripts/script.js"></script>
```

Все функции доступны глобально через `window` для обратной совместимости:
```javascript
window.loadPage('/data/');
window.showSuccessNotification('Success!');
```

## Преимущества

1. **Модульность** - код разделён по логическим блокам
2. **Поддержка** - легче находить и исправлять ошибки
3. **Расширяемость** - просто добавлять новые функции
4. **Читаемость** - каждый модуль имеет чёткую ответственность
5. **Совместимость** - работает со старым HTML кодом 