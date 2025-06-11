import { showErrorNotification, showSuccessNotification } from './ui.js';

function getTagOnDate() {
    console.log('getTagOnDate called');
    
    // Показываем индикатор загрузки
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
    
    // Получаем значения полей ввода
    const tag = document.getElementById("searchInput")?.value || '';
    const dateFrom = document.getElementById("dateFrom")?.value || '';
    const dateTo = document.getElementById("dateTo")?.value || '';
    const searchCount = document.getElementById("searchCount")?.value || '300';
    
    console.log('Search params:', { tag, dateFrom, dateTo, searchCount });
    
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
    
    // Получаем API endpoint
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
    
    // Конвертируем даты в нужный формат
    const fromFormatted = convertDateTimeLocal(dateFrom);
    const toFormatted = convertDateTimeLocal(dateTo);
    
    console.log('Formatted dates:', { fromFormatted, toFormatted });
    
    // Формируем URL для API запроса
    const url = `${api}/get/tag/?tag=${encodeURIComponent(tag)}&from=${encodeURIComponent(fromFormatted)}&to=${encodeURIComponent(toFormatted)}&format=json`;
    
    console.log('API URL:', url);
    
    // Делаем AJAX запрос
    fetch(url)
        .then(response => {
            console.log('Response status:', response.status);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            return response.text();
        })
        .then(data => {
            console.log('Raw response data:', data);
            
            // Пытаемся распарсить как JSON
            let parsedData;
            try {
                parsedData = JSON.parse(data);
                console.log('Parsed JSON data:', parsedData);
            } catch (e) {
                console.log('Not JSON, treating as text data');
                parsedData = data;
            }
            
            // Обновляем таблицу с полученными данными
            updateDataTable(parsedData, tag);
            
            showSuccessNotification(`Найдено записей: ${Array.isArray(parsedData) ? parsedData.length : 'неизвестно'}`);
        })
        .catch(error => {
            console.error('Error fetching data:', error);
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
            // Восстанавливаем кнопку
            if (searchBtn) {
                searchBtn.disabled = false;
                searchBtn.innerHTML = originalText;
            }
        });
}

// Функция для конвертации datetime-local в формат API
function convertDateTimeLocal(datetimeLocal) {
    if (!datetimeLocal) return '';
    
    try {
        // datetime-local формат: 2023-10-02T21:00
        // API ожидает: 02.10.2023 21:00:00
        const date = new Date(datetimeLocal);
        const day = String(date.getDate()).padStart(2, '0');
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const year = date.getFullYear();
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');
        
        return `${day}.${month}.${year} ${hours}:${minutes}:00`;
    } catch (e) {
        console.error('Error converting date:', e);
        return datetimeLocal;
    }
}

// Функция для обновления таблицы данных
function updateDataTable(data, currentTag) {
    console.log('updateDataTable called with:', { data, currentTag });
    
    const tbody = document.getElementById('data-results');
    if (!tbody) {
        console.error('Table body not found');
        return;
    }
    
    // Очищаем таблицу
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
    
    // Если данные - строка, пытаемся её распарсить
    if (typeof data === 'string') {
        console.log('Data is string, trying to parse...');
        
        // Если строка содержит ошибку
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
        
        // Пытаемся разбить строку на строки данных
        const lines = data.split('\n').filter(line => line.trim());
        console.log('Split lines:', lines);
        
        if (lines.length === 0) {
            updateDataTable([], currentTag);
            return;
        }
        
        // Конвертируем строки в объекты данных
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
    
    // Если данные - массив объектов
    if (Array.isArray(data)) {
        console.log('Processing array data:', data);
        
        data.forEach((item, index) => {
            const row = document.createElement('tr');
            row.className = 'data-row';
            
            let timestamp, tag, value, quality, unit, description;
            
            if (typeof item === 'object' && item !== null) {
                // Если элемент - объект
                timestamp = formatTimestamp(item.timestamp || item.time || item.date);
                tag = item.tag || currentTag;
                value = formatValue(item.value);
                quality = item.quality || 'OK';
                unit = item.unit || getUnitForTag(tag);
                description = item.description || getDescriptionForTag(tag);
            } else {
                // Если элемент - строка
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
        
        // Применяем стилизацию значений
        styleValueCells();
    }
    
    console.log('Table updated successfully');
}

function getTagList() {
    var tag = document.getElementById("searchInput").value;

    if (!document.getElementById('apiserver')) {
        return;
    }
    const api = document.getElementById('apiserver').textContent;
    var url = api + "/tags/?like=" + tag;

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
        document.getElementById("content").innerHTML = '<iframe src="/swagger" style = "text-center" width="800px" height="100%" frameborder="0"></iframe>';
    } else {
        console.error('Element with ID "content" not found');
    }
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
    const q = quality.toLowerCase();
    switch(q) {
        case 'ok':
        case 'good':
            return 'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium quality-ok';
        case 'bad':
        case 'error':
            return 'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium quality-bad';
        case 'uncertain':
        case 'warning':
            return 'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium quality-warning';
        case 'unknown':
        default:
            return 'inline-flex items-center px-2 py-1 rounded-full text-xs font-medium quality-unknown';
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
            date = new Date(timestamp);
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
        console.warn('Error formatting timestamp:', timestamp, e);
        return String(timestamp);
    }
}

function formatValue(value) {
    if (value === null || value === undefined) return '—';
    
    const num = parseFloat(value);
    if (isNaN(num)) return String(value);
    
    // Округляем до 2 знаков после запятой
    return num.toFixed(2);
}

function getCurrentTag() {
    const tagInput = document.getElementById('searchInput');
    return tagInput ? tagInput.value || 'Unknown' : 'Unknown';
}

function parseDataString(rawData) {
    try {
        console.log('Parsing raw data:', rawData);
        
        // Clean the raw data
        const cleaned = rawData.trim();
        
        // Strategy 1: Space-separated format "DD.MM.YYYY HH:MM:SS value"
        const timeValueMatch = cleaned.match(/^(\d{2}\.\d{2}\.\d{4}\s+\d{2}:\d{2}:\d{2})\s+(.+)$/);
        if (timeValueMatch) {
            const [, timestamp, value] = timeValueMatch;
            console.log('Matched time-value format:', { timestamp, value });
            
            return {
                time: timestamp,
                tag: getCurrentTag(),
                value: formatValue(value),
                quality: 'OK',
                unit: getUnitForTag(getCurrentTag()),
                description: getDescriptionForTag(getCurrentTag())
            };
        }
        
        // Strategy 2: Формат времени ISO или UTC
        const isoTimeMatch = cleaned.match(/^(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}(?:\s+[+]\d{4}\s+UTC)?)\s+(.+)$/);
        if (isoTimeMatch) {
            const [, timestamp, value] = isoTimeMatch;
            console.log('Matched ISO time format:', { timestamp, value });
            
            // Конвертируем в нужный формат
            const date = new Date(timestamp.replace(/\s+[+]\d{4}\s+UTC/, ''));
            const formattedTime = formatTimestamp(date);
            
            return {
                time: formattedTime,
                tag: getCurrentTag(),
                value: formatValue(value),
                quality: 'OK',
                unit: getUnitForTag(getCurrentTag()),
                description: getDescriptionForTag(getCurrentTag())
            };
        }
        
        // Strategy 3: Format "10.02.2023 21:00:01 208.48"
        const parts = cleaned.split(/\s+/);
        console.log('Split parts:', parts);
        
        if (parts.length >= 3) {
            const date = parts[0];
            const time = parts[1];
            const value = parts[2];
            
            // Проверяем формат DD.MM.YYYY
            if (date.match(/^\d{2}\.\d{2}\.\d{4}$/) && time.match(/^\d{2}:\d{2}:\d{2}$/)) {
                const timestamp = `${date} ${time}`;
                console.log('Reconstructed timestamp:', timestamp);
                
                return {
                    time: timestamp,
                    tag: getCurrentTag(),
                    value: formatValue(value),
                    quality: 'OK',
                    unit: getUnitForTag(getCurrentTag()),
                    description: getDescriptionForTag(getCurrentTag())
                };
            }
        }
        
        // Strategy 4: Просто числовое значение
        const numericValue = parseFloat(cleaned);
        if (!isNaN(numericValue)) {
            console.log('Parsed as numeric value:', numericValue);
            
            return {
                time: formatTimestamp(new Date()),
                tag: getCurrentTag(),
                value: formatValue(numericValue),
                quality: 'OK',
                unit: getUnitForTag(getCurrentTag()),
                description: getDescriptionForTag(getCurrentTag())
            };
        }
        
        // Strategy 5: Пытаемся парсить как JSON
        try {
            const jsonData = JSON.parse(cleaned);
            console.log('Parsed as JSON:', jsonData);
            
            return {
                time: formatTimestamp(jsonData.timestamp || jsonData.time || jsonData.date || new Date()),
                tag: jsonData.tag || getCurrentTag(),
                value: formatValue(jsonData.value),
                quality: jsonData.quality || 'OK',
                unit: jsonData.unit || getUnitForTag(getCurrentTag()),
                description: jsonData.description || getDescriptionForTag(getCurrentTag())
            };
        } catch (jsonError) {
            // Не JSON, продолжаем
        }
        
    } catch (e) {
        console.warn('Could not parse data:', rawData, 'Error:', e);
    }
    
    // Fallback - возвращаем что-то разумное
    return {
        time: formatTimestamp(new Date()),
        tag: getCurrentTag(),
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
    // navigate to data page with selected tag
    const encodedTag = encodeURIComponent(tag);
    if (typeof window.loadPage === 'function') {
        window.loadPage(`/data/?tag=${encodedTag}`);
    }
}

export { 
    getTagOnDate,
    getTagList, 
    loadHomePageData,
    loadSwagger,
    updateDataTable,
    searchTagData,
    formatTimestamp,
    formatValue,
    getCurrentTag,
    getUnitForTag,
    getDescriptionForTag
}; 