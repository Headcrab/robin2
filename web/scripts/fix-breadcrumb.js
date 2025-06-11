// Отладка и проверка breadcrumb навигации

console.log('Fix breadcrumb script loaded');

// Проверка breadcrumb
function checkBreadcrumb() {
    console.log('\n=== BREADCRUMB CHECK ===');
    
    const breadcrumbContainer = document.getElementById('breadcrumb-container');
    const pageTitle = document.getElementById('page-title');
    
    console.log('Breadcrumb container found:', !!breadcrumbContainer);
    console.log('Page title found:', !!pageTitle);
    
    if (pageTitle) {
        console.log('Page title text:', pageTitle.textContent);
        console.log('Page title data-i18n:', pageTitle.getAttribute('data-i18n'));
    }
    
    if (breadcrumbContainer) {
        console.log('Breadcrumb content:', breadcrumbContainer.innerHTML);
        console.log('Breadcrumb text:', breadcrumbContainer.textContent);
    }
    
    console.log('Current URL:', window.location.pathname);
    console.log('=== END BREADCRUMB CHECK ===\n');
}

// Тест навигации
function testNavigation() {
    console.log('\n🧪 TESTING NAVIGATION...');
    
    const pages = [
        { url: '/', title: 'Главная' },
        { url: '/data/', title: 'Данные АСУТП' },
        { url: '/tags/', title: 'Управление тегами' },
        { url: '/logs/', title: 'Логи системы' },
        { url: '/swagger/', title: 'API документация' }
    ];
    
    pages.forEach(page => {
        console.log(`\n--- Testing ${page.url} ---`);
        
        // Симулируем updateBreadcrumb
        if (window.updateBreadcrumb || (window.navigation && window.navigation.updateBreadcrumb)) {
            const updateFunc = window.updateBreadcrumb || window.navigation.updateBreadcrumb;
            updateFunc(page.url);
            
            setTimeout(() => {
                console.log(`Page: ${page.url}`);
                console.log(`Expected title: ${page.title}`);
                
                const pageTitle = document.getElementById('page-title');
                const breadcrumb = document.getElementById('breadcrumb-container');
                
                if (pageTitle) {
                    console.log(`Actual title: ${pageTitle.textContent}`);
                }
                
                if (breadcrumb) {
                    const breadcrumbText = breadcrumb.textContent.trim();
                    console.log(`Breadcrumb: "${breadcrumbText}"`);
                    
                    // Проверяем что нет дублирования
                    if (breadcrumbText && breadcrumbText.includes(pageTitle.textContent)) {
                        console.warn('⚠️ DUPLICATION DETECTED in breadcrumb!');
                    } else {
                        console.log('✅ No duplication');
                    }
                }
            }, 200);
        }
    });
    
    console.log('🧪 NAVIGATION TEST COMPLETED');
}

// Исправление дублирования
function fixDuplication() {
    console.log('🛠️ FIXING BREADCRUMB DUPLICATION...');
    
    const breadcrumbContainer = document.getElementById('breadcrumb-container');
    const pageTitle = document.getElementById('page-title');
    
    if (!breadcrumbContainer || !pageTitle) {
        console.warn('Elements not found for fixing');
        return;
    }
    
    const currentUrl = window.location.pathname;
    
    // Принудительно вызываем updateBreadcrumb с исправленной логикой
    if (typeof updateBreadcrumb === 'function') {
        updateBreadcrumb(currentUrl);
        console.log('✅ Breadcrumb updated');
    } else {
        console.warn('updateBreadcrumb function not available');
    }
    
    // Проверяем результат
    setTimeout(checkBreadcrumb, 300);
}

// Проверка правильности отображения
function validateBreadcrumb() {
    console.log('\n🔍 VALIDATING BREADCRUMB...');
    
    const currentUrl = window.location.pathname;
    const breadcrumb = document.getElementById('breadcrumb-container');
    const pageTitle = document.getElementById('page-title');
    
    if (!breadcrumb || !pageTitle) return;
    
    const breadcrumbText = breadcrumb.textContent.trim();
    const titleText = pageTitle.textContent.trim();
    
    console.log(`URL: ${currentUrl}`);
    console.log(`Title: "${titleText}"`);
    console.log(`Breadcrumb: "${breadcrumbText}"`);
    
    const rules = {
        '/': { 
            breadcrumb: '', 
            title: 'Система получения данных АСУТП'
        },
        '/data/': { 
            breadcrumb: 'Главная', 
            title: 'Данные АСУТП' 
        },
        '/tags/': { 
            breadcrumb: 'Главная', 
            title: 'Управление тегами' 
        },
        '/logs/': { 
            breadcrumb: 'Главная', 
            title: 'Логи системы' 
        },
        '/swagger/': { 
            breadcrumb: 'Главная', 
            title: 'API документация' 
        }
    };
    
    const expected = rules[currentUrl];
    if (expected) {
        const breadcrumbMatch = breadcrumbText === expected.breadcrumb;
        const titleMatch = titleText.includes(expected.title.split(' ')[0]); // Частичное сравнение для переводов
        
        console.log(`Expected breadcrumb: "${expected.breadcrumb}" ${breadcrumbMatch ? '✅' : '❌'}`);
        console.log(`Expected title contains: "${expected.title.split(' ')[0]}" ${titleMatch ? '✅' : '❌'}`);
        
        if (breadcrumbMatch && titleMatch) {
            console.log('🎉 BREADCRUMB IS CORRECT!');
        } else {
            console.warn('⚠️ BREADCRUMB NEEDS FIXING');
        }
    }
    
    console.log('🔍 VALIDATION COMPLETED\n');
}

// Экспорт функций
window.checkBreadcrumb = checkBreadcrumb;
window.testNavigation = testNavigation;
window.fixDuplication = fixDuplication;
window.validateBreadcrumb = validateBreadcrumb;

// Автоматическая проверка
setTimeout(() => {
    checkBreadcrumb();
    validateBreadcrumb();
}, 1500); 