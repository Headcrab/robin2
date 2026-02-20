function fetchStatus() {
    const apiElement = document.getElementById('apiserver');
    if (!apiElement) {
        return;
    }

    const api = apiElement.textContent;

    fetch(`${api}/api/status/`)
        .then(response => response.json())
        .then(data => {
            updateSystemStatus(data);
            updateLastUpdateTime();
        })
        .catch(() => {
            setStatusError();
        });
}

// update all status indicators
function updateSystemStatus(data) {
    updateElementText('dbserver', data.dbserver);
    updateElementText('dbtype', data.dbtype);
    updateElementText('dbversion', data.dbversion);
    updateElementText('dbuptime', data.dbuptime);
    updateElementText('appuptime', data.appuptime);

    const isOk = data.dbstatus === 'green';
    const statusClass = isOk ? 'bg-green-500' : 'bg-red-500';
    const statusText = isOk ? 'Работает' : 'Ошибка';

    updateStatusIndicator('dbstatus', statusClass);
    updateStatusIndicator('mobile-dbstatus', statusClass);
    updateStatusIndicator('header-status', statusClass);
    updateStatusIndicator('home-status', statusClass);
    updateStatusIndicator('db-health-indicator', statusClass);

    updateElementText('dbstatus-text', statusText);
    updateElementText('mobile-dbstatus-text', statusText);
    updateElementText('db-health-text', statusText);
    updateElementText('system-status-text', statusText);
}

function updateStatusIndicator(elementId, statusClass) {
    const element = document.getElementById(elementId);
    if (element) {
        element.className = element.className.replace(/bg-(green|red|gray)-\d+/, statusClass);
    }
}

function updateElementText(elementId, text) {
    const element = document.getElementById(elementId);
    if (element && text !== undefined && text !== null) {
        element.textContent = text;
    }
}

function setStatusError() {
    const errorClass = 'bg-red-500';
    const errorText = 'Ошибка связи';

    updateStatusIndicator('dbstatus', errorClass);
    updateStatusIndicator('mobile-dbstatus', errorClass);
    updateStatusIndicator('header-status', errorClass);
    updateStatusIndicator('home-status', errorClass);
    updateStatusIndicator('db-health-indicator', errorClass);

    updateElementText('dbstatus-text', errorText);
    updateElementText('mobile-dbstatus-text', errorText);
    updateElementText('db-health-text', errorText);
    updateElementText('system-status-text', errorText);
}

function updateLastUpdateTime() {
    const now = new Date();
    updateElementText('last-update', now.toLocaleTimeString('ru-RU'));
}

function updateHomePageStats(data) {
    if (data) {
        const isOk = data.dbstatus === 'green';
        const statusClass = isOk ? 'bg-green-500' : 'bg-red-500';
        const statusText = isOk ? 'Работает' : 'Ошибка';
        updateStatusIndicator('home-status', statusClass);
        updateElementText('system-status-text', statusText);
        updateStatusIndicator('db-health-indicator', statusClass);
        updateElementText('db-health-text', statusText);
    }

    const apiElement = document.getElementById('apiserver');
    if (!apiElement) {
        return;
    }
    const api = apiElement.textContent;
    loadStatistics(api);
    loadRecentActivity(api);
}

function loadStatistics(api) {
    // Use existing API routes only.
    fetch(`${api}/get/tag/list/?like=%25&format=json`)
        .then(response => {
            if (!response.ok) throw new Error('tags unavailable');
            return response.json();
        })
        .then(tags => {
            let count = 0;
            if (Array.isArray(tags)) {
                count = tags.length;
            } else if (tags && Array.isArray(tags.rows)) {
                count = tags.rows.length;
            } else if (tags && Array.isArray(tags.tags)) {
                count = tags.tags.length;
            }
            updateElementText('active-tags-count', count > 0 ? count.toLocaleString('ru-RU') : 'Н/Д');
        })
        .catch(() => {
            updateElementText('active-tags-count', 'Н/Д');
        });

    fetch(`${api}/api/info/`)
        .then(response => {
            if (!response.ok) throw new Error('info unavailable');
            return response.json();
        })
        .then(info => {
            const opCount = Number(info.op_count || 0);
            updateElementText('data-records-count', opCount.toLocaleString('ru-RU'));
        })
        .catch(() => {
            updateElementText('data-records-count', 'Н/Д');
        });
}

function loadRecentActivity(api) {
    const recentActivityContainer = document.getElementById('recent-activity');
    if (!recentActivityContainer) {
        return;
    }

    fetch(`${api}/api/log/?format=json`)
        .then(response => {
            if (!response.ok) throw new Error('logs unavailable');
            return response.json();
        })
        .then(logs => {
            if (!Array.isArray(logs) || logs.length === 0) {
                showNoActivityMessage();
                return;
            }

            const recent = logs.slice(-5).reverse();
            const activityHTML = recent.map((log) => {
                const level = readLogField(log, ['level', 'Level']) || 'INF';
                const levelClass = getLevelClass(level);
                const timestamp = readLogField(log, ['date', 'Date']) || '';
                const time = formatTime(timestamp);
                const message = readLogField(log, ['msg', 'Msg', 'message']) || 'Системное событие';

                return `
                    <div class="flex items-center space-x-3 text-sm">
                        <div class="h-2 w-2 rounded-full ${levelClass}"></div>
                        <span class="text-gray-500 font-mono text-xs">${time}</span>
                        <span class="text-gray-900 flex-1">${escapeHtml(message).slice(0, 80)}</span>
                    </div>
                `;
            }).join('');

            recentActivityContainer.innerHTML = activityHTML;
        })
        .catch(() => {
            showNoActivityMessage();
        });
}

function readLogField(log, keys) {
    for (const key of keys) {
        if (log && log[key] !== undefined && log[key] !== null) {
            return String(log[key]);
        }
    }
    return '';
}

function getLevelClass(level) {
    switch (String(level).toUpperCase()) {
        case 'ERR':
        case 'ERROR':
        case 'FATAL':
            return 'bg-red-500';
        case 'WRN':
        case 'WARN':
        case 'WARNING':
            return 'bg-yellow-500';
        case 'DBG':
        case 'DEBUG':
            return 'bg-blue-500';
        default:
            return 'bg-green-500';
    }
}

function formatTime(timestamp) {
    if (!timestamp) return '--:--';
    const date = new Date(timestamp);
    if (Number.isNaN(date.getTime())) return '--:--';
    return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
}

function showNoActivityMessage() {
    const recentActivityContainer = document.getElementById('recent-activity');
    if (recentActivityContainer) {
        recentActivityContainer.innerHTML = `
            <div class="flex items-center space-x-3 text-sm text-gray-500">
                <div class="h-2 w-2 bg-gray-300 rounded-full"></div>
                <span>Нет недавних событий</span>
            </div>
        `;
    }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

export {
    fetchStatus,
    updateSystemStatus,
    updateHomePageStats,
    loadStatistics,
    loadRecentActivity
};
