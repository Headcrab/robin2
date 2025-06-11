# Система переключения тем и языков

## Обзор

В приложение добавлена система переключения между темами (светлая/темная) и языками (русский/казахский/английский).

## Файлы системы

### JavaScript модули
- `themes.js` - управление темами
- `i18n.js` - интернационализация  
- `ui.js` - обновлен для переключателей

### CSS
- `dark-theme.css` - стили темной темы

### HTML шаблоны
- `base.html` - подключение CSS
- `header.html` - placeholder для переключателей
- `home_i18n.html` - пример интернационализации

## Использование

### Переключение тем

```javascript
// переключить тему
window.toggleTheme();

// установить конкретную тему
window.setTheme('dark'); // или 'light'

// получить текущую тему
const currentTheme = window.getCurrentTheme();

// получить список доступных тем
const themes = window.getThemes();
```

### Переключение языков

```javascript
// установить язык
window.setLanguage('ru'); // 'ru', 'kk', 'en'

// получить текущий язык
const currentLang = window.getCurrentLanguage();

// получить список языков
const languages = window.getLanguages();

// перевести текст
const translated = window.t('nav.home');

// обновить переводы на странице
window.updateTranslations();
```

### Интернационализация в HTML

Используйте data-атрибуты для автоматического перевода:

```html
<!-- текст элемента -->
<span data-i18n="nav.home">Главная</span>

<!-- placeholder input -->
<input data-i18n-placeholder="common.search" placeholder="Поиск" />

<!-- title атрибут -->
<button data-i18n-title="common.refresh" title="Обновить">
  Refresh
</button>
```

## Структура переводов

Переводы организованы иерархически:

```javascript
{
  ru: {
    nav: {
      home: 'Главная',
      data: 'Данные'
    },
    common: {
      loading: 'Загрузка...',
      error: 'Ошибка'
    }
  }
}
```

## Автоматические функции

### Сохранение настроек
- Тема и язык сохраняются в localStorage
- Восстанавливаются при загрузке страницы

### Определение браузерного языка
```javascript
// автоматически определяется при первом запуске
// приоритет: localStorage > браузерный язык > русский
```

### Определение системной темы
```javascript
// поддержка prefers-color-scheme
// автоматическое переключение при изменении системной темы
```

## События

Система генерирует события для интеграции с другими компонентами:

```javascript
// смена темы
window.addEventListener('themeChanged', (event) => {
  console.log('Новая тема:', event.detail.theme);
});

// смена языка
window.addEventListener('languageChanged', (event) => {
  console.log('Новый язык:', event.detail.language);
});
```

## Стили темной темы

CSS переменные для кастомизации:

```css
:root[data-theme="dark"] {
  --bg-primary: #1a1b23;
  --text-primary: #e2e4e9;
  --accent-primary: #3b82f6;
  /* и другие переменные */
}
```

## Добавление новых переводов

1. Добавьте ключи в `i18n.js`:
```javascript
this.translations = {
  ru: { 'new.key': 'Новый текст' },
  kk: { 'new.key': 'Жаңа мәтін' },
  en: { 'new.key': 'New text' }
};
```

2. Используйте в HTML:
```html
<span data-i18n="new.key">Новый текст</span>
```

3. Или в JavaScript:
```javascript
const text = window.t('new.key');
```

## Добавление новых языков

1. Добавьте язык в `i18n.js`:
```javascript
this.languages = {
  // существующие языки...
  de: {
    name: 'Deutsch',
    nativeName: 'Deutsch',
    flag: '🇩🇪'
  }
};
```

2. Добавьте переводы:
```javascript
this.translations = {
  // существующие переводы...
  de: {
    'nav.home': 'Startseite',
    // другие переводы...
  }
};
```

## Интеграция в существующие страницы

1. Добавьте CSS темной темы в `<head>`:
```html
<link rel="stylesheet" href="../css/dark-theme.css" />
```

2. Добавьте data-атрибуты к элементам:
```html
<h1 data-i18n="page.title">Заголовок</h1>
```

3. Переключатели добавляются автоматически при загрузке страницы.

## Отладка

Используйте консоль браузера:
```javascript
// проверить текущие настройки
console.log('Тема:', window.getCurrentTheme());
console.log('Язык:', window.getCurrentLanguage());

// принудительно обновить переводы
window.updateTranslations();
``` 