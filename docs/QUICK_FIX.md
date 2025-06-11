# Быстрое исправление проблем с темами и переводами

## Проблема: Тема переключается не полностью

### Решение:
1. Откройте консоль браузера (F12)
2. Выполните команду: `debugThemeAndTranslations()`
3. Проверьте вывод и следуйте инструкциям ниже

### Возможные причины:

#### 1. CSS переменные не работают
```javascript
// в консоли браузера:
getComputedStyle(document.documentElement).getPropertyValue('--bg-primary')
```
Если пустой результат - добавьте `!important` к CSS переменным

#### 2. Tailwind классы переопределяют темную тему  
Решение: обновите `dark-theme.css` с `!important`:
```css
[data-theme="dark"] .bg-white {
    background-color: var(--bg-card) !important;
}
```

#### 3. Тема не применяется к элементам
```javascript
// принудительно установите тему:
document.documentElement.setAttribute('data-theme', 'dark');
```

## Проблема: Переводы не переключаются

### Диагностика:
```javascript
// в консоли браузера:
window.t('nav.home')           // должно вернуть перевод
window.getCurrentLanguage()    // текущий язык
checkTranslationIssues()       // найти проблемы
```

### Решения:

#### 1. Функции не загружены
```javascript
manualInit(); // принудительная инициализация
```

#### 2. Переводы не обновляются
```javascript
forceUpdateTranslations(); // принудительное обновление
```

#### 3. HTML элементы без data-i18n
Добавьте атрибуты:
```html
<span data-i18n="nav.home">Главная</span>
<input data-i18n-placeholder="common.search" placeholder="Поиск">
<button data-i18n-title="common.refresh" title="Обновить">
```

## Быстрые команды для отладки

```javascript
// Диагностика
debugThemeAndTranslations();
checkTranslationIssues();

// Тестирование
testThemeSwitching();
testLanguageSwitching();

// Принудительное исправление
manualInit();
forceUpdateTranslations();

// Ручное переключение
window.toggleTheme();
window.setLanguage('en');
window.setLanguage('ru');
```

## Проверка состояния

### Тема:
- Текущая тема: `window.getCurrentTheme()`
- Атрибут документа: `document.documentElement.getAttribute('data-theme')`
- CSS переменная: `getComputedStyle(document.documentElement).getPropertyValue('--bg-primary')`

### Язык:
- Текущий язык: `window.getCurrentLanguage()`
- Атрибут документа: `document.documentElement.getAttribute('lang')`
- Тест перевода: `window.t('nav.home')`

### DOM элементы:
- Кнопка темы: `document.getElementById('theme-toggle')`
- Кнопка языка: `document.getElementById('language-toggle')`
- Переводимые элементы: `document.querySelectorAll('[data-i18n]').length`

## Если ничего не помогает

1. **Перезагрузите страницу** с очисткой кэша (Ctrl+F5)

2. **Очистите localStorage**:
```javascript
localStorage.removeItem('theme');
localStorage.removeItem('language');
location.reload();
```

3. **Принудительная инициализация**:
```javascript
// в консоли после загрузки страницы:
setTimeout(() => {
    manualInit();
    setTimeout(() => {
        window.setTheme('dark');
        window.setLanguage('ru');
        forceUpdateTranslations();
    }, 500);
}, 1000);
```

4. **Проверьте ошибки в консоли** - могут быть проблемы с загрузкой модулей

## Временное решение

Если переключатели не работают, можно использовать прямые команды:

```javascript
// Переключение темы
document.documentElement.setAttribute('data-theme', 'dark');
// или
document.documentElement.setAttribute('data-theme', 'light');

// Переключение языка + обновление переводов
document.documentElement.setAttribute('lang', 'en');
if (window.i18nManager) {
    window.i18nManager.currentLang = 'en';
    window.i18nManager.updateTranslations();
}
``` 