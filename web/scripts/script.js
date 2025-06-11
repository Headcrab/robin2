// модульная система подключена через global.js
// старый монолитный код разбит на модули:
// - core.js - основные функции загрузки страниц
// - ui.js - интерфейс, уведомления, мобильное меню
// - status.js - работа со статусом системы
// - data.js - работа с данными и таблицами
// - navigation.js - навигация и breadcrumbs
// - utils.js - утилитарные функции
// - export.js - функции экспорта
// - global.js - главный файл с глобальными экспортами

// подключаем главный модуль
import './global.js';

// подключаем отладку в режиме разработки (раскомментируйте для отладки)
// import './debug.js';
// import './theme-debug.js';
import './fix-translation.js';
import './debug-i18n.js';
import './fix-navigation-translation.js';
import './fix-table-theme.js';
import './fix-breadcrumb.js'; 