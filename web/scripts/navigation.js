// breadcrumb navigation
function updateBreadcrumb(url) {
    const breadcrumbContainer = document.getElementById('breadcrumb-container');
    const pageTitle = document.getElementById('page-title');
    
    if (!breadcrumbContainer || !pageTitle) return;
    
    let title = 'Система АСУТП';
    let breadcrumbs = [];
    
    switch(url) {
        case '/':
            title = 'Главная';
            breadcrumbs = ['Главная'];
            break;
        case '/data/':
            title = 'Данные';
            breadcrumbs = ['Главная', 'Данные'];
            break;
        case '/tags/':
            title = 'Теги';
            breadcrumbs = ['Главная', 'Теги'];
            break;
        case '/logs/':
            title = 'Логи';
            breadcrumbs = ['Главная', 'Логи'];
            break;
        case '/swagger/':
            title = 'API';
            breadcrumbs = ['Главная', 'API'];
            break;
        default:
            title = 'Система АСУТП';
            breadcrumbs = ['Главная'];
    }
    
    pageTitle.textContent = title;
    
    const breadcrumbHTML = breadcrumbs.map((crumb, index) => {
        if (index === breadcrumbs.length - 1) {
            return `<span class="text-gray-900 font-medium">${crumb}</span>`;
        } else {
            return `<span class="text-gray-500">${crumb}</span><span class="text-gray-300 mx-1">/</span>`;
        }
    }).join('');
    
    breadcrumbContainer.innerHTML = breadcrumbHTML;
}

export { updateBreadcrumb }; 