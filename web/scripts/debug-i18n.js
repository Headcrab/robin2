// отладка переводов

function debugTranslations() {
    console.log('=== отладка переводов ===');
    
    if (!window.i18nManager) {
        console.error('i18nManager не найден');
        return;
    }
    
    console.log('i18nManager:', window.i18nManager);
    console.log('currentLang:', window.i18nManager.currentLang);
    console.log('translations:', window.i18nManager.translations);
    
    if (window.i18nManager.translations) {
        console.log('доступные языки:', Object.keys(window.i18nManager.translations));
        
        if (window.i18nManager.translations.ru) {
            console.log('русские переводы:', window.i18nManager.translations.ru);
            console.log('home.title:', window.i18nManager.translations.ru['home.title']);
        }
    }
    
    // тест ключей
    const testKeys = ['home.title', 'common.search', 'common.system'];
    testKeys.forEach(key => {
        const result = window.i18nManager.t(key);
        console.log(`${key} -> ${result}`);
    });
}

// принудительная загрузка переводов
function forceLoadTranslations() {
    if (!window.i18nManager) {
        console.error('i18nManager не найден');
        return;
    }
    
    // прямая установка переводов
    window.i18nManager.translations = {
        ru: {
            'home.title': 'Система получения данных АСУТП',
            'home.subtitle': 'Современный интерфейс для мониторинга и управления промышленными данными',
            'common.search': 'Поиск по системе...',
            'common.system': 'Система',
            'common.refresh': 'Обновить данные',
            'common.quick_actions': 'Быстрые действия',
            'home.stats.active_tags': 'Активных тегов',
            'home.stats.data_records': 'Записей данных',
            'home.stats.system_status': 'Статус системы',
            'status.checking': 'Проверка...',
            'home.recent_activity': 'Последняя активность',
            'home.system_health': 'Состояние системы'
        }
    };
    
    console.log('переводы принудительно загружены');
    
    // обновляем интерфейс
    window.i18nManager.updateTranslations();
    
    console.log('интерфейс обновлен');
}

window.debugTranslations = debugTranslations;
window.forceLoadTranslations = forceLoadTranslations;

console.log('debug-i18n загружен. функции: debugTranslations(), forceLoadTranslations()'); 