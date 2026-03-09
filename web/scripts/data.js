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
            updateDataInsights(normalizedData, tag, dateFrom, dateTo);

            const rowsCount = Array.isArray(normalizedData) ? normalizedData.length : String(data).split('\n').filter(Boolean).length;
            showSuccessNotification(`Найдено записей: ${rowsCount}`);
        })
        .catch(error => {
            showErrorNotification(`Ошибка загрузки данных: ${error.message}`);
            resetDataInsights();
            
            // Показываем пустую таблицу
            const tbody = document.getElementById('data-results');
            if (tbody) {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="6" class="table-empty-cell">
                            <div class="table-empty-state table-empty-state-error">
                                <svg class="empty-state-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                                <div>
                                    <p>Ошибка загрузки</p>
                                    <span>${error.message}</span>
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

function applyDataRangePreset(preset) {
    const dateFrom = document.getElementById('dateFrom');
    const dateTo = document.getElementById('dateTo');
    if (!dateFrom || !dateTo) return;

    const now = new Date();
    const from = new Date(now);
    const presetMap = {
        '5m': 5 * 60 * 1000,
        '1h': 60 * 60 * 1000,
        '8h': 8 * 60 * 60 * 1000,
        '24h': 24 * 60 * 60 * 1000
    };
    from.setTime(now.getTime() - (presetMap[preset] || 5 * 60 * 1000));

    dateFrom.value = from.toISOString().slice(0, 16);
    dateTo.value = now.toISOString().slice(0, 16);
    syncSnapshotDate();
}

function syncSnapshotDate() {
    const snapshotDate = document.getElementById('snapshotDate');
    const dateTo = document.getElementById('dateTo');
    if (snapshotDate && dateTo && dateTo.value) {
        snapshotDate.value = dateTo.value;
    }
}

function getPrimaryTag() {
    const raw = document.getElementById('searchInput')?.value || '';
    return String(raw)
        .split(',')
        .map((tag) => tag.trim())
        .filter(Boolean)[0] || '';
}

function getDataRangeParams() {
    const tag = getPrimaryTag();
    const dateFrom = document.getElementById('dateFrom')?.value || '';
    const dateTo = document.getElementById('dateTo')?.value || '';

    return {
        tag,
        from: convertDateTimeLocal(dateFrom),
        to: convertDateTimeLocal(dateTo),
        fromLocal: dateFrom,
        toLocal: dateTo
    };
}

async function fetchDataApi(path, params, expectJson = false) {
    const apiElement = document.getElementById('apiserver');
    if (!apiElement) throw new Error('API сервер недоступен');

    const search = new URLSearchParams();
    Object.entries(params || {}).forEach(([key, value]) => {
        if (value !== undefined && value !== null && String(value) !== '') {
            search.set(key, String(value));
        }
    });

    const response = await fetch(`${apiElement.textContent.trim()}${path}?${search.toString()}`);
    if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    const text = await response.text();
    if (text.startsWith('#Error:')) {
        throw new Error(text);
    }

    if (!expectJson) {
        return text;
    }

    try {
        return JSON.parse(text);
    } catch (_) {
        return text;
    }
}

function setPanelResult(id, html) {
    const el = document.getElementById(id);
    if (el) el.innerHTML = html;
}

function resetDataInsights() {
    const defaults = {
        dataSummaryPoints: '0',
        dataSummaryWindow: '—',
        dataSummaryMin: '—',
        dataSummaryMax: '—',
        dataSummaryAvg: '—'
    };
    Object.entries(defaults).forEach(([id, value]) => {
        const el = document.getElementById(id);
        if (el) el.textContent = value;
    });

    setPanelResult('snapshotResult', 'Выбери тег и дату');
    setPanelResult('dataDecodeResult', 'Нажми decode для текущего тега');
    setPanelResult('eventResult', 'Используется текущий тег и текущий диапазон');
    setPanelResult('aggregateResult', insightRow('AVG', '—'));
}

function updateDataInsights(data, currentTag, from, to) {
    if (!Array.isArray(data) || !data.length) {
        resetDataInsights();
        return;
    }

    const numericValues = data
        .map((item) => Number.parseFloat(String(item?.value ?? '').replace(',', '.')))
        .filter((value) => Number.isFinite(value));

    const points = document.getElementById('dataSummaryPoints');
    const windowEl = document.getElementById('dataSummaryWindow');
    const minEl = document.getElementById('dataSummaryMin');
    const maxEl = document.getElementById('dataSummaryMax');
    const avgEl = document.getElementById('dataSummaryAvg');

    if (points) points.textContent = String(data.length);
    if (windowEl) windowEl.textContent = describeWindow(from, to);

    if (!numericValues.length) {
        if (minEl) minEl.textContent = '—';
        if (maxEl) maxEl.textContent = '—';
        if (avgEl) avgEl.textContent = '—';
        return;
    }

    const min = Math.min(...numericValues);
    const max = Math.max(...numericValues);
    const avg = numericValues.reduce((sum, value) => sum + value, 0) / numericValues.length;

    if (minEl) minEl.textContent = formatValue(min);
    if (maxEl) maxEl.textContent = formatValue(max);
    if (avgEl) avgEl.textContent = formatValue(avg);

    if (currentTag) {
        setPanelResult('aggregateResult', [
            insightRow('AVG', formatValue(avg)),
            insightRow('MIN', formatValue(min)),
            insightRow('MAX', formatValue(max)),
            insightRow('COUNT', String(data.length))
        ].join(''));
    }
}

function describeWindow(from, to) {
    if (!from || !to) return '—';
    const fromDate = new Date(from);
    const toDate = new Date(to);
    if (Number.isNaN(fromDate.getTime()) || Number.isNaN(toDate.getTime())) return '—';
    const diff = Math.max(0, toDate.getTime() - fromDate.getTime());
    const minutes = Math.round(diff / 60000);
    if (minutes < 60) return `${minutes} мин`;
    const hours = (minutes / 60).toFixed(minutes % 60 === 0 ? 0 : 1);
    return `${hours} ч`;
}

function insightRow(label, value) {
    return `<div class="insight-row"><span>${label}</span><strong>${value}</strong></div>`;
}

async function queryCurrentValue() {
    const tag = getPrimaryTag();
    const snapshotDate = document.getElementById('snapshotDate')?.value || document.getElementById('dateTo')?.value || '';
    if (!tag) {
        showErrorNotification('Сначала укажи тег');
        return;
    }
    if (!snapshotDate) {
        showErrorNotification('Укажи дату снимка');
        return;
    }

    setPanelResult('snapshotResult', 'Загрузка...');
    try {
        const payload = await fetchDataApi('/get/tag/', {
            tag,
            date: convertDateTimeLocal(snapshotDate),
            format: 'json'
        }, true);
        const value = extractScalarValue(payload);
        setPanelResult('snapshotResult', `<strong>${formatValue(value)}</strong><span>${tag}</span>`);
    } catch (error) {
        setPanelResult('snapshotResult', `<span class="result-error">${error.message}</span>`);
    }
}

async function queryAggregate(group) {
    const { tag, from, to } = getDataRangeParams();
    if (!tag) {
        showErrorNotification('Сначала укажи тег');
        return;
    }
    if (!from || !to) {
        showErrorNotification('Укажи диапазон');
        return;
    }

    setPanelResult('aggregateResult', insightRow(group.toUpperCase(), '...'));
    try {
        const payload = await fetchDataApi('/get/tag/', {
            tag,
            from,
            to,
            group,
            format: 'json'
        }, true);
        const value = extractScalarValue(payload);
        setPanelResult('aggregateResult', insightRow(group.toUpperCase(), formatValue(value)));
    } catch (error) {
        setPanelResult('aggregateResult', `<div class="result-error">${error.message}</div>`);
    }
}

async function queryEvent(type) {
    const { tag, from, to } = getDataRangeParams();
    const count = document.getElementById('eventIndex')?.value || '0';
    if (!tag) {
        showErrorNotification('Сначала укажи тег');
        return;
    }
    if (!from || !to) {
        showErrorNotification('Укажи диапазон');
        return;
    }

    const endpoint = type === 'down' ? '/get/tag/down/' : '/get/tag/up/';
    setPanelResult('eventResult', 'Загрузка...');
    try {
        const text = await fetchDataApi(endpoint, { tag, from, to, count });
        setPanelResult('eventResult', `<strong>${text || 'Не найдено'}</strong>`);
    } catch (error) {
        setPanelResult('eventResult', `<span class="result-error">${error.message}</span>`);
    }
}

async function decodeCurrentTag() {
    const tag = getPrimaryTag();
    if (!tag) {
        showErrorNotification('Сначала укажи тег');
        return;
    }

    setPanelResult('dataDecodeResult', 'Загрузка...');
    try {
        const payload = await fetchDataApi('/tag/decode/', { tag, format: 'json' }, true);
        const data = payload?.[tag] || payload?.tag || payload;
        if (!data || typeof data !== 'object') {
            setPanelResult('dataDecodeResult', `<span>${String(data || 'Нет данных')}</span>`);
            return;
        }

        const rows = Object.entries(data)
            .filter(([, value]) => value !== null && value !== undefined && String(value).trim() !== '')
            .map(([key, value]) => insightRow(key, escapeHtml(String(value))));

        setPanelResult('dataDecodeResult', rows.length ? rows.join('') : 'Пустой ответ');
    } catch (error) {
        setPanelResult('dataDecodeResult', `<span class="result-error">${error.message}</span>`);
    }
}

function openDataInCharts() {
    const tagValue = document.getElementById('searchInput')?.value?.trim() || '';
    if (!tagValue) {
        showErrorNotification('Сначала укажи тег');
        return;
    }

    const params = new URLSearchParams();
    params.set('tags', tagValue);

    const dateFrom = document.getElementById('dateFrom')?.value || '';
    const dateTo = document.getElementById('dateTo')?.value || '';
    const count = document.getElementById('searchCount')?.value || '300';

    if (dateFrom) params.set('from', dateFrom);
    if (dateTo) params.set('to', dateTo);
    if (count) params.set('count', count);

    if (typeof window.loadPage === 'function') {
        window.loadPage(`/charts/?${params.toString()}`);
    }
}

function extractScalarValue(payload) {
    if (payload && typeof payload === 'object') {
        if (Object.prototype.hasOwnProperty.call(payload, 'value')) {
            return payload.value;
        }
        if (Array.isArray(payload.rows) && payload.rows[0] && payload.rows[0].length) {
            return payload.rows[0][payload.rows[0].length - 1];
        }
    }
    return payload;
}

function escapeHtml(value) {
    return String(value)
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#39;');
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
                <td colspan="6" class="table-empty-cell">
                    <div class="table-empty-state">
                        <svg class="empty-state-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-2m3 2v-4m3 4v-6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                        </svg>
                        <div>
                            <p>Нет данных</p>
                            <span>По указанным параметрам данные не найдены</span>
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
                    <td colspan="6" class="table-empty-cell">
                        <div class="table-empty-state table-empty-state-error">
                            <svg class="empty-state-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                            <div>
                                <p>Ошибка API</p>
                                <span>${data}</span>
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
                    <span class="quality-pill ${getQualityClass(quality)}">
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
            <td colspan="6" class="table-empty-cell">
                <div class="table-empty-state">
                    <svg class="empty-state-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-2m3 2v-4m3 4v-6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    <div>
                        <p>Неподдерживаемый формат данных</p>
                        <span>Проверьте параметры запроса или формат ответа API</span>
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

    syncSnapshotDate();
    dateTo?.addEventListener('change', syncSnapshotDate);
    formatServerRenderedRows();
    updateDataInsights(readTableRowsAsData(), searchInput?.value || '', dateFrom?.value || '', dateTo?.value || '');
}

function readTableRowsAsData() {
    const rows = document.querySelectorAll('#data-results tr.data-row');
    return Array.from(rows)
        .map((row) => {
            const raw = row.getAttribute('data-raw') || row.getAttribute('data-original') || '';
            if (!raw) return null;

            const parsed = parseDataString(raw);
            return {
                timestamp: parsed.time,
                tag: parsed.tag || getPrimaryTag() || getCurrentTag(),
                value: parsed.value,
                quality: parsed.quality || 'OK',
                unit: parsed.unit || getUnitForTag(parsed.tag || getCurrentTag()),
                description: parsed.description || getDescriptionForTag(parsed.tag || getCurrentTag())
            };
        })
        .filter(Boolean);
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
            qualityCell.className = `quality-pill ${getQualityClass(parsed.quality)}`;
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
    applyDataRangePreset,
    getTagList, 
    loadHomePageData,
    loadSwagger,
    initializeDataPage,
    updateDataTable,
    queryCurrentValue,
    queryAggregate,
    queryEvent,
    decodeCurrentTag,
    openDataInCharts,
    searchTagData,
    formatTimestamp,
    formatValue,
    getCurrentTag,
    getUnitForTag,
    getDescriptionForTag
}; 
