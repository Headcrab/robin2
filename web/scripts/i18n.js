// internationalization module

class I18nManager {
    constructor() {
        this.currentLang = 'ru';
        this.translations = {};
        this.languages = {
            ru: {
                name: 'Русский',
                nativeName: 'Русский',
                flag: '🇷🇺'
            },
            kk: {
                name: 'Қазақша',
                nativeName: 'Қазақша', 
                flag: '🇰🇿'
            },
            en: {
                name: 'English',
                nativeName: 'English',
                flag: '🇺🇸'
            }
        };
        this.init();
    }

    async init() {
        console.log('инициализация i18nManager...');
        
        // load translations first
        await this.loadTranslations();
        console.log('переводы загружены');
        
        // load saved language from localStorage
        const savedLang = localStorage.getItem('language') || this.detectBrowserLanguage();
        console.log('установка языка:', savedLang);
        
        this.setLanguage(savedLang);
        
        console.log('i18nManager инициализирован');
    }

    detectBrowserLanguage() {
        const browserLang = navigator.language || navigator.userLanguage;
        if (browserLang.startsWith('kk')) return 'kk';
        if (browserLang.startsWith('en')) return 'en';
        return 'ru'; // default
    }

    async loadTranslations() {
        // define translations inline for now - flat structure
        this.translations = {
            ru: {
                nav: {
                    home: 'Главная',
                    data: 'Данные',
                    charts: 'Графики',
                    tags: 'Теги',
                    logs: 'Логи',
                    docs: 'Документация',
                    api: 'API'
                },
                
                common: {
                    loading: 'Загрузка...',
                    error: 'Ошибка',
                    success: 'Успешно',
                    search: 'Поиск по системе...',
                    export: 'Экспорт',
                    clear: 'Очистить',
                    refresh: 'Обновить данные',
                    close: 'Закрыть',
                    cancel: 'Отмена',
                    save: 'Сохранить',
                    delete: 'Удалить',
                    edit: 'Редактировать',
                    view: 'Просмотр',
                    copy: 'Копировать',
                    system: 'Система',
                    quick_actions: 'Быстрые действия'
                },
                
                home: {
                    title: 'Система получения данных АСУТП',
                    subtitle: 'Современный интерфейс для мониторинга и управления промышленными данными',
                    recent_activity: 'Последняя активность',
                    system_health: 'Состояние системы',
                    stats: {
                        active_tags: 'Активных тегов',
                        data_records: 'Записей данных',
                        system_status: 'Статус системы'
                    }
                },
                
                data: {
                    title: 'Данные АСУТП',
                    subtitle: 'Поиск и анализ данных по тегам',
                    search_params: 'Параметры поиска',
                    tag: 'Тег',
                    tag_placeholder: 'Введите название тега',
                    date_from: 'Дата начала',
                    date_to: 'Дата окончания',
                    record_count: 'Количество записей',
                    find_data: 'Найти данные',
                    search_results: 'Результаты поиска',
                    export_data: 'Экспорт данных',
                    table: {
                        time: 'Время',
                        tag: 'Тег',
                        value: 'Значение',
                        quality: 'Качество',
                        unit: 'Единица',
                        description: 'Описание'
                    }
                },

                charts: {
                    title: 'Графики',
                    subtitle: 'Расширенный анализ трендов: X/Y масштаб, панорамирование, сглаживание и управление диапазоном',
                    search_params: 'Параметры графика',
                    tag_input: 'Тег',
                    tag_placeholder: 'Введите имя тега (например A20_WT_01)',
                    add_tag: 'Добавить',
                    search_mask: 'Маска поиска тегов',
                    find_tags: 'Найти теги',
                    date_from: 'Дата начала',
                    date_to: 'Дата окончания',
                    points_count: 'Количество точек',
                    draw_chart: 'Построить график',
                    selected_tags: 'Выбранные теги',
                    clear_tags: 'Очистить теги',
                    range_left: 'Сместить влево',
                    range_right: 'Сместить вправо',
                    range_expand: 'Расширить диапазон',
                    range_shrink: 'Сузить диапазон',
                    found_tags: 'Найденные теги',
                    result_chart: 'График',
                    zoom_x: 'Масштаб X',
                    zoom_y: 'Масштаб Y',
                    reset_view: 'Сбросить вид',
                    scale_mode: 'Шкалы',
                    scale_common: 'Общая шкала',
                    scale_per_tag: 'Отдельная шкала на тег',
                    curve_mode: 'Линии',
                    curve_linear: 'Ломаная',
                    curve_smooth: 'Сглаженная',
                    smoothing: 'Сглаживание (окно)',
                    show_points: 'Показывать точки',
                    show_values: 'Показывать значения',
                    scroll_hint: 'Панорамирование: перетаскивание мышью (Shift + drag для Y), скролл по X. Масштаб: Ctrl + колесо (X) и Alt + колесо (Y).',
                    empty: 'Выберите теги и нажмите "Построить график"'
                },
                
                tags: {
                    title: 'Управление тегами',
                    subtitle: 'Просмотр и поиск доступных тегов системы',
                    search: 'Поиск тегов',
                    search_mask: 'Маска поиска',
                    search_placeholder: 'Введите маску для поиска тегов (например: A20_*)',
                    find_tags: 'Найти теги',
                    found_tags: 'Найденные теги'
                },
                
                logs: {
                    title: 'Логи системы',
                    subtitle: 'Просмотр системных логов и событий',
                    log_entries: 'Записи логов',
                    refresh_logs: 'Обновить логи',
                    export_logs: 'Экспорт логов',
                    clear_logs: 'Очистить логи'
                },
                
                docs: {
                    title: 'Документация',
                    subtitle: 'Техническая документация и руководства проекта',
                    available_docs: 'Доступные документы',
                    back_to_list: 'Назад к списку'
                },
                
                api: {
                    title: 'API документация'
                },
                
                status: {
                    working: 'Работает',
                    error: 'Ошибка',
                    checking: 'Проверка...',
                    connection_error: 'Ошибка связи',
                    system_status: 'Статус системы'
                },
                
                theme: {
                    light: 'Светлая',
                    dark: 'Темная',
                    toggle: 'Переключить тему'
                },
                
                lang: {
                    switch: 'Сменить язык'
                }
            },
            
            kk: {
                nav: {
                    home: 'Басты бет',
                    data: 'Деректер',
                    charts: 'Графиктер',
                    tags: 'Тегтер',
                    logs: 'Логтер',
                    docs: 'Құжаттама',
                    api: 'API'
                },
                
                common: {
                    loading: 'Жүктелуде...',
                    error: 'Қате',
                    success: 'Сәтті',
                    search: 'Іздеу...',
                    export: 'Экспорт',
                    clear: 'Тазарту',
                    refresh: 'Деректерді жаңарту',
                    close: 'Жабу',
                    cancel: 'Болдырмау',
                    save: 'Сақтау',
                    delete: 'Жою',
                    edit: 'Өңдеу',
                    view: 'Қарау',
                    copy: 'Көшіру',
                    system: 'Жүйе',
                    quick_actions: 'Жылдам әрекеттер'
                },
                
                home: {
                    title: 'АСУТП деректерін алу жүйесі',
                    subtitle: 'Өндірістік деректерді бақылау және басқару үшін заманауи интерфейс',
                    recent_activity: 'Соңғы әрекет',
                    system_health: 'Жүйе денсаулығы',
                    stats: {
                        active_tags: 'Белсенді тегтер',
                        data_records: 'Деректер жазбалары',
                        system_status: 'Жүйе күйі'
                    }
                },
                
                data: {
                    title: 'АСУТП деректері',
                    subtitle: 'Тегтер бойынша деректерді іздеу және талдау',
                    search_params: 'Іздеу параметрлері',
                    tag: 'Тег',
                    tag_placeholder: 'Тег атауын енгізіңіз',
                    date_from: 'Басталу күні',
                    date_to: 'Аяқталу күні',
                    record_count: 'Жазбалар саны',
                    find_data: 'Деректерді табу',
                    search_results: 'Іздеу нәтижелері',
                    export_data: 'Деректерді экспорттау',
                    table: {
                        time: 'Уақыт',
                        tag: 'Тег',
                        value: 'Мән',
                        quality: 'Сапа',
                        unit: 'Өлшем бірлігі',
                        description: 'Сипаттама'
                    }
                },

                charts: {
                    title: 'Графиктер',
                    subtitle: 'Трендтерді кеңейтілген талдау: X/Y масштабы, панорамалау, тегістеу және диапазонды басқару',
                    search_params: 'График параметрлері',
                    tag_input: 'Тег',
                    tag_placeholder: 'Тег атауын енгізіңіз (мысалы A20_WT_01)',
                    add_tag: 'Қосу',
                    search_mask: 'Тег іздеу маскасы',
                    find_tags: 'Тегтерді табу',
                    date_from: 'Басталу күні',
                    date_to: 'Аяқталу күні',
                    points_count: 'Нүктелер саны',
                    draw_chart: 'График салу',
                    selected_tags: 'Таңдалған тегтер',
                    clear_tags: 'Тегтерді тазарту',
                    range_left: 'Солға жылжыту',
                    range_right: 'Оңға жылжыту',
                    range_expand: 'Диапазонды кеңейту',
                    range_shrink: 'Диапазонды тарылту',
                    found_tags: 'Табылған тегтер',
                    result_chart: 'График',
                    zoom_x: 'X масштабы',
                    zoom_y: 'Y масштабы',
                    reset_view: 'Көріністі қалпына келтіру',
                    scale_mode: 'Шкалалар',
                    scale_common: 'Ортақ шкала',
                    scale_per_tag: 'Әр тегке бөлек шкала',
                    curve_mode: 'Сызықтар',
                    curve_linear: 'Сынық сызық',
                    curve_smooth: 'Тегістелген',
                    smoothing: 'Тегістеу (терезе)',
                    show_points: 'Нүктелерді көрсету',
                    show_values: 'Мәндерді көрсету',
                    scroll_hint: 'Панорамалау: тінтуірмен сүйреу (Y үшін Shift + drag), X бойынша айналдыру. Масштаб: Ctrl + дөңгелек (X) және Alt + дөңгелек (Y).',
                    empty: 'Тегтерді таңдап, "График салу" түймесін басыңыз'
                },
                
                tags: {
                    title: 'Тегтерді басқару',
                    subtitle: 'Жүйенің қолжетімді тегтерін қарау және іздеу',
                    search: 'Тегтерді іздеу',
                    search_mask: 'Іздеу маскасы',
                    search_placeholder: 'Тегтерді іздеу үшін маска енгізіңіз (мысалы: A20_*)',
                    find_tags: 'Тегтерді табу',
                    found_tags: 'Табылған тегтер'
                },
                
                logs: {
                    title: 'Жүйе логтары',
                    subtitle: 'Жүйелік логтар мен оқиғаларды қарау',
                    log_entries: 'Лог жазбалары',
                    refresh_logs: 'Логтарды жаңарту',
                    export_logs: 'Логтарды экспорттау',
                    clear_logs: 'Логтарды тазарту'
                },
                
                docs: {
                    title: 'Құжаттама',
                    subtitle: 'Техникалық құжаттама және жоба нұсқаулықтары',
                    available_docs: 'Қолжетімді құжаттар',
                    back_to_list: 'Тізімге қайту'
                },
                
                api: {
                    title: 'API құжаттамасы'
                },
                
                status: {
                    working: 'Жұмыс істеп тұр',
                    error: 'Қате',
                    checking: 'Тексерілуде...',
                    connection_error: 'Байланыс қатесі',
                    system_status: 'Жүйе күйі'
                },
                
                theme: {
                    light: 'Ашық',
                    dark: 'Қараңғы',
                    toggle: 'Тақырыпты ауыстыру'
                },
                
                lang: {
                    switch: 'Тілді ауыстыру'
                }
            },
            
            en: {
                nav: {
                    home: 'Home',
                    data: 'Data',
                    charts: 'Charts',
                    tags: 'Tags',
                    logs: 'Logs',
                    docs: 'Documentation',
                    api: 'API'
                },
                
                common: {
                    loading: 'Loading...',
                    error: 'Error',
                    success: 'Success',
                    search: 'Search...',
                    export: 'Export',
                    clear: 'Clear',
                    refresh: 'Refresh data',
                    close: 'Close',
                    cancel: 'Cancel',
                    save: 'Save',
                    delete: 'Delete',
                    edit: 'Edit',
                    view: 'View',
                    copy: 'Copy',
                    system: 'System',
                    quick_actions: 'Quick actions'
                },
                
                home: {
                    title: 'SCADA Data Acquisition System',
                    subtitle: 'Modern interface for monitoring and managing industrial data',
                    recent_activity: 'Recent Activity',
                    system_health: 'System Health',
                    stats: {
                        active_tags: 'Active Tags',
                        data_records: 'Data Records',
                        system_status: 'System Status'
                    }
                },
                
                data: {
                    title: 'SCADA Data',
                    subtitle: 'Search and analyze tag data',
                    search_params: 'Search Parameters',
                    tag: 'Tag',
                    tag_placeholder: 'Enter tag name',
                    date_from: 'Start Date',
                    date_to: 'End Date',
                    record_count: 'Record Count',
                    find_data: 'Find Data',
                    search_results: 'Search Results',
                    export_data: 'Export data',
                    table: {
                        time: 'Time',
                        tag: 'Tag',
                        value: 'Value',
                        quality: 'Quality',
                        unit: 'Unit',
                        description: 'Description'
                    }
                },

                charts: {
                    title: 'Charts',
                    subtitle: 'Advanced trend analysis: X/Y zoom, panning, smoothing, and fetch-range controls',
                    search_params: 'Chart Parameters',
                    tag_input: 'Tag',
                    tag_placeholder: 'Enter tag name (e.g. A20_WT_01)',
                    add_tag: 'Add',
                    search_mask: 'Tag search mask',
                    find_tags: 'Find tags',
                    date_from: 'Start Date',
                    date_to: 'End Date',
                    points_count: 'Points Count',
                    draw_chart: 'Draw Chart',
                    selected_tags: 'Selected Tags',
                    clear_tags: 'Clear tags',
                    range_left: 'Shift left',
                    range_right: 'Shift right',
                    range_expand: 'Expand range',
                    range_shrink: 'Shrink range',
                    found_tags: 'Found Tags',
                    result_chart: 'Chart',
                    zoom_x: 'Zoom X',
                    zoom_y: 'Zoom Y',
                    reset_view: 'Reset View',
                    scale_mode: 'Scales',
                    scale_common: 'Common scale',
                    scale_per_tag: 'Per-tag scale',
                    curve_mode: 'Lines',
                    curve_linear: 'Linear',
                    curve_smooth: 'Smoothed',
                    smoothing: 'Smoothing (window)',
                    show_points: 'Show points',
                    show_values: 'Show values',
                    scroll_hint: 'Pan by dragging (Shift + drag for Y), scroll on X. Zoom with Ctrl + wheel (X) and Alt + wheel (Y).',
                    empty: 'Select tags and click "Draw Chart"'
                },
                
                tags: {
                    title: 'Tag Management',
                    subtitle: 'View and search available system tags',
                    search: 'Tag Search',
                    search_mask: 'Search Mask',
                    search_placeholder: 'Enter mask to search tags (e.g.: A20_*)',
                    find_tags: 'Find Tags',
                    found_tags: 'Found Tags'
                },
                
                logs: {
                    title: 'System Logs',
                    subtitle: 'View system logs and events',
                    log_entries: 'Log Entries',
                    refresh_logs: 'Refresh logs',
                    export_logs: 'Export logs',
                    clear_logs: 'Clear logs'
                },
                
                docs: {
                    title: 'Documentation',
                    subtitle: 'Technical documentation and project guides',
                    available_docs: 'Available Documents',
                    back_to_list: 'Back to List'
                },
                
                api: {
                    title: 'API Documentation'
                },
                
                status: {
                    working: 'Working',
                    error: 'Error',
                    checking: 'Checking...',
                    connection_error: 'Connection Error',
                    system_status: 'System Status'
                },
                
                theme: {
                    light: 'Light',
                    dark: 'Dark',
                    toggle: 'Toggle Theme'
                },
                
                lang: {
                    switch: 'Switch Language'
                }
            }
        };
    }

    setLanguage(lang) {
        if (!this.languages[lang]) {
            console.warn(`Language ${lang} not found`);
            return;
        }

        this.currentLang = lang;
        
        // save to localStorage
        localStorage.setItem('language', lang);
        
        // update document lang attribute
        document.documentElement.setAttribute('lang', lang);
        
        // update all translatable elements
        this.updateTranslations();
        
        // update language selector
        this.updateLanguageSelector();
        
        // dispatch event for other components
        window.dispatchEvent(new CustomEvent('languageChanged', {
            detail: { language: lang }
        }));
        
        // update theme button title with new language
        if (window.themeManager) {
            window.themeManager.updateThemeButton();
        }
        
        console.log(`Language changed to: ${lang}`);
    }

    updateTranslations() {
        // find all elements with data-i18n attribute
        const elements = document.querySelectorAll('[data-i18n]');
        elements.forEach(element => {
            const key = element.getAttribute('data-i18n');
            const translation = this.t(key);
            if (translation) {
                element.textContent = translation;
            }
        });

        // update placeholders
        const placeholderElements = document.querySelectorAll('[data-i18n-placeholder]');
        placeholderElements.forEach(element => {
            const key = element.getAttribute('data-i18n-placeholder');
            const translation = this.t(key);
            if (translation) {
                element.setAttribute('placeholder', translation);
            }
        });

        // update titles  
        const titleElements = document.querySelectorAll('[data-i18n-title]');
        titleElements.forEach(element => {
            const key = element.getAttribute('data-i18n-title');
            const translation = this.t(key);
            if (translation) {
                element.setAttribute('title', translation);
            }
        });
    }

    updateLanguageSelector() {
        const langBtn = document.getElementById('language-toggle');
        if (langBtn) {
            const currentLangData = this.languages[this.currentLang];
            langBtn.innerHTML = `
                <span class="mr-1">${currentLangData.flag}</span>
                <span class="hidden sm:inline text-sm">${this.currentLang.toUpperCase()}</span>
                <svg class="ml-1 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
            `;
            langBtn.title = this.t('lang.switch');
        }

        // update dropdown if exists
        const dropdown = document.getElementById('language-dropdown');
        if (dropdown) {
            Object.keys(this.languages).forEach(lang => {
                const option = dropdown.querySelector(`[data-lang="${lang}"]`);
                if (option) {
                    option.classList.toggle('active', lang === this.currentLang);
                    if (lang === this.currentLang) {
                        option.style.backgroundColor = 'var(--accent-primary)';
                        option.style.color = 'white';
                    } else {
                        option.style.backgroundColor = '';
                        option.style.color = '';
                    }
                }
            });
        }
    }

    t(key, params = {}) {
        if (!this.translations || !this.translations[this.currentLang]) {
            console.warn(`Translations not loaded for language: ${this.currentLang}`);
            return key;
        }
        
        const keys = key.split('.');
        let value = this.translations[this.currentLang];
        
        for (const k of keys) {
            if (value && typeof value === 'object') {
                value = value[k];
            } else {
                value = undefined;
                break;
            }
        }
        
        if (typeof value !== 'string') {
            console.warn(`Translation not found: ${key} for language: ${this.currentLang}`, this.translations[this.currentLang]);
            return key; // fallback to key
        }
        
        // simple parameter replacement
        let result = value;
        Object.keys(params).forEach(param => {
            result = result.replace(new RegExp(`{${param}}`, 'g'), params[param]);
        });
        
        return result;
    }

    getCurrentLanguage() {
        return this.currentLang;
    }

    getLanguages() {
        return this.languages;
    }
}

// create global instance
const i18nManager = new I18nManager();

// export functions for global access
function setLanguage(lang) {
    i18nManager.setLanguage(lang);
}

function getCurrentLanguage() {
    return i18nManager.getCurrentLanguage();
}

function getLanguages() {
    return i18nManager.getLanguages();
}

function t(key, params = {}) {
    return i18nManager.t(key, params);
}

function updateTranslations() {
    i18nManager.updateTranslations();
}

export { 
    setLanguage, 
    getCurrentLanguage, 
    getLanguages,
    t,
    updateTranslations,
    i18nManager
}; 
