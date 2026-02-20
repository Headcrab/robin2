import { showErrorNotification, showSuccessNotification } from './ui.js';

function downloadFile(filename, content, mimeType) {
    const blob = new Blob([content], { type: mimeType });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
}

function clearSearchForm() {
    const form = document.querySelector('form');
    if (form) {
        form.reset();
    }

    const searchInput = document.getElementById('searchInput');
    const dateFrom = document.getElementById('dateFrom');
    const dateTo = document.getElementById('dateTo');
    const searchCount = document.getElementById('searchCount');

    if (searchInput) searchInput.value = '';
    if (dateFrom) dateFrom.value = '';
    if (dateTo) dateTo.value = '';
    if (searchCount) searchCount.value = '300';

    const results = document.getElementById('data-results');
    if (results) {
        results.innerHTML = `
            <tr>
                <td colspan="6" class="text-center py-8 text-gray-500">
                    <div class="flex flex-col items-center space-y-3">
                        <svg class="h-12 w-12 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                        </svg>
                        <div>
                            <p class="text-lg font-medium text-gray-900">Введите параметры поиска</p>
                            <p class="text-gray-500">Заполните форму выше для поиска данных</p>
                        </div>
                    </div>
                </td>
            </tr>
        `;
    }
}

function exportData() {
    const table = document.querySelector('.data-table');
    if (!table) {
        showErrorNotification('Таблица данных не найдена');
        return;
    }

    const rows = table.querySelectorAll('tr');
    const csvRows = [];

    rows.forEach((row) => {
        const cols = row.querySelectorAll('th,td');
        if (!cols.length) return;
        if (cols.length === 1 && cols[0].getAttribute('colspan') === '6') return;
        const values = Array.from(cols).map((col) => {
            const text = col.textContent.trim();
            return `"${text.replace(/"/g, '""')}"`;
        });
        csvRows.push(values.join(','));
    });

    if (csvRows.length <= 1) {
        showErrorNotification('Нет данных для экспорта');
        return;
    }

    downloadFile(
        `data_export_${new Date().toISOString().slice(0, 10)}.csv`,
        csvRows.join('\n'),
        'text/csv;charset=utf-8'
    );
    showSuccessNotification('Данные экспортированы');
}

function clearTagSearch() {
    const searchInput = document.querySelector('#searchInput');
    if (searchInput) {
        searchInput.value = '';
        searchInput.focus();
    }
}

function exportTags() {
    const tagElements = document.querySelectorAll('.select-all');
    const tags = Array.from(tagElements).map(el => el.textContent.trim()).filter(Boolean);

    if (!tags.length) {
        showErrorNotification('Нет тегов для экспорта');
        return;
    }

    const csv = ['Tag', ...tags.map(tag => `"${tag.replace(/"/g, '""')}"`)].join('\n');
    downloadFile(
        `tags_export_${new Date().toISOString().slice(0, 10)}.csv`,
        csv,
        'text/csv;charset=utf-8'
    );
    showSuccessNotification('Список тегов экспортирован');
}

function exportLogs() {
    const logElements = document.querySelectorAll('.log-entry');
    const logs = Array.from(logElements).map(el => el.textContent.trim()).filter(Boolean);

    if (!logs.length) {
        showErrorNotification('Нет логов для экспорта');
        return;
    }

    downloadFile(
        `logs_export_${new Date().toISOString().slice(0, 10)}.txt`,
        logs.join('\n'),
        'text/plain;charset=utf-8'
    );
    showSuccessNotification('Логи экспортированы');
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
