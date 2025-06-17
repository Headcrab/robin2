// Исправление проблем с переводом навигации и языком

console.log('Fix navigation translation script loaded');

// проверка инициализации языка на странице
function checkLanguageOnPage() {
    const currentLang = localStorage.getItem('selectedLanguage') || 'ru';
    console.log('Current language from localStorage:', currentLang);
    
    if (window.i18nManager) {
        console.log('i18nManager available, current language:', window.i18nManager.currentLanguage);
        window.i18nManager.updateTranslations();
    } else {
        console.warn('i18nManager not available');
    }
}

// принудительная инициализация языка
function forceLanguageInit() {
    const savedLang = localStorage.getItem('selectedLanguage') || 'ru';
    console.log('Force init language:', savedLang);
    
    if (window.i18nManager && window.i18nManager.setLanguage) {
        window.i18nManager.setLanguage(savedLang);
        console.log('Language forced to:', savedLang);
    }
}

// проверка переводов в навигации
function checkNavigationTranslations() {
    const navItems = [
        { selector: '[data-i18n="nav.home"]', expected: ['Главная', 'Басты бет', 'Home'] },
        { selector: '[data-i18n="nav.data"]', expected: ['Данные', 'Деректер', 'Data'] },
        { selector: '[data-i18n="nav.tags"]', expected: ['Теги', 'Тегтер', 'Tags'] },
        { selector: '[data-i18n="nav.logs"]', expected: ['Логи', 'Логтер', 'Logs'] },
        { selector: '[data-i18n="nav.docs"]', expected: ['Документация', 'Құжаттама', 'Documentation'] },
        { selector: '[data-i18n="nav.api"]', expected: ['API', 'API', 'API'] }
    ];
    
    console.log('Checking navigation translations:');
    
    navItems.forEach(item => {
        const elements = document.querySelectorAll(item.selector);
        elements.forEach(el => {
            const text = el.textContent.trim();
            console.log(`${item.selector}: "${text}"`);
            
            if (!item.expected.includes(text)) {
                console.warn(`❌ Wrong translation for ${item.selector}: "${text}" (expected one of: ${item.expected.join(', ')})`);
            } else {
                console.log(`✅ Correct translation for ${item.selector}: "${text}"`);
            }
        });
    });
}

// проверка работы переключателя языка
function testLanguageSwitching() {
    console.log('Testing language switching...');
    
    const languages = ['ru', 'kk', 'en'];
    let index = 0;
    
    const testNext = () => {
        if (index >= languages.length) {
            console.log('Language switching test completed');
            return;
        }
        
        const lang = languages[index];
        console.log(`Switching to ${lang}...`);
        
        if (window.i18nManager) {
            window.i18nManager.setLanguage(lang);
            setTimeout(() => {
                checkNavigationTranslations();
                index++;
                setTimeout(testNext, 1000);
            }, 500);
        } else {
            console.error('i18nManager not available for testing');
        }
    };
    
    testNext();
}

// автоматическая диагностика
function diagnosePage() {
    console.log('\n=== PAGE LANGUAGE DIAGNOSTICS ===');
    
    // 1. Проверить localStorage
    const savedLang = localStorage.getItem('selectedLanguage');
    console.log('1. SavedLanguage in localStorage:', savedLang);
    
    // 2. Проверить глобальные менеджеры
    console.log('2. Available managers:');
    console.log('   - window.i18nManager:', !!window.i18nManager);
    console.log('   - window.themeManager:', !!window.themeManager);
    
    // 3. Проверить текущий язык
    if (window.i18nManager) {
        console.log('3. Current language:', window.i18nManager.currentLanguage);
    }
    
    // 4. Проверить переводы в DOM
    console.log('4. Checking DOM translations...');
    const elementsWithI18n = document.querySelectorAll('[data-i18n]');
    console.log(`   Found ${elementsWithI18n.length} elements with data-i18n attributes`);
    
    // 5. Проверить навигацию
    console.log('5. Navigation check:');
    checkNavigationTranslations();
    
    console.log('=== END DIAGNOSTICS ===\n');
}

// восстановление после загрузки страницы
function restoreLanguageAfterPageLoad() {
    const savedLang = localStorage.getItem('selectedLanguage') || 'ru';
    console.log('Restoring language after page load:', savedLang);
    
    // ждем полной инициализации
    setTimeout(() => {
        if (window.i18nManager) {
            window.i18nManager.setLanguage(savedLang);
            console.log('Language restored to:', savedLang);
        } else {
            console.warn('Cannot restore language - i18nManager not available');
        }
    }, 500);
}

// экспорт функций для консоли
window.checkLanguageOnPage = checkLanguageOnPage;
window.forceLanguageInit = forceLanguageInit;
window.checkNavigationTranslations = checkNavigationTranslations;
window.testLanguageSwitching = testLanguageSwitching;
window.diagnosePage = diagnosePage;
window.restoreLanguageAfterPageLoad = restoreLanguageAfterPageLoad;

// автоматический вызов диагностики
setTimeout(diagnosePage, 1000); 