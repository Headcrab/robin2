import { showErrorNotification, showSuccessNotification } from './ui.js';

function clearSearchForm() {
    const form = document.querySelector('form');
    if (form) {
        form.reset();
        
        // clear specific fields that might not be in the form
        const tagSelect = document.getElementById('tag');
        const startDate = document.getElementById('start');
        const endDate = document.getElementById('end');
        
        if (tagSelect) tagSelect.value = '';
        if (startDate) startDate.value = '';
        if (endDate) endDate.value = '';
        
        // trigger search with cleared parameters
        if (typeof window.getTagOnDate === 'function') {
            window.getTagOnDate();
        }
    }
}

function exportData() {
    const tagInput = document.getElementById('searchInput');
    const currentTag = tagInput ? tagInput.value || 'Unknown' : 'Unknown';
    const startDate = document.getElementById('start')?.value;
    const endDate = document.getElementById('end')?.value;
    const api = document.getElementById('apiserver')?.textContent;
    
    if (!api) {
        showErrorNotification('API сервер не настроен');
        return;
    }
    
    let exportUrl = `${api}/api/export/data`;
    const params = new URLSearchParams();
    
    if (currentTag) params.append('tag', currentTag);
    if (startDate) params.append('start', startDate);
    if (endDate) params.append('end', endDate);
    
    if (params.toString()) {
        exportUrl += '?' + params.toString();
    }
    
    window.open(exportUrl, '_blank');
}

function clearTagSearch() {
    const searchInput = document.querySelector('#tag-search');
    if (searchInput) {
        searchInput.value = '';
        searchInput.dispatchEvent(new Event('input'));
    }
}

function exportTags() {
    const api = document.getElementById('apiserver')?.textContent;
    
    if (!api) {
        showErrorNotification('API сервер не настроен');
        return;
    }
    
    const exportUrl = `${api}/api/export/tags`;
    window.open(exportUrl, '_blank');
}

function exportLogs() {
    const api = document.getElementById('apiserver')?.textContent;
    
    if (!api) {
        showErrorNotification('API сервер не настроен');
        return;
    }
    
    const exportUrl = `${api}/api/export/logs`;
    window.open(exportUrl, '_blank');
}

function clearLogs() {
    const api = document.getElementById('apiserver')?.textContent;
    
    if (!api) {
        showErrorNotification('API сервер не настроен');
        return;
    }
    
    // подтверждение действия
    if (!confirm('Вы уверены, что хотите очистить все логи? Это действие необратимо.')) {
        return;
    }
    
    const clearUrl = `${api}/api/log/clear/`;
    
    fetch(clearUrl, {
        method: 'POST'
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`Ошибка HTTP: ${response.status}`);
        }
        return response.text();
    })
    .then(data => {
        showSuccessNotification('Логи успешно очищены');
        // перезагружаем страницу с параметром refresh чтобы очистить кеш
        setTimeout(() => {
            if (typeof window.loadPage === 'function') {
                window.loadPage('/logs/?refresh=1');
            }
        }, 1000);
    })
    .catch(error => {
        console.error('Ошибка при очистке логов:', error);
        showErrorNotification('Ошибка при очистке логов: ' + error.message);
    });
}

export { 
    clearSearchForm,
    exportData,
    clearTagSearch,
    exportTags,
    exportLogs,
    clearLogs
}; 