// Исправление проблем с темой таблицы

console.log('Fix table theme script loaded');

// Проверка таблицы данных
function checkDataTableTheme() {
    console.log('\n=== DATA TABLE THEME CHECK ===');
    
    const table = document.querySelector('.data-table');
    if (!table) {
        console.warn('❌ .data-table not found');
        return;
    }
    
    console.log('✅ Table found:', table);
    
    // Проверка темы
    const currentTheme = document.documentElement.getAttribute('data-theme') || 'light';
    console.log('Current theme:', currentTheme);
    
    // Проверка заголовков
    const headers = table.querySelectorAll('thead th');
    console.log(`Found ${headers.length} headers:`);
    
    headers.forEach((th, index) => {
        const text = th.textContent.trim();
        const styles = window.getComputedStyle(th);
        console.log(`Header ${index + 1}: "${text}"`);
        console.log(`  - Background: ${styles.backgroundColor}`);
        console.log(`  - Color: ${styles.color}`);
        console.log(`  - Font-weight: ${styles.fontWeight}`);
        console.log(`  - Text-transform: ${styles.textTransform}`);
    });
    
    // Проверка строк таблицы
    const rows = table.querySelectorAll('tbody tr');
    console.log(`Found ${rows.length} table rows`);
    
    if (rows.length > 0) {
        const firstRow = rows[0];
        const firstRowStyles = window.getComputedStyle(firstRow);
        console.log('First row styles:');
        console.log(`  - Background: ${firstRowStyles.backgroundColor}`);
        console.log(`  - Color: ${firstRowStyles.color}`);
    }
    
    console.log('=== END TABLE CHECK ===\n');
}

// Принудительное применение темы к таблице
function forceTableTheme() {
    const currentTheme = document.documentElement.getAttribute('data-theme') || 'light';
    console.log('Forcing table theme:', currentTheme);
    
    const table = document.querySelector('.data-table');
    if (!table) {
        console.warn('Table not found for theme forcing');
        return;
    }
    
    // Убираем и заново добавляем классы
    table.classList.remove('data-table');
    setTimeout(() => {
        table.classList.add('data-table');
        console.log('Table theme forced');
        checkDataTableTheme();
    }, 100);
}

// Проверка переводов заголовков
function checkTableTranslations() {
    console.log('\n=== TABLE TRANSLATIONS CHECK ===');
    
    const expectedTranslations = {
        'ru': ['ВРЕМЯ', 'ТЕГ', 'ЗНАЧЕНИЕ', 'КАЧЕСТВО', 'ЕДИНИЦА', 'ОПИСАНИЕ'],
        'kk': ['УАҚЫТ', 'ТЕГ', 'МӘН', 'САПА', 'ӨЛШЕМ БІРЛІГІ', 'СИПАТТАМА'],
        'en': ['TIME', 'TAG', 'VALUE', 'QUALITY', 'UNIT', 'DESCRIPTION']
    };
    
    const currentLang = localStorage.getItem('selectedLanguage') || 'ru';
    const expected = expectedTranslations[currentLang];
    
    const headers = document.querySelectorAll('.data-table thead th');
    
    console.log(`Current language: ${currentLang}`);
    console.log(`Expected headers: ${expected.join(', ')}`);
    console.log('Actual headers:');
    
    headers.forEach((th, index) => {
        const text = th.textContent.trim().toUpperCase();
        const expectedText = expected[index];
        const isCorrect = text === expectedText;
        
        console.log(`${index + 1}. "${text}" ${isCorrect ? '✅' : '❌'} (expected: "${expectedText}")`);
    });
    
    console.log('=== END TRANSLATIONS CHECK ===\n');
}

// Исправление стилей заголовков
function fixHeaderStyles() {
    console.log('Fixing header styles...');
    
    const headers = document.querySelectorAll('.data-table thead th');
    
    headers.forEach((th, index) => {
        // Принудительное применение стилей
        th.style.cssText = `
            color: white !important;
            background: transparent !important;
            font-weight: 700 !important;
            font-size: 0.8rem !important;
            text-transform: uppercase !important;
            letter-spacing: 0.08em !important;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif !important;
        `;
    });
    
    console.log('Header styles fixed');
}

// Полная диагностика таблицы
function diagnoseTable() {
    console.log('\n🔍 FULL TABLE DIAGNOSTICS 🔍');
    
    checkDataTableTheme();
    checkTableTranslations();
    
    // Проверка CSS файлов
    const stylesheets = document.querySelectorAll('link[rel="stylesheet"]');
    console.log('Loaded stylesheets:');
    stylesheets.forEach((link, index) => {
        console.log(`${index + 1}. ${link.href}`);
    });
    
    // Проверка theme manager
    console.log('Theme manager available:', !!window.themeManager);
    console.log('I18n manager available:', !!window.i18nManager);
    
    console.log('🔍 END DIAGNOSTICS 🔍\n');
}

// Автоматическое исправление таблицы
function autoFixTable() {
    console.log('🛠️ AUTO-FIXING TABLE...');
    
    // 1. Исправить стили заголовков
    fixHeaderStyles();
    
    // 2. Принудительно применить тему
    forceTableTheme();
    
    // 3. Обновить переводы
    if (window.i18nManager) {
        window.i18nManager.updateTranslations();
    }
    
    // 4. Повторная проверка через секунду
    setTimeout(diagnoseTable, 1000);
    
    console.log('🛠️ AUTO-FIX COMPLETED');
}

// Экспорт функций
window.checkDataTableTheme = checkDataTableTheme;
window.forceTableTheme = forceTableTheme;
window.checkTableTranslations = checkTableTranslations;
window.fixHeaderStyles = fixHeaderStyles;
window.diagnoseTable = diagnoseTable;
window.autoFixTable = autoFixTable;

// Автоматическая диагностика через 2 секунды
setTimeout(diagnoseTable, 2000); 