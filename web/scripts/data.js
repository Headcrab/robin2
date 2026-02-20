import { showErrorNotification, showSuccessNotification } from './ui.js';

function getTagOnDate() {
    const searchBtn = document.getElementById('searchBtn');
    const originalText = searchBtn ? searchBtn.innerHTML : '';
    if (searchBtn) {
        searchBtn.disabled = true;
        searchBtn.innerHTML = `
            <svg class="animate-spin h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Поиск...
        `;
    }
    
    const tag = document.getElementById("searchInput")?.value || '';
    const dateFrom = document.getElementById("dateFrom")?.value || '';
    const dateTo = document.getElementById("dateTo")?.value || '';
    const searchCount = document.getElementById("searchCount")?.value || '300';
    
    if (!tag.trim()) {
        showErrorNotification('Введите название тега');
        if (searchBtn) {
            searchBtn.disabled = false;
            searchBtn.innerHTML = originalText;
        }
        return;
    }
    
    if (!dateFrom || !dateTo) {
        showErrorNotification('Укажите период поиска');
        if (searchBtn) {
            searchBtn.disabled = false;
            searchBtn.innerHTML = originalText;
        }
        return;
    }
    
    const apiElement = document.getElementById('apiserver');
    if (!apiElement) {
        showErrorNotification('API сервер недоступен');
        if (searchBtn) {
            searchBtn.disabled = false;
            searchBtn.innerHTML = originalText;
        }
        return;
    }
    
    const api = apiElement.textContent;
    
    const fromFormatted = convertDateTimeLocal(dateFrom);
    const toFormatted = convertDateTimeLocal(dateTo);

    const params = new URLSearchParams({
        tag,
        from: fromFormatted,
        to: toFormatted,
        format: 'json'
    });

    const normalizedCount = parseInt(searchCount, 10);
    if (!Number.isNaN(normalizedCount) && normalizedCount > 0) {
        params.set('count', String(normalizedCount));
    }

    const url = `${api}/get/tag/?${params.toString()}`;

    fetch(url)
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            return response.text();
        })
        .then(data => {
            let parsedData;
            try {
                parsedData = JSON.parse(data);
            } catch (e) {
                parsedData = data;
            }

            const normalizedData = normalizeApiData(parsedData, tag);
            updateDataTable(normalizedData, tag);

            const rowsCount = Array.isArray(normalizedData) ? normalizedData.length : String(data).split('\n').filter(Boolean).length;
            showSuccessNotification(`Найдено записей: ${rowsCount}`);
        })
        .catch(error => {
            showErrorNotification(`Ошибка загрузки данных: ${error.message}`);
            
            // Показываем пустую таблицу
            const tbody = document.getElementById('data-results');
            if (tbody) {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="6" class="text-center py-8 text-gray-500">
                            <div class="flex flex-col items-center space-y-3">
                                <svg class="h-12 w-12 text-red-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                                <div>
                                    <p class="text-lg font-medium text-gray-900">Ошибка загрузки</p>
                                    <p class="text-gray-500">${error.message}</p>
                                </div>
                            </div>
                        </td>
                    </tr>
                `;
            }
        })
        .finally(() => {
            if (searchBtn) {
                searchBtn.disabled = false;
                searchBtn.innerHTML = originalText;
            }
        });
}

function convertDateTimeLocal(datetimeLocal) {
    if (!datetimeLocal) return '';

    const date = new Date(datetimeLocal);
    if (Number.isNaN(date.getTime())) {
        return datetimeLocal;
    }

    const day = String(date.getDate()).padStart(2, '0');
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const year = date.getFullYear();
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');

    return `${day}.${month}.${year} ${hours}:${minutes}:00`;
}

function updateDataTable(data, currentTag) {
    const tbody = document.getElementById('data-results');
    if (!tbody) {
        return;
    }

    tbody.innerHTML = '';
    
    if (!data || (Array.isArray(data) && data.length === 0)) {
        tbody.innerHTML = `
            <tr>
                <td colspan="6" class="text-center py-8 text-gray-500">
                    <div class="flex flex-col items-center space-y-3">
                        <svg class="h-12 w-12 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-2m3 2v-4m3 4v-6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                        </svg>
                        <div>
                            <p class="text-lg font-medium text-gray-900">Нет данных</p>
                            <p class="text-gray-500">По указанным параметрам данные не найдены</p>
                        </div>
                    </div>
                </td>
            </tr>
        `;
        return;
    }
    
    if (typeof data === 'string') {
        if (data.startsWith('#Error:')) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="6" class="text-center py-8 text-red-500">
                        <div class="flex flex-col items-center space-y-3">
                            <svg class="h-12 w-12 text-red-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                            <div>
                                <p class="text-lg font-medium text-gray-900">Ошибка API</p>
                                <p class="text-red-500">${data}</p>
                            </div>
                        </div>
                    </td>
                </tr>
            `;
            return;
        }

        const lines = data.split('\n').filter(line => line.trim());

        if (lines.length === 0) {
            updateDataTable([], currentTag);
            return;
        }

        data = lines.map(line => {
            const parsed = parseDataString(line);
            return {
                timestamp: parsed.time,
                tag: parsed.tag || currentTag,
                value: parsed.value,
                quality: parsed.quality || 'OK',
                unit: parsed.unit || getUnitForTag(currentTag),
                description: parsed.description || getDescriptionForTag(currentTag)
            };
        });
    }

    data = normalizeApiData(data, currentTag);

    if (Array.isArray(data)) {
        data.forEach((item) => {
            const row = document.createElement('tr');
            row.className = 'data-row';
            
            let timestamp, tag, value, quality, unit, description;
            
            if (typeof item === 'object' && item !== null) {
                timestamp = formatTimestamp(item.timestamp || item.time || item.date);
                tag = item.tag || item.name || currentTag;
                value = formatValue(item.value);
                quality = item.quality || 'OK';
                unit = item.unit || getUnitForTag(tag);
                description = item.description || getDescriptionForTag(tag);
            } else {
                const parsed = parseDataString(String(item));
                timestamp = parsed.time;
                tag = parsed.tag || currentTag;
                value = parsed.value;
                quality = parsed.quality || 'OK';
                unit = parsed.unit || getUnitForTag(tag);
                description = parsed.description || getDescriptionForTag(tag);
            }
            
            row.innerHTML = `
                <td class="col-time">${timestamp}</td>
                <td class="col-tag">${tag}</td>
                <td class="col-value value-cell" data-value="${value}">${value}</td>
                <td class="col-quality">
                    <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getQualityClass(quality)}">
                        ${quality}
                    </span>
                </td>
                <td class="col-unit">${unit}</td>
                <td class="col-description">${description}</td>
            `;
            
            tbody.appendChild(row);
        });

        styleValueCells();
        return;
    }

    tbody.innerHTML = `
        <tr>
            <td colspan="6" class="text-center py-8 text-gray-500">
                <div class="flex flex-col items-center space-y-3">
                    <svg class="h-12 w-12 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-2m3 2v-4m3 4v-6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    <div>
                        <p class="text-lg font-medium text-gray-900">Неподдерживаемый формат данных</p>
                        <p class="text-gray-500">Проверьте параметры запроса или формат ответа API</p>
                    </div>
                </div>
            </td>
        </tr>
    `;
}

function normalizeApiData(data, currentTag) {
    if (!data || Array.isArray(data) || typeof data === 'string') {
        return data;
    }

    if (typeof data !== 'object') {
        return data;
    }

    if (Array.isArray(data.rows)) {
        return data.rows.map((row) => ({
            timestamp: row[1] || row[0] || '',
            tag: row[0] || currentTag,
            value: row[2] ?? row[1] ?? '',
            quality: 'OK',
            unit: getUnitForTag(row[0] || currentTag),
            description: getDescriptionForTag(row[0] || currentTag)
        }));
    }

    if (Object.prototype.hasOwnProperty.call(data, 'value')) {
        return [{
            timestamp: new Date().toISOString(),
            tag: currentTag,
            value: data.value,
            quality: 'OK',
            unit: getUnitForTag(currentTag),
            description: getDescriptionForTag(currentTag)
        }];
    }

    const normalizedRows = [];
    let nestedSeriesDetected = false;

    Object.entries(data).forEach(([tagName, payload]) => {
        if (payload && typeof payload === 'object' && !Array.isArray(payload)) {
            nestedSeriesDetected = true;
            Object.entries(payload).forEach(([timestamp, value]) => {
                normalizedRows.push({
                    timestamp,
                    tag: tagName || currentTag,
                    value,
                    quality: 'OK',
                    unit: getUnitForTag(tagName || currentTag),
                    description: getDescriptionForTag(tagName || currentTag)
                });
            });
        }
    });

    if (nestedSeriesDetected) {
        normalizedRows.sort((a, b) => {
            const ta = Date.parse(String(a.timestamp).replace(' ', 'T'));
            const tb = Date.parse(String(b.timestamp).replace(' ', 'T'));
            if (Number.isNaN(ta) || Number.isNaN(tb)) {
                return String(a.timestamp).localeCompare(String(b.timestamp));
            }
            return ta - tb;
        });
        return normalizedRows;
    }

    const simpleSeries = Object.entries(data);
    if (simpleSeries.length > 0 && simpleSeries.every(([, value]) => typeof value === 'string' || typeof value === 'number')) {
        return simpleSeries.map(([timestamp, value]) => ({
            timestamp,
            tag: currentTag,
            value,
            quality: 'OK',
            unit: getUnitForTag(currentTag),
            description: getDescriptionForTag(currentTag)
        }));
    }

    return data;
}

function getTagList() {
    const tag = document.getElementById("searchInput")?.value?.trim() || '';
    if (!tag) {
        showErrorNotification('Введите маску для поиска');
        return;
    }

    const url = `/tags/?like=${encodeURIComponent(tag)}`;

    if (typeof window.loadPage === 'function') {
        window.loadPage(url);
    }
}

// load comprehensive home page data
function loadHomePageData() {
    if (!document.getElementById('apiserver')) {
        return;
    }
    
    const api = document.getElementById('apiserver').textContent;
    
    // load statistics
    if (typeof window.loadStatistics === 'function') {
        window.loadStatistics(api);
    }
    
    // load recent activity
    if (typeof window.loadRecentActivity === 'function') {
        window.loadRecentActivity(api);
    }
}

function loadSwagger() {
    if (document.getElementById("content")!=null) {
        document.getElementById("content").innerHTML = '<iframe src="/swagger/" style="text-center" width="800px" height="100%" frameborder="0"></iframe>';
    } else {
        showErrorNotification('Контейнер Swagger не найден');
    }
}

function initializeDataPage() {
    const now = new Date();
    const fiveMinutesAgo = new Date(now.getTime() - 5 * 60000);

    const params = new URLSearchParams(window.location.search);

    const searchInput = document.getElementById('searchInput');
    const dateFrom = document.getElementById('dateFrom');
    const dateTo = document.getElementById('dateTo');
    const searchCount = document.getElementById('searchCount');

    if (searchInput && params.get('tag')) {
        searchInput.value = params.get('tag');
    }
    if (dateFrom && params.get('from')) {
        dateFrom.value = params.get('from');
    }
    if (dateTo && params.get('to')) {
        dateTo.value = params.get('to');
    }
    if (searchCount && params.get('count')) {
        searchCount.value = params.get('count');
    }

    if (dateFrom && !dateFrom.value) {
        dateFrom.value = fiveMinutesAgo.toISOString().slice(0, 16);
    }
    if (dateTo && !dateTo.value) {
        dateTo.value = now.toISOString().slice(0, 16);
    }

    formatServerRenderedRows();
}

function formatServerRenderedRows() {
    const rows = document.querySelectorAll('#data-results tr.data-row');
    if (!rows.length) {
        return;
    }

    rows.forEach((row) => {
        const raw = row.getAttribute('data-raw');
        if (!raw) return;
        const parsed = parseDataString(raw);

        const timeCell = row.querySelector('.col-time');
        const tagCell = row.querySelector('.col-tag');
        const valueCell = row.querySelector('.col-value');
        const qualityCell = row.querySelector('.col-quality span');
        const unitCell = row.querySelector('.col-unit');
        const descriptionCell = row.querySelector('.col-description');

        if (timeCell) timeCell.textContent = parsed.time;
        if (tagCell) tagCell.textContent = parsed.tag;
        if (valueCell) {
            valueCell.textContent = parsed.value;
            valueCell.setAttribute('data-value', parsed.value);
            valueCell.classList.add('value-cell');
        }
        if (qualityCell) {
            qualityCell.textContent = parsed.quality;
            qualityCell.className = `inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getQualityClass(parsed.quality)}`;
        }
        if (unitCell) unitCell.textContent = parsed.unit || '—';
        if (descriptionCell) descriptionCell.textContent = parsed.description || '—';
    });

    styleValueCells();
}

// Дополнительные функции для работы с данными

function styleValueCells() {
    const valueCells = document.querySelectorAll('.value-cell');
    valueCells.forEach(cell => {
        const value = parseFloat(cell.getAttribute('data-value'));
        
        if (!isNaN(value)) {
            cell.classList.remove('data-value-positive', 'data-value-negative', 'data-value-zero');
            
            if (value > 0) {
                cell.classList.add('data-value-positive');
            } else if (value < 0) {
                cell.classList.add('data-value-negative');
            } else {
                cell.classList.add('data-value-zero');
            }
        }
    });
}

function getQualityClass(quality) {
    const q = String(quality).toLowerCase();
    switch(q) {
        case 'ok':
        case 'good':
            return 'quality-ok';
        case 'bad':
        case 'error':
            return 'quality-bad';
        case 'uncertain':
        case 'warning':
            return 'quality-warning';
        case 'unknown':
        default:
            return 'quality-unknown';
    }
}

function formatTimestamp(timestamp) {
    if (!timestamp) return '--:--:--';
    
    try {
        let date;
        
        // Если уже отформатированная строка DD.MM.YYYY HH:MM:SS
        if (typeof timestamp === 'string' && timestamp.includes('.') && timestamp.includes(':')) {
            return timestamp;
        }
        
        // Если это Date объект или строка даты
        if (timestamp instanceof Date) {
            date = timestamp;
        } else if (typeof timestamp === 'string') {
            const normalized = timestamp.replace(/\s\+\d{4}\sUTC$/, 'Z').replace(' ', 'T');
            date = new Date(normalized);
            if (isNaN(date.getTime())) {
                date = new Date(timestamp);
            }
        } else {
            return String(timestamp);
        }
        
        if (isNaN(date.getTime())) {
            return String(timestamp);
        }
        
        // Форматируем как DD.MM.YYYY HH:MM:SS
        const day = String(date.getDate()).padStart(2, '0');
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const year = date.getFullYear();
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');
        const seconds = String(date.getSeconds()).padStart(2, '0');
        
        return `${day}.${month}.${year} ${hours}:${minutes}:${seconds}`;
    } catch (e) {
        return String(timestamp);
    }
}

function formatValue(value) {
    if (value === null || value === undefined) return '—';

    const normalized = typeof value === 'string' ? value.replace(',', '.') : value;
    const num = parseFloat(normalized);
    if (isNaN(num)) return String(value);
    
    // Округляем до 2 знаков после запятой
    return num.toFixed(2);
}

function getCurrentTag() {
    const tagInput = document.getElementById('searchInput');
    return tagInput ? tagInput.value || 'Unknown' : 'Unknown';
}

function parseDataString(rawData) {
    const fallbackTag = getCurrentTag();
    const cleaned = String(rawData || '').trim();

    if (!cleaned) {
        return {
            time: formatTimestamp(new Date()),
            tag: fallbackTag,
            value: '—',
            quality: 'Unknown',
            unit: '—',
            description: '—'
        };
    }

    try {
        // Strategy 0: server format "timestamp|value"
        if (cleaned.includes('|')) {
            const [rawTime, rawValue] = cleaned.split('|');
            return {
                time: formatTimestamp(rawTime.trim()),
                tag: fallbackTag,
                value: formatValue(rawValue),
                quality: 'OK',
                unit: getUnitForTag(fallbackTag),
                description: getDescriptionForTag(fallbackTag)
            };
        }

        // Strategy 1: Space-separated format "DD.MM.YYYY HH:MM:SS value"
        const timeValueMatch = cleaned.match(/^(\d{2}\.\d{2}\.\d{4}\s+\d{2}:\d{2}:\d{2})\s+(.+)$/);
        if (timeValueMatch) {
            const [, timestamp, value] = timeValueMatch;
            
            return {
                time: timestamp,
                tag: fallbackTag,
                value: formatValue(value),
                quality: 'OK',
                unit: getUnitForTag(fallbackTag),
                description: getDescriptionForTag(fallbackTag)
            };
        }

        // Strategy 2: Формат времени ISO или UTC
        const isoTimeMatch = cleaned.match(/^(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}(?:\s+[+]\d{4}\s+UTC)?)\s+(.+)$/);
        if (isoTimeMatch) {
            const [, timestamp, value] = isoTimeMatch;

            const date = new Date(timestamp.replace(/\s+[+]\d{4}\s+UTC/, ''));
            const formattedTime = formatTimestamp(date);
            
            return {
                time: formattedTime,
                tag: fallbackTag,
                value: formatValue(value),
                quality: 'OK',
                unit: getUnitForTag(fallbackTag),
                description: getDescriptionForTag(fallbackTag)
            };
        }

        // Strategy 3: Format "10.02.2023 21:00:01 208.48"
        const parts = cleaned.split(/\s+/);

        if (parts.length >= 3) {
            const date = parts[0];
            const time = parts[1];
            const value = parts[2];
            
            if (date.match(/^\d{2}\.\d{2}\.\d{4}$/) && time.match(/^\d{2}:\d{2}:\d{2}$/)) {
                const timestamp = `${date} ${time}`;
                
                return {
                    time: timestamp,
                    tag: fallbackTag,
                    value: formatValue(value),
                    quality: 'OK',
                    unit: getUnitForTag(fallbackTag),
                    description: getDescriptionForTag(fallbackTag)
                };
            }
        }

        // Strategy 4: Просто числовое значение
        const numericValue = parseFloat(cleaned);
        if (!isNaN(numericValue)) {
            return {
                time: formatTimestamp(new Date()),
                tag: fallbackTag,
                value: formatValue(numericValue),
                quality: 'OK',
                unit: getUnitForTag(fallbackTag),
                description: getDescriptionForTag(fallbackTag)
            };
        }

        // Strategy 5: Пытаемся парсить как JSON
        try {
            const jsonData = JSON.parse(cleaned);
            
            return {
                time: formatTimestamp(jsonData.timestamp || jsonData.time || jsonData.date || new Date()),
                tag: jsonData.tag || fallbackTag,
                value: formatValue(jsonData.value),
                quality: jsonData.quality || 'OK',
                unit: jsonData.unit || getUnitForTag(fallbackTag),
                description: jsonData.description || getDescriptionForTag(fallbackTag)
            };
        } catch (jsonError) {
            // Not JSON.
        }
        
    } catch (e) {
        // Ignore and use fallback below.
    }
    
    return {
        time: formatTimestamp(new Date()),
        tag: fallbackTag,
        value: cleaned || '—',
        quality: 'Unknown',
        unit: '—',
        description: '—'
    };
}

function getUnitForTag(tag) {
    // Используем декодированный тег для определения единиц измерения
    const decodedTag = decodeTag(tag);
    
    // Определяем единицы измерения на основе типа устройства
    if (decodedTag.device_type) {
        const deviceType = decodedTag.device_type;
        
        if (deviceType.includes('температур')) return '°C';
        if (deviceType.includes('давлен')) return 'bar';
        if (deviceType.includes('расходомер') || deviceType.includes('счетчик расходомера')) return 'm³/h';
        if (deviceType.includes('уровнемер')) return 'm';
        if (deviceType.includes('вес') || deviceType.includes('масса') || deviceType.includes('счетчик веса')) return 't';
        if (deviceType.includes('насос') || deviceType.includes('агитатор') || deviceType.includes('вентилятор')) return 'об/мин';
        if (deviceType.includes('питатель') || deviceType.includes('дробилка') || deviceType.includes('конвейер')) return 't/h';
    }
    
    // Дополнительная проверка по имени тега
    const tagLower = tag.toLowerCase();
    
    if (tagLower.includes('temp') || tagLower.includes('_tt_') || tagLower.includes('_ti_')) return '°C';
    if (tagLower.includes('press') || tagLower.includes('_pt_') || tagLower.includes('_pi_')) return 'bar';
    if (tagLower.includes('flow') || tagLower.includes('_ft_') || tagLower.includes('_fi_') || tagLower.includes('_fqt_')) return 'm³/h';
    if (tagLower.includes('level') || tagLower.includes('_lt_') || tagLower.includes('_li_')) return 'm';
    if (tagLower.includes('_wt_') || tagLower.includes('_wqt_') || tagLower.includes('mass')) return 't';
    if (tagLower.includes('volt') || tagLower.includes('_v_')) return 'V';
    if (tagLower.includes('current') || tagLower.includes('_i_')) return 'A';
    if (tagLower.includes('power') || tagLower.includes('_w_')) return 'W';
    if (tagLower.includes('freq') || tagLower.includes('_f_')) return 'Hz';
    if (tagLower.includes('speed') || tagLower.includes('rpm')) return 'об/мин';
    if (tagLower.includes('_pmp_') || tagLower.includes('_agt_') || tagLower.includes('_fan_')) return 'об/мин';
    
    // Для состояний и тревог единицы измерения не нужны
    if (decodedTag.tag_type === 'alarm' || decodedTag.tag_type === 'state') {
        return '';
    }
    
    return '—';
}

function getDescriptionForTag(tag) {
    // Создаем объект с описанием тега
    const decodedTag = decodeTag(tag);
    
    // Формируем описание из декодированных данных
    let description = '';
    
    if (decodedTag.device_type && decodedTag.device_num) {
        description = `${decodedTag.device_type} №${decodedTag.device_num}`;
    } else if (decodedTag.device_type) {
        description = decodedTag.device_type;
    } else {
        description = 'технологический параметр';
    }
    
    if (decodedTag.area_descr) {
        description += ` (${decodedTag.area_descr})`;
    } else if (decodedTag.area) {
        description += ` (${decodedTag.area})`;
    }
    
    if (decodedTag.tag_descr) {
        description += ` - ${decodedTag.tag_descr}`;
    }
    
    // Первая буква заглавная
    return description.charAt(0).toUpperCase() + description.slice(1);
}

function decodeTag(tagName) {
    const decoded = {
        tag_name: tagName
    };
    
    // Области (A10, A15, A20 и т.д.)
    const areaRegex = /^A(\d{2})/;
    const areaMatch = tagName.match(areaRegex);
    if (areaMatch) {
        decoded.area = areaMatch[0];
        const areaMap = {
            'A10': 'Дробление',
            'A11': 'Тоннель золотой цепочки',
            'A15': 'Тоннель медной цепочки',
            'A20': 'Измельчение золотой цепочки',
            'A25': 'Измельчение медной цепочки',
            'A30': 'Trash screening, CIP',
            'A31': 'Регенерация',
            'A32': 'Детоксикация',
            'A35': 'Флотация',
            'A36': 'Очистка флотации',
            'A37': 'Перечистка флотации',
            'A40': 'Acid wash',
            'A45': 'Элюация',
            'A50': 'Goldroom',
            'A55': 'Сгущение',
            'A70': 'Water dist',
            'A71': 'Fire water',
            'A80': 'Цианирование',
            'A81': 'Air service',
            'A85': 'Флокулянт'
        };
        decoded.area_descr = areaMap[decoded.area] || decoded.area;
    }
    
    // Типы устройств
    const deviceRegex = /_(TT|TI|PT|PI|F(?:|Q)T|FI|LT|LI|SIREN|FAN|FPC|PMP|HTR|FCV|AGT|ISC|APF|CRU|CVR|FDR|HPP|SCR|WT|WQT|FTP|MASS|SMP)(?:_)?(\d{1,2}(?:\.)?)/;
    const deviceMatch = tagName.match(deviceRegex);
    if (deviceMatch) {
        const deviceType = deviceMatch[1];
        decoded.device_num = deviceMatch[2];
        
        const deviceMap = {
            'TT': 'датчик температуры',
            'TI': 'датчик температуры',
            'PT': 'датчик давления',
            'PI': 'датчик давления',
            'FT': 'расходомер',
            'FI': 'расходомер',
            'FQT': 'счетчик расходомера',
            'LT': 'уровнемер',
            'LI': 'уровнемер',
            'SIREN': 'сирена',
            'FAN': 'вентилятор',
            'FPC': 'контроллер вентилятора',
            'PMP': 'насос',
            'HTR': 'подогреватель',
            'FCV': 'клапан',
            'AGT': 'агитатор',
            'ISC': 'перекачной насос',
            'APF': 'питатель пластинчатый',
            'CRU': 'дробилка',
            'CVR': 'конвейер',
            'FDR': 'вибропитатель',
            'HPP': 'hydraulic power pack',
            'SCR': 'conveyer scrubber',
            'WT': 'вес',
            'WQT': 'счетчик веса',
            'FTP': 'фильтр-пресс',
            'MASS': 'масса',
            'SMP': 'пробоотборник'
        };
        decoded.device_type = deviceMap[deviceType] || deviceType.toLowerCase();
    }
    
    // Тревоги
    const alarmRegex = /_(AH|AHH|AL|ALL|ALARM|ALM|CBRS(?:|1|2|3|4))_/;
    const alarmMatch = tagName.match(alarmRegex);
    if (alarmMatch) {
        decoded.tag_type = 'alarm';
        const alarmType = alarmMatch[1];
        const alarmMap = {
            'AH': 'высокий уровень',
            'AHH': 'критически высокий уровень',
            'AL': 'низкий уровень',
            'ALL': 'критически низкий уровень',
            'ALARM': 'тревога',
            'ALM': 'тревога',
            'CBRS': 'тревога',
            'CBRS1': 'тревога',
            'CBRS2': 'тревога',
            'CBRS3': 'тревога',
            'CBRS4': 'тревога'
        };
        decoded.tag_descr = alarmMap[alarmType] || 'тревога';
    }
    
    // Значения тревог
    const alarmValueRegex = /_(HI|HIHI|LO|LOLO)_/;
    const alarmValueMatch = tagName.match(alarmValueRegex);
    if (alarmValueMatch) {
        decoded.tag_type = 'alarm';
        const valueType = alarmValueMatch[1];
        const valueMap = {
            'HI': 'высокий уровень - значение',
            'HIHI': 'критически высокий уровень - значение',
            'LO': 'низкий уровень - значение',
            'LOLO': 'критически низкий уровень - значение'
        };
        decoded.tag_descr = valueMap[valueType] || 'значение тревоги';
    }
    
    // Состояния
    const stateRegex = /_(URS|UMH|SAS|SST|SSP|DQS|SLR|DFST|USH|DIR|SDI|HR|DMR|DOF|XY|RST|ET|PR)_/;
    const stateMatch = tagName.match(stateRegex);
    if (stateMatch) {
        decoded.tag_type = 'state';
        const stateType = stateMatch[1];
        const stateMap = {
            'URS': 'в работе',
            'UMH': 'MCC статус',
            'SAS': 'авто/мануал',
            'SST': 'scada старт',
            'SSP': 'scada стоп',
            'DQS': 'drive sequence start',
            'SLR': 'локал/ремоут',
            'DFST': 'старт по месту',
            'USH': 'стоп по месту',
            'DIR': 'отсутствие блокировок',
            'SDI': 'блокировки отключены',
            'HR': 'сброс моточасов',
            'DMR': 'готовность',
            'DOF': 'ошибка запуска',
            'XY': 'команда запуска',
            'RST': 'сброс',
            'ET': 'время ожидания',
            'PR': 'шаг'
        };
        decoded.tag_descr = stateMap[stateType] || 'состояние';
    }
    
    // Моточасы
    const motohourRegex = /_(DRH|DRM)_/;
    const motohourMatch = tagName.match(motohourRegex);
    if (motohourMatch) {
        const timeType = motohourMatch[1];
        const timeMap = {
            'DRH': 'часы',
            'DRM': 'минуты'
        };
        decoded.tag_descr = timeMap[timeType] || 'время работы';
    }
    
    // Ручные описания для специальных тегов
    const manualDescriptions = {
        'A15_RST_RST_WQT_03_TOT': 'Сброс веса счетчика медного конвейера'
    };
    
    if (manualDescriptions[tagName]) {
        decoded.tag_hand = manualDescriptions[tagName];
        return { ...decoded, description: manualDescriptions[tagName] };
    }
    
    return decoded;
}

function searchTagData(tag) {
    const now = new Date();
    const fiveMinutesAgo = new Date(now.getTime() - 5 * 60000);
    const from = fiveMinutesAgo.toISOString().slice(0, 16);
    const to = now.toISOString().slice(0, 16);

    const encodedTag = encodeURIComponent(tag);
    const encodedFrom = encodeURIComponent(from);
    const encodedTo = encodeURIComponent(to);

    if (typeof window.loadPage === 'function') {
        window.loadPage(`/data/?tag=${encodedTag}&from=${encodedFrom}&to=${encodedTo}`);
    }
}

export { 
    getTagOnDate,
    getTagList, 
    loadHomePageData,
    loadSwagger,
    initializeDataPage,
    updateDataTable,
    searchTagData,
    formatTimestamp,
    formatValue,
    getCurrentTag,
    getUnitForTag,
    getDescriptionForTag
}; 
