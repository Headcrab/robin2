// breadcrumb navigation
function updateBreadcrumb(url) {
    const breadcrumbContainer = document.getElementById('breadcrumb-container');
    const pageTitle = document.getElementById('page-title');
    
    if (!breadcrumbContainer || !pageTitle) return;

    const path = (url || window.location.pathname || '/').split('?')[0];
    
    let titleKey = 'home.title';
    let breadcrumbKeys = [];
    
    switch(path) {
        case '/':
        case '':
            titleKey = 'home.title';
            breadcrumbKeys = []; // Главная страница - нет breadcrumb
            break;
        case '/data/':
        case '/data':
            titleKey = 'data.title';
            breadcrumbKeys = ['nav.home']; // Главная > текущая страница в заголовке
            break;
        case '/charts/':
        case '/charts':
            titleKey = 'charts.title';
            breadcrumbKeys = ['nav.home'];
            break;
        case '/tags/':
        case '/tags':
            titleKey = 'tags.title';
            breadcrumbKeys = ['nav.home'];
            break;
        case '/logs/':
        case '/logs':
            titleKey = 'logs.title';
            breadcrumbKeys = ['nav.home'];
            break;
        case '/docs/':
        case '/docs':
            titleKey = 'docs.title';
            breadcrumbKeys = ['nav.home'];
            break;
        case '/docs/view/':
        case '/docs/view':
            titleKey = 'docs.title';
            breadcrumbKeys = ['nav.home', 'nav.docs'];
            break;
        case '/swagger/':
        case '/swagger':
            titleKey = 'api.title';
            breadcrumbKeys = ['nav.home'];
            break;
        default:
            titleKey = 'home.title';
            breadcrumbKeys = [];
    }
    
    // set title using translation key
    pageTitle.setAttribute('data-i18n', titleKey);
    if (window.i18nManager) {
        pageTitle.textContent = window.i18nManager.t(titleKey);
    }
    
    // update breadcrumbs - показываем только путь, без текущей страницы
    if (breadcrumbKeys.length === 0) {
        breadcrumbContainer.innerHTML = '';
        return;
    }
    
    const breadcrumbHTML = breadcrumbKeys.map((key, index) => {
        const translated = window.i18nManager ? window.i18nManager.t(key) : key;
        const isLast = index === breadcrumbKeys.length - 1;
        
        if (isLast) {
            return `<span class="text-gray-500">${translated}</span>`;
        } else {
            return `<span class="text-gray-500">${translated}</span><span class="text-gray-300 mx-1">/</span>`;
        }
    }).join('');
    
    breadcrumbContainer.innerHTML = breadcrumbHTML;
}

export { updateBreadcrumb }; 
