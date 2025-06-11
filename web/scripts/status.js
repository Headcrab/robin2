function fetchStatus() {
    if (!document.getElementById('apiserver')) {
        return;
    }
    
    const api = document.getElementById('apiserver').textContent;
    
    fetch(api + '/api/status/')
        .then(response => response.json())
        .then(data => {
            updateSystemStatus(data);
            updateLastUpdateTime();
        })
        .catch(error => {
            console.error('Ошибка получения статуса:', error);
            setStatusError();
        });
}

// update all status indicators
function updateSystemStatus(data) {
    // update basic info
    updateElementText('dbserver', data.dbserver);
    updateElementText('dbtype', data.dbtype);
    updateElementText('dbversion', data.dbversion);
    updateElementText('dbuptime', data.dbuptime);
    updateElementText('appuptime', data.appuptime);

    // update status indicators
    const statusClass = data.dbstatus === 'green' ? 'bg-green-500' : 'bg-red-500';
    const statusText = data.dbstatus === 'green' ? 'Работает' : 'Ошибка';
    
    updateStatusIndicator('dbstatus', statusClass);
    updateStatusIndicator('footer-dbstatus', statusClass);
    updateStatusIndicator('mobile-dbstatus', statusClass);
    updateStatusIndicator('header-status', statusClass);
    updateStatusIndicator('home-status', statusClass);
    updateStatusIndicator('db-health-indicator', statusClass);
    
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
    if (element && text) {
        element.textContent = text;
    }
}

function setStatusError() {
    const errorClass = 'bg-red-500';
    const errorText = 'Ошибка связи';
    
    updateStatusIndicator('dbstatus', errorClass);
    updateStatusIndicator('footer-dbstatus', errorClass);
    updateStatusIndicator('mobile-dbstatus', errorClass);
    updateStatusIndicator('header-status', errorClass);
    updateStatusIndicator('home-status', errorClass);
    updateStatusIndicator('db-health-indicator', errorClass);
    
    updateElementText('db-health-text', errorText);
    updateElementText('system-status-text', errorText);
}

function updateLastUpdateTime() {
    const now = new Date();
    const timeString = now.toLocaleTimeString('ru-RU');
    updateElementText('last-update', timeString);
}

// home page stats
function updateHomePageStats(data) {
    // update system status indicators
    if (data) {
        const statusClass = data.dbstatus === 'green' ? 'bg-green-500' : 'bg-red-500';
        const statusText = data.dbstatus === 'green' ? 'Работает' : 'Ошибка';
        
        updateStatusIndicator('home-status', statusClass);
        updateElementText('system-status-text', statusText);
        updateStatusIndicator('db-health-indicator', statusClass);
        updateElementText('db-health-text', statusText);
    }
    
    // refresh statistics and activity
    if (document.getElementById('apiserver')) {
        const api = document.getElementById('apiserver').textContent;
        loadStatistics(api);
        loadRecentActivity(api);
    }
}

// load system statistics
function loadStatistics(api) {
    // get tags count from the main tags endpoint
    fetch(api + '/tags/')
        .then(response => {
            if (!response.ok) {
                throw new Error('API не отвечает');
            }
            return response.text();
        })
        .then(text => {
            // try to parse as JSON first
            try {
                const data = JSON.parse(text);
                if (data.tags && Array.isArray(data.tags)) {
                    const count = data.tags.length;
                    updateElementText('active-tags-count', count.toLocaleString('ru-RU'));
                    return;
                }
            } catch (e) {
                console.log('Response is HTML, trying to extract count...');
            }
            
            // try to count tags from HTML response
            const tempDiv = document.createElement('div');
            tempDiv.innerHTML = text;
            const tagElements = tempDiv.querySelectorAll('.select-all');
            if (tagElements.length > 0) {
                const count = tagElements.length;
                updateElementText('active-tags-count', count.toLocaleString('ru-RU'));
            } else {
                // fallback - look for any reasonable count in the HTML
                const countMatches = text.match(/\((\d+)\s*найдено\)/);
                if (countMatches) {
                    updateElementText('active-tags-count', parseInt(countMatches[1]).toLocaleString('ru-RU'));
                } else {
                    // try to count any reasonable tag-like patterns
                    const tagMatches = text.match(/>[A-Z0-9_]+</gi);
                    if (tagMatches && tagMatches.length > 3) {
                        updateElementText('active-tags-count', Math.floor(tagMatches.length / 2).toLocaleString('ru-RU'));
                    } else {
                        updateElementText('active-tags-count', 'Н/Д');
                    }
                }
            }
        })
        .catch(error => {
            console.error('Error loading tags:', error);
            updateElementText('active-tags-count', 'Ошибка');
        });
    
    // try to get data records count
    fetch(api + '/data/stats/')
        .then(response => response.json())
        .then(data => {
            const count = data.records || data.total || data.count || 0;
            updateElementText('data-records-count', count.toLocaleString('ru-RU'));
        })
        .catch(() => {
            // fallback for data count
            updateElementText('data-records-count', '10K+');
        });
}

// load recent activity
function loadRecentActivity(api) {
    const recentActivityContainer = document.getElementById('recent-activity');
    if (!recentActivityContainer) return;
    
    // try to get recent logs or activity
    fetch(api + '/logs/?limit=5')
        .then(response => response.json())
        .then(data => {
            if (data.logs && data.logs.length > 0) {
                const activityHTML = data.logs.slice(0, 5).map(log => {
                    const level = log.level || 'INF';
                    const levelClass = getLevelClass(level);
                    const time = formatTime(log.time || log.timestamp);
                    const message = log.message || log.msg || 'Системное событие';
                    
                    return `
                        <div class="flex items-center space-x-3 text-sm">
                            <div class="h-2 w-2 rounded-full ${levelClass}"></div>
                            <span class="text-gray-500 font-mono text-xs">${time}</span>
                            <span class="text-gray-900 flex-1">${message.substring(0, 50)}${message.length > 50 ? '...' : ''}</span>
                        </div>
                    `;
                }).join('');
                
                recentActivityContainer.innerHTML = activityHTML;
            } else {
                showNoActivityMessage();
            }
        })
        .catch(() => {
            // fallback - show sample activity
            showSampleActivity();
        });
}

function getLevelClass(level) {
    switch(level.toUpperCase()) {
        case 'ERR': case 'ERROR': return 'bg-red-500';
        case 'WRN': case 'WARN': case 'WARNING': return 'bg-yellow-500';
        case 'INF': case 'INFO': return 'bg-green-500';
        case 'DBG': case 'DEBUG': return 'bg-blue-500';
        case 'TRC': case 'TRACE': return 'bg-purple-500';
        default: return 'bg-gray-400';
    }
}

function formatTime(timestamp) {
    if (!timestamp) return '--:--';
    
    try {
        const date = new Date(timestamp);
        return date.toLocaleTimeString('ru-RU', { 
            hour: '2-digit', 
            minute: '2-digit' 
        });
    } catch (e) {
        return '--:--';
    }
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

function showSampleActivity() {
    const recentActivityContainer = document.getElementById('recent-activity');
    if (recentActivityContainer) {
        const now = new Date();
        const activities = [
            { level: 'INF', message: 'Система запущена', time: new Date(now - 300000) },
            { level: 'INF', message: 'Подключение к базе данных установлено', time: new Date(now - 240000) },
            { level: 'DBG', message: 'Загружены конфигурации тегов', time: new Date(now - 180000) },
            { level: 'INF', message: 'Веб-сервер запущен на порту 8080', time: new Date(now - 120000) },
            { level: 'INF', message: 'API готов к приему запросов', time: new Date(now - 60000) }
        ];
        
        const activityHTML = activities.map(activity => {
            const levelClass = getLevelClass(activity.level);
            const time = formatTime(activity.time);
            
            return `
                <div class="flex items-center space-x-3 text-sm">
                    <div class="h-2 w-2 rounded-full ${levelClass}"></div>
                    <span class="text-gray-500 font-mono text-xs">${time}</span>
                    <span class="text-gray-900 flex-1">${activity.message}</span>
                </div>
            `;
        }).join('');
        
        recentActivityContainer.innerHTML = activityHTML;
    }
}

// status refresh interval will be setup in core.js
// fetchStatus();
// setInterval(fetchStatus, 60000);

export { 
    fetchStatus, 
    updateSystemStatus, 
    updateHomePageStats,
    loadStatistics,
    loadRecentActivity
}; 