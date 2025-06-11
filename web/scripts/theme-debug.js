// theme and translation debug utilities

function debugThemeAndTranslations() {
    console.log('=== THEME AND TRANSLATIONS DEBUG ===');
    
    // check if managers are loaded
    console.log('ThemeManager available:', !!window.themeManager);
    console.log('I18nManager available:', !!window.i18nManager);
    
    // check functions
    console.log('Functions available:');
    console.log('- toggleTheme:', typeof window.toggleTheme);
    console.log('- setTheme:', typeof window.setTheme);
    console.log('- getCurrentTheme:', typeof window.getCurrentTheme);
    console.log('- setLanguage:', typeof window.setLanguage);
    console.log('- getCurrentLanguage:', typeof window.getCurrentLanguage);
    console.log('- t (translate):', typeof window.t);
    console.log('- updateTranslations:', typeof window.updateTranslations);
    
    // current state
    if (window.getCurrentTheme) {
        console.log('Current theme:', window.getCurrentTheme());
    }
    if (window.getCurrentLanguage) {
        console.log('Current language:', window.getCurrentLanguage());
    }
    
    // check DOM elements
    console.log('DOM elements:');
    console.log('- theme-toggle:', !!document.getElementById('theme-toggle'));
    console.log('- language-toggle:', !!document.getElementById('language-toggle'));
    console.log('- language-dropdown:', !!document.getElementById('language-dropdown'));
    console.log('- theme-language-switchers:', !!document.getElementById('theme-language-switchers'));
    
    // check data attributes
    console.log('Document theme attribute:', document.documentElement.getAttribute('data-theme'));
    console.log('Document lang attribute:', document.documentElement.getAttribute('lang'));
    
    // check CSS variables
    const rootStyles = getComputedStyle(document.documentElement);
    console.log('CSS variables:');
    console.log('--bg-primary:', rootStyles.getPropertyValue('--bg-primary'));
    console.log('--text-primary:', rootStyles.getPropertyValue('--text-primary'));
    
    // check translation elements
    const i18nElements = document.querySelectorAll('[data-i18n]');
    console.log(`Found ${i18nElements.length} translatable elements`);
    
    // test translation
    if (window.t) {
        console.log('Translation test:');
        console.log('- nav.home:', window.t('nav.home'));
        console.log('- common.loading:', window.t('common.loading'));
        console.log('- theme.toggle:', window.t('theme.toggle'));
    }
    
    console.log('=== END DEBUG ===');
}

// test theme switching
function testThemeSwitching() {
    console.log('Testing theme switching...');
    if (window.toggleTheme) {
        window.toggleTheme();
        setTimeout(() => {
            console.log('Theme after toggle:', window.getCurrentTheme());
            window.toggleTheme(); // switch back
        }, 1000);
    }
}

// test language switching
function testLanguageSwitching() {
    console.log('Testing language switching...');
    if (window.setLanguage) {
        const originalLang = window.getCurrentLanguage();
        console.log('Original language:', originalLang);
        
        window.setLanguage('en');
        setTimeout(() => {
            console.log('Language after change:', window.getCurrentLanguage());
            console.log('Translation test (en):', window.t('nav.home'));
            
            // switch back
            window.setLanguage(originalLang);
            setTimeout(() => {
                console.log('Language after restore:', window.getCurrentLanguage());
                console.log('Translation test (restored):', window.t('nav.home'));
            }, 500);
        }, 500);
    }
}

// force update translations
function forceUpdateTranslations() {
    console.log('Force updating translations...');
    if (window.updateTranslations) {
        window.updateTranslations();
        console.log('Translations updated');
    }
}

// check for translation issues
function checkTranslationIssues() {
    console.log('Checking for translation issues...');
    
    const i18nElements = document.querySelectorAll('[data-i18n]');
    const issues = [];
    
    i18nElements.forEach((element, index) => {
        const key = element.getAttribute('data-i18n');
        const translation = window.t ? window.t(key) : null;
        
        if (!translation || translation === key) {
            issues.push({
                element: element,
                key: key,
                text: element.textContent,
                index: index
            });
        }
    });
    
    if (issues.length > 0) {
        console.warn(`Found ${issues.length} translation issues:`);
        issues.forEach(issue => {
            console.warn(`- Element ${issue.index}: key "${issue.key}" not translated, showing "${issue.text}"`);
        });
    } else {
        console.log('No translation issues found');
    }
    
    return issues;
}

// manual initialization if auto-init failed
function manualInit() {
    console.log('Manual initialization...');
    
    // try to initialize theme and language system
    if (window.initializeThemeAndLanguage) {
        window.initializeThemeAndLanguage();
    }
    
    // force translations update
    setTimeout(() => {
        if (window.updateTranslations) {
            window.updateTranslations();
        }
    }, 500);
}

// test switcher restoration after navigation
function testSwitcherRestoration() {
    console.log('Testing switcher restoration after simulated navigation...');
    
    // simulate switcher removal
    const switchers = document.getElementById('theme-language-switchers');
    if (switchers) {
        switchers.remove();
        console.log('Switchers removed, waiting for restoration...');
        
        setTimeout(() => {
            const restored = document.getElementById('theme-language-switchers');
            if (restored) {
                console.log('✅ Switchers automatically restored!');
            } else {
                console.log('❌ Switchers NOT restored automatically');
                console.log('Trying manual restoration...');
                if (window.restoreSwitchers) {
                    window.restoreSwitchers();
                }
            }
        }, 1000);
    } else {
        console.log('No switchers found to test with');
    }
}

// export debug functions to window
window.debugThemeAndTranslations = debugThemeAndTranslations;
window.testThemeSwitching = testThemeSwitching;
window.testLanguageSwitching = testLanguageSwitching;
window.forceUpdateTranslations = forceUpdateTranslations;
window.checkTranslationIssues = checkTranslationIssues;
window.manualInit = manualInit;
window.testSwitcherRestoration = testSwitcherRestoration;

// auto-run debug when script loads
if (window.console) {
    setTimeout(debugThemeAndTranslations, 1000);
}

console.log('Theme debug utilities loaded. Available functions:');
console.log('- debugThemeAndTranslations()');
console.log('- testThemeSwitching()');  
console.log('- testLanguageSwitching()');
console.log('- forceUpdateTranslations()');
console.log('- checkTranslationIssues()');
console.log('- manualInit()');
console.log('- testSwitcherRestoration()');
console.log('- restoreSwitchers()'); 