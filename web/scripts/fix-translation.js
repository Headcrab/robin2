// фикс переводов - диагностика и решение проблем

// тестовая функция для проверки переводов
function testTranslations() {
    console.log('=== диагностика переводов ===');
    
    // проверяем существование i18nManager
    if (!window.i18nManager) {
        console.error('❌ window.i18nManager не найден');
        return;
    }
    
    console.log('✅ i18nManager найден');
    console.log('текущий язык:', window.i18nManager.currentLang);
    
    // проверяем переводы
    const testKeys = ['home.title', 'common.search', 'common.system'];
    testKeys.forEach(key => {
        const translation = window.i18nManager.t(key);
        console.log(`перевод "${key}":`, translation);
    });
    
    // проверяем элементы с data-i18n
    const i18nElements = document.querySelectorAll('[data-i18n]');
    console.log(`найдено ${i18nElements.length} элементов с data-i18n атрибутами`);
    
    i18nElements.forEach((element, index) => {
        const key = element.getAttribute('data-i18n');
        const translation = window.i18nManager.t(key);
        console.log(`элемент ${index + 1}: key="${key}", text="${element.textContent}", перевод="${translation}"`);
    });
    
    // проверяем работу updateTranslations
    console.log('вызываем updateTranslations...');
    window.i18nManager.updateTranslations();
    console.log('updateTranslations выполнен');
}

// функция для принудительного обновления переводов
function forceUpdateTranslations() {
    if (!window.i18nManager) {
        console.error('i18nManager не найден');
        return;
    }
    
    console.log('принудительное обновление переводов...');
    
    // обновляем все элементы с data-i18n
    const elements = document.querySelectorAll('[data-i18n]');
    elements.forEach(element => {
        const key = element.getAttribute('data-i18n');
        const translation = window.i18nManager.t(key);
        if (translation && translation !== key) {
            element.textContent = translation;
            console.log(`обновлен "${key}" -> "${translation}"`);
        }
    });
    
    // обновляем placeholders
    const placeholderElements = document.querySelectorAll('[data-i18n-placeholder]');
    placeholderElements.forEach(element => {
        const key = element.getAttribute('data-i18n-placeholder');
        const translation = window.i18nManager.t(key);
        if (translation && translation !== key) {
            element.setAttribute('placeholder', translation);
            console.log(`обновлен placeholder "${key}" -> "${translation}"`);
        }
    });
    
    // обновляем titles
    const titleElements = document.querySelectorAll('[data-i18n-title]');
    titleElements.forEach(element => {
        const key = element.getAttribute('data-i18n-title');
        const translation = window.i18nManager.t(key);
        if (translation && translation !== key) {
            element.setAttribute('title', translation);
            console.log(`обновлен title "${key}" -> "${translation}"`);
        }
    });
    
    console.log('принудительное обновление завершено');
}

// функция для тестирования смены языка
function testLanguageSwitch(lang) {
    if (!window.i18nManager) {
        console.error('i18nManager не найден');
        return;
    }
    
    console.log(`тестируем смену языка на: ${lang}`);
    
    console.log('до смены:', {
        currentLang: window.i18nManager.currentLang,
        title: document.querySelector('[data-i18n="home.title"]')?.textContent
    });
    
    window.i18nManager.setLanguage(lang);
    
    setTimeout(() => {
        console.log('после смены:', {
            currentLang: window.i18nManager.currentLang,
            title: document.querySelector('[data-i18n="home.title"]')?.textContent
        });
    }, 100);
}

// экспортируем функции в глобальную область
window.testTranslations = testTranslations;
window.forceUpdateTranslations = forceUpdateTranslations; 
window.testLanguageSwitch = testLanguageSwitch;

console.log('фикс переводов загружен. доступные функции:');
console.log('- testTranslations() - диагностика');
console.log('- forceUpdateTranslations() - принудительное обновление');
console.log('- testLanguageSwitch("kk") - тест смены языка'); 