import { showErrorNotification, showSuccessNotification } from './ui.js';

const CHART_COLORS = [
    '#2563eb',
    '#0891b2',
    '#059669',
    '#d97706',
    '#dc2626',
    '#7c3aed',
    '#0f766e',
    '#ea580c',
    '#db2777',
    '#65a30d',
    '#0284c7',
    '#4f46e5'
];

const PRESET_STORAGE_KEY = 'robin2:chart-presets:v1';
const DEFAULT_RENDER_TARGET = 1200;

const DOWNSAMPLE_INTERVAL_MS = {
    '30s': 30 * 1000,
    '1m': 60 * 1000,
    '5m': 5 * 60 * 1000,
    '15m': 15 * 60 * 1000,
    '1h': 60 * 60 * 1000
};

const chartState = {
    selectedTags: [],
    seriesByTag: new Map(),
    hiddenTags: new Set(),
    zoomX: 1,
    zoomY: 1,
    panX: 0,
    panY: 0,
    scaleMode: 'common',
    curveMode: 'linear',
    aggregationMode: 'raw',
    downsampleInterval: 'off',
    renderTargetPoints: DEFAULT_RENDER_TARGET,
    smoothingWindow: 1,
    showPoints: true,
    showValues: false,
    presets: []
};

let autoFetchTimer = null;
let autoFetchInFlight = false;
let lastAutoFetchSignature = '';
let hotkeysBound = false;

function initializeChartsPage() {
    const now = new Date();
    const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
    const dateFrom = document.getElementById('chartDateFrom');
    const dateTo = document.getElementById('chartDateTo');
    const likeInput = document.getElementById('chartLikeInput');

    if (dateFrom && !dateFrom.value) dateFrom.value = toDateTimeLocalValue(oneHourAgo);
    if (dateTo && !dateTo.value) dateTo.value = toDateTimeLocalValue(now);
    if (likeInput && !likeInput.value) likeInput.value = 'A20*';

    applyChartParamsFromURL();
    loadChartPresets();
    bindChartControlEvents();
    renderSelectedChartTags();
    renderPresetSelect();
    setupChartInteractions();
    updateViewLabels();
    renderTrendChart();

    const shouldAutoload = chartState.selectedTags.length > 0 && dateFrom && dateFrom.value && dateTo && dateTo.value;
    if (shouldAutoload) {
        setTimeout(() => {
            drawChartData();
        }, 0);
    }
}

function bindChartControlEvents() {
    const root = document.querySelector('.charts-workbench');
    if (!root || root.dataset.bound === '1') return;
    root.dataset.bound = '1';

    const tagInput = document.getElementById('chartTagInput');
    if (tagInput) {
        tagInput.addEventListener('keydown', (event) => {
            if (event.key === 'Enter') {
                event.preventDefault();
                addChartTagsFromInput();
            }
        });
    }

    bindRangeEvent('chartZoomRangeX', (value) => setChartZoomX(value));
    bindRangeEvent('chartZoomRangeY', (value) => setChartZoomY(value));

    bindRangeEvent('chartSmoothingRange', (value) => {
        chartState.smoothingWindow = clamp(Math.floor(value), 1, 25);
        updateSmoothingLabel();
        writeChartParamsToURL();
        renderTrendChart();
    });

    bindSelectEvent('chartScaleMode', (value) => {
        chartState.scaleMode = value === 'per_tag' ? 'per_tag' : 'common';
        writeChartParamsToURL();
        renderTrendChart();
    });

    bindSelectEvent('chartCurveMode', (value) => {
        chartState.curveMode = value === 'smooth' ? 'smooth' : 'linear';
        writeChartParamsToURL();
        renderTrendChart();
    });

    bindSelectEvent('chartAggregationMode', (value) => {
        chartState.aggregationMode = sanitizeAggregation(value);
        writeChartParamsToURL();
        renderTrendChart();
    });

    bindSelectEvent('chartDownsampleInterval', (value) => {
        chartState.downsampleInterval = sanitizeDownsampleInterval(value);
        writeChartParamsToURL();
        renderTrendChart();
    });

    bindInputEvent('chartRenderTargetPoints', (value) => {
        const numeric = Number.parseInt(value, 10);
        chartState.renderTargetPoints = clamp(Number.isFinite(numeric) ? numeric : DEFAULT_RENDER_TARGET, 100, 20000);
        writeChartParamsToURL();
        renderTrendChart();
    });

    bindCheckboxEvent('chartShowPoints', (checked) => {
        chartState.showPoints = checked;
        writeChartParamsToURL();
        renderTrendChart();
    });

    bindCheckboxEvent('chartShowValues', (checked) => {
        chartState.showValues = checked;
        writeChartParamsToURL();
        renderTrendChart();
    });

    const presetSelect = document.getElementById('chartPresetSelect');
    const presetName = document.getElementById('chartPresetName');
    if (presetSelect && presetName) {
        presetSelect.addEventListener('change', () => {
            if (!presetSelect.value) return;
            presetName.value = presetSelect.value;
        });
    }

    if (!hotkeysBound) {
        hotkeysBound = true;
        document.addEventListener('keydown', handleChartHotkeys);
    }
}

function bindRangeEvent(id, handler) {
    const element = document.getElementById(id);
    if (!element) return;
    element.addEventListener('input', (event) => {
        const value = Number(event.target.value);
        if (!Number.isFinite(value)) return;
        handler(value);
    });
}

function bindSelectEvent(id, handler) {
    const element = document.getElementById(id);
    if (!element) return;
    element.addEventListener('change', (event) => {
        handler(String(event.target.value || ''));
    });
}

function bindInputEvent(id, handler) {
    const element = document.getElementById(id);
    if (!element) return;
    element.addEventListener('change', (event) => {
        handler(String(event.target.value || ''));
    });
}

function bindCheckboxEvent(id, handler) {
    const element = document.getElementById(id);
    if (!element) return;
    element.addEventListener('change', (event) => {
        handler(Boolean(event.target.checked));
    });
}

function applyChartParamsFromURL() {
    const params = new URLSearchParams(window.location.search);

    const tagsParam = params.get('tags') || params.get('tag');
    if (tagsParam) {
        chartState.selectedTags = Array.from(
            new Set(
                tagsParam
                    .split(/[;,]+/)
                    .map((tag) => tag.trim())
                    .filter(Boolean)
            )
        );
    }

    const dateFrom = document.getElementById('chartDateFrom');
    const dateTo = document.getElementById('chartDateTo');
    const countInput = document.getElementById('chartCount');

    const from = params.get('from');
    const to = params.get('to');
    const count = params.get('count');

    if (dateFrom && from) dateFrom.value = from;
    if (dateTo && to) dateTo.value = to;
    if (countInput && count) countInput.value = count;

    chartState.scaleMode = params.get('scale') === 'per_tag' ? 'per_tag' : 'common';
    chartState.curveMode = params.get('curve') === 'smooth' ? 'smooth' : 'linear';
    chartState.aggregationMode = sanitizeAggregation(params.get('agg') || 'raw');
    chartState.downsampleInterval = sanitizeDownsampleInterval(params.get('ds') || 'off');
    chartState.zoomX = clamp(Number(params.get('zoomx') || '1') || 1, 0.01, 1000000);
    chartState.zoomY = clamp(Number(params.get('zoomy') || '1') || 1, 0.01, 1000000);

    const targetParam = Number.parseInt(params.get('target') || '', 10);
    chartState.renderTargetPoints = clamp(Number.isFinite(targetParam) ? targetParam : DEFAULT_RENDER_TARGET, 100, 20000);

    const smoothParam = Number.parseInt(params.get('smooth') || '', 10);
    chartState.smoothingWindow = clamp(Number.isFinite(smoothParam) ? smoothParam : 1, 1, 25);
    chartState.panX = Number(params.get('panx') || '0') || 0;
    chartState.panY = clamp(Number(params.get('pany') || '0') || 0, -1, 1);

    chartState.showPoints = parseBoolParam(params.get('points'), true);
    chartState.showValues = parseBoolParam(params.get('values'), false);
}

function writeChartParamsToURL() {
    const params = new URLSearchParams();
    const dateFrom = document.getElementById('chartDateFrom');
    const dateTo = document.getElementById('chartDateTo');
    const countInput = document.getElementById('chartCount');

    if (chartState.selectedTags.length) params.set('tags', chartState.selectedTags.join(','));
    if (dateFrom && dateFrom.value) params.set('from', dateFrom.value);
    if (dateTo && dateTo.value) params.set('to', dateTo.value);
    if (countInput && countInput.value) params.set('count', countInput.value);

    if (chartState.scaleMode !== 'common') params.set('scale', chartState.scaleMode);
    if (chartState.curveMode !== 'linear') params.set('curve', chartState.curveMode);
    if (chartState.aggregationMode !== 'raw') params.set('agg', chartState.aggregationMode);
    if (chartState.downsampleInterval !== 'off') params.set('ds', chartState.downsampleInterval);
    if (chartState.renderTargetPoints !== DEFAULT_RENDER_TARGET) params.set('target', String(chartState.renderTargetPoints));
    if (chartState.smoothingWindow !== 1) params.set('smooth', String(chartState.smoothingWindow));
    if (Math.abs(chartState.zoomX - 1) > 1e-9) params.set('zoomx', String(round3(chartState.zoomX)));
    if (Math.abs(chartState.zoomY - 1) > 1e-9) params.set('zoomy', String(round3(chartState.zoomY)));
    if (Math.abs(chartState.panX) > 1e-9) params.set('panx', String(round3(chartState.panX)));
    if (Math.abs(chartState.panY) > 1e-9) params.set('pany', String(round3(chartState.panY)));
    if (!chartState.showPoints) params.set('points', '0');
    if (chartState.showValues) params.set('values', '1');

    const query = params.toString();
    history.replaceState(null, '', query ? `/charts/?${query}` : '/charts/');
}

function addChartTagsFromInput() {
    const input = document.getElementById('chartTagInput');
    if (!input) return;

    const tags = String(input.value || '')
        .split(/[,\s;]+/)
        .map((tag) => tag.trim())
        .filter(Boolean);

    let added = 0;
    tags.forEach((tag) => {
        if (addChartTag(tag)) added += 1;
    });

    input.value = '';
    renderSelectedChartTags();
    if (added === 0 && tags.length > 0) showErrorNotification('Теги уже добавлены');
}

function addChartTag(tag) {
    const clean = String(tag || '').trim();
    if (!clean || chartState.selectedTags.includes(clean)) return false;

    chartState.selectedTags.push(clean);
    chartState.hiddenTags.delete(clean);
    writeChartParamsToURL();
    return true;
}

function removeChartTag(tag) {
    chartState.selectedTags = chartState.selectedTags.filter((item) => item !== tag);
    chartState.seriesByTag.delete(tag);
    chartState.hiddenTags.delete(tag);
    renderSelectedChartTags();
    renderTrendChart();
    writeChartParamsToURL();
}

function clearSelectedChartTags() {
    chartState.selectedTags = [];
    chartState.seriesByTag = new Map();
    chartState.hiddenTags = new Set();
    renderSelectedChartTags();
    renderTrendChart();
    writeChartParamsToURL();
}

function renderSelectedChartTags() {
    const containers = ['chartSelectedTags', 'chartSelectedTagsSecondary']
        .map((id) => document.getElementById(id))
        .filter(Boolean);
    if (!containers.length) return;

    if (chartState.selectedTags.length === 0) {
        containers.forEach((container) => {
            container.innerHTML = '<span class="tag-empty-hint">Теги не выбраны</span>';
        });
        return;
    }

    const markup = chartState.selectedTags
        .map(
            (tag, index) => `
        <span class="tag-chip" style="border-color:${CHART_COLORS[index % CHART_COLORS.length]}44;">
            <span class="tag-chip-dot" style="background:${CHART_COLORS[index % CHART_COLORS.length]};"></span>
            <span class="tag-chip-name">${escapeHTML(tag)}</span>
            <button type="button" class="tag-chip-remove" onclick="removeChartTag('${escapeJS(tag)}')">&times;</button>
        </span>
    `
        )
        .join('');

    containers.forEach((container) => {
        container.innerHTML = markup;
    });
}

async function findChartTagsByMask() {
    const input = document.getElementById('chartLikeInput');
    const apiElement = document.getElementById('apiserver');
    const container = document.getElementById('chartSuggestions');
    if (!input || !apiElement || !container) return;

    const api = apiElement.textContent.trim();
    const like = normalizeLikeMask(String(input.value || '').trim());
    const url = `${api}/get/tag/list/?format=json&like=${encodeURIComponent(like)}`;
    container.innerHTML = '<span class="tag-empty-hint">Загрузка...</span>';

    try {
        const response = await fetch(url);
        if (!response.ok) throw new Error(`HTTP ${response.status}`);

        const text = await response.text();
        let payload = text;
        try {
            payload = JSON.parse(text);
        } catch (_) {
            // keep text payload
        }

        const tags = extractTagNames(payload).slice(0, 500);
        renderTagSuggestions(tags);
    } catch (error) {
        container.innerHTML = `<span class="tag-empty-hint">Ошибка: ${escapeHTML(error.message)}</span>`;
        showErrorNotification(`Ошибка поиска тегов: ${error.message}`);
    }
}

function normalizeLikeMask(rawLike) {
    if (!rawLike) return '%';
    let like = rawLike.replaceAll('*', '%').replaceAll('?', '_');
    if (!like.includes('%') && !like.includes('_')) like += '%';
    return like;
}

function extractTagNames(payload) {
    const tags = [];

    if (typeof payload === 'string') {
        const text = payload.trim();
        if (!text || text.startsWith('#Error')) return [];
        if (!text.includes('\n') && !text.includes('\t') && !text.includes('{')) return [text];
        text.split('\n').forEach((line) => {
            const value = line.trim().split(/\s+/)[0];
            if (value) tags.push(value);
        });
        return Array.from(new Set(tags));
    }

    if (Array.isArray(payload)) {
        payload.forEach((item) => {
            if (typeof item === 'string') tags.push(item);
            else if (Array.isArray(item) && item[0]) tags.push(String(item[0]));
            else if (item && typeof item === 'object') {
                if (item.tag) tags.push(String(item.tag));
                if (item.name) tags.push(String(item.name));
            }
        });
        return Array.from(new Set(tags));
    }

    if (payload && typeof payload === 'object') {
        if (Array.isArray(payload.rows)) {
            payload.rows.forEach((row) => {
                if (Array.isArray(row) && row[0]) tags.push(String(row[0]));
            });
        }
        if (Array.isArray(payload.tags)) {
            payload.tags.forEach((tag) => tags.push(String(tag)));
        }
    }

    return Array.from(new Set(tags));
}

function renderTagSuggestions(tags) {
    const container = document.getElementById('chartSuggestions');
    if (!container) return;

    if (!tags.length) {
        container.innerHTML = '<span class="tag-empty-hint">Теги не найдены</span>';
        return;
    }

    container.innerHTML = tags
        .map(
            (tag) => `
        <button type="button" class="tag-suggestion-btn" onclick="addSuggestedChartTag('${escapeJS(tag)}')">
            <span class="tag-suggestion-name">${escapeHTML(tag)}</span>
            <span class="tag-suggestion-action">+ Добавить</span>
        </button>
    `
        )
        .join('');
}

function addSuggestedChartTag(tag) {
    if (addChartTag(tag)) renderSelectedChartTags();
}

async function drawChartData(options = {}) {
    const silent = Boolean(options.silent);
    const apiElement = document.getElementById('apiserver');
    const dateFrom = document.getElementById('chartDateFrom');
    const dateTo = document.getElementById('chartDateTo');
    const countInput = document.getElementById('chartCount');
    const drawBtn = document.getElementById('drawChartBtn');

    if (!apiElement || !dateFrom || !dateTo || !countInput) return;
    if (chartState.selectedTags.length === 0) {
        showErrorNotification('Добавьте минимум один тег');
        return;
    }
    if (!dateFrom.value || !dateTo.value) {
        showErrorNotification('Укажите период');
        return;
    }

    const count = Number(countInput.value || '300');
    if (!Number.isFinite(count) || count < 1) {
        showErrorNotification('Количество точек должно быть больше 0');
        return;
    }

    const fromDate = new Date(dateFrom.value);
    const toDate = new Date(dateTo.value);
    if (Number.isNaN(fromDate.getTime()) || Number.isNaN(toDate.getTime()) || fromDate >= toDate) {
        showErrorNotification('Некорректный период');
        return;
    }

    writeChartParamsToURL();
    if (drawBtn) drawBtn.disabled = true;

    try {
        const api = apiElement.textContent.trim();
        const params = new URLSearchParams({
            tag: chartState.selectedTags.join(','),
            from: convertDateTimeLocal(dateFrom.value),
            to: convertDateTimeLocal(dateTo.value),
            format: 'json',
            count: String(Math.floor(count))
        });

        const response = await fetch(`${api}/get/tag/?${params.toString()}`);
        if (!response.ok) throw new Error(`HTTP ${response.status}`);

        const text = await response.text();
        if (text.startsWith('#Error:')) throw new Error(text);

        let payload;
        try {
            payload = JSON.parse(text);
        } catch (_) {
            throw new Error('API вернул невалидный JSON');
        }

        chartState.seriesByTag = normalizeSeriesResponse(payload, chartState.selectedTags);
        chartState.hiddenTags.forEach((tag) => {
            if (!chartState.selectedTags.includes(tag)) chartState.hiddenTags.delete(tag);
        });

        renderTrendChart();
        const totalRawPoints = Array.from(chartState.seriesByTag.values()).reduce((sum, points) => sum + points.length, 0);
        if (!silent) {
            showSuccessNotification(`График построен. Сырых точек: ${totalRawPoints}`);
        }
    } catch (error) {
        chartState.seriesByTag = new Map();
        renderTrendChart();
        if (!silent) {
            showErrorNotification(`Ошибка построения графика: ${error.message}`);
        }
    } finally {
        if (drawBtn) drawBtn.disabled = false;
    }
}

function normalizeSeriesResponse(payload, fallbackTags) {
    const seriesByTag = new Map();
    if (!payload || typeof payload !== 'object') return seriesByTag;

    const entries = Object.entries(payload);
    const nested = entries.some(([, value]) => value && typeof value === 'object' && !Array.isArray(value));

    if (nested) {
        entries.forEach(([tag, value]) => {
            if (!value || typeof value !== 'object' || Array.isArray(value)) return;
            const points = parseSeriesObject(value);
            if (points.length) seriesByTag.set(tag, points);
        });
        return seriesByTag;
    }

    if (fallbackTags.length === 1) {
        const points = parseSeriesObject(payload);
        if (points.length) seriesByTag.set(fallbackTags[0], points);
    }
    return seriesByTag;
}

function parseSeriesObject(seriesObj) {
    const points = [];
    Object.entries(seriesObj).forEach(([timeRaw, valueRaw]) => {
        const timestamp = parseTimestamp(timeRaw);
        const value = parseNumeric(valueRaw);
        if (!timestamp || !Number.isFinite(value)) return;
        points.push({ timestamp, value });
    });
    points.sort((a, b) => a.timestamp - b.timestamp);
    return points;
}

function parseTimestamp(value) {
    if (value instanceof Date) return Number.isNaN(value.getTime()) ? null : value;
    const str = String(value || '').trim();
    if (!str) return null;

    let date = new Date(str.replace(' ', 'T'));
    if (Number.isNaN(date.getTime())) date = new Date(str);
    return Number.isNaN(date.getTime()) ? null : date;
}

function parseNumeric(value) {
    if (typeof value === 'number') return value;
    const parsed = Number.parseFloat(String(value || '').replace(',', '.').replace(/\s+/g, ''));
    return Number.isFinite(parsed) ? parsed : NaN;
}

function getLoadedTimeBounds() {
    const allPoints = Array.from(chartState.seriesByTag.values()).flat();
    if (!allPoints.length) return null;
    const minTime = Math.min(...allPoints.map((point) => point.timestamp.getTime()));
    const maxTime = Math.max(...allPoints.map((point) => point.timestamp.getTime()));
    return { minTime, maxTime };
}

function getViewWindow(minTime, maxTime) {
    const view = getViewWindow(minTime, maxTime);
    const viewMin = view.min;
    const viewMax = view.max;
    return { min: viewMin, max: viewMax, span: visibleSpan, totalSpan: timeSpan };
}

function estimateVisiblePoints(viewMin, viewMax) {
    let total = 0;
    let seriesCount = 0;
    chartState.selectedTags.forEach((tag) => {
        const points = chartState.seriesByTag.get(tag) || [];
        if (points.length === 0) return;
        seriesCount += 1;
        for (let index = 0; index < points.length; index += 1) {
            const ts = points[index].timestamp.getTime();
            if (ts >= viewMin && ts <= viewMax) total += 1;
        }
    });

    return seriesCount > 0 ? Math.round(total / seriesCount) : 0;
}

function scheduleAutoDataLoad(reason) {
    if (autoFetchInFlight) return;
    if (autoFetchTimer) clearTimeout(autoFetchTimer);
    autoFetchTimer = setTimeout(() => {
        ensureDataCoverage(reason).catch(() => {
            // fail silently for background auto-load
        });
    }, 260);
}

async function ensureDataCoverage(reason) {
    if (autoFetchInFlight) return;
    if (chartState.selectedTags.length === 0) return;

    const bounds = getLoadedTimeBounds();
    if (!bounds) return;

    const dateFrom = document.getElementById('chartDateFrom');
    const dateTo = document.getElementById('chartDateTo');
    const countInput = document.getElementById('chartCount');
    if (!dateFrom || !dateTo || !countInput) return;

    const view = getViewWindow(bounds.minTime, bounds.maxTime);
    const margin = Math.max(30 * 1000, view.span * 0.15);

    let desiredFrom = bounds.minTime;
    let desiredTo = bounds.maxTime;
    let desiredCount = Number.parseInt(countInput.value || '300', 10);
    if (!Number.isFinite(desiredCount) || desiredCount < 10) desiredCount = 300;
    let needFetch = false;

    if (view.min > bounds.maxTime || view.max < bounds.minTime) {
        desiredFrom = view.min - view.span * 0.6;
        desiredTo = view.max + view.span * 0.6;
        needFetch = true;
    }

    if (view.min < bounds.minTime + margin) {
        desiredFrom = Math.min(desiredFrom, view.min - view.span * 0.5);
        needFetch = true;
    }
    if (view.max > bounds.maxTime - margin) {
        desiredTo = Math.max(desiredTo, view.max + view.span * 0.5);
        needFetch = true;
    }

    const visiblePoints = estimateVisiblePoints(view.min, view.max);
    const targetVisible = clamp(Math.round(chartState.renderTargetPoints * 0.6), 250, 8000);
    if (chartState.zoomX > 1.3 && visiblePoints > 0 && visiblePoints < Math.round(targetVisible * 0.4)) {
        desiredCount = clamp(Math.max(desiredCount, Math.round(desiredCount * 1.6), targetVisible * 2), 10, 30000);
        needFetch = true;
    }

    if (!needFetch) return;

    const nextFrom = new Date(desiredFrom);
    const nextTo = new Date(desiredTo);
    if (!(nextFrom < nextTo)) return;

    const signature = `${toDateTimeLocalValue(nextFrom)}|${toDateTimeLocalValue(nextTo)}|${desiredCount}`;
    if (signature === lastAutoFetchSignature) return;
    lastAutoFetchSignature = signature;

    dateFrom.value = toDateTimeLocalValue(nextFrom);
    dateTo.value = toDateTimeLocalValue(nextTo);
    countInput.value = String(desiredCount);

    writeChartParamsToURL();
    autoFetchInFlight = true;
    try {
        await drawChartData({ silent: true });
    } finally {
        autoFetchInFlight = false;
    }
}

function renderTrendChart() {
    const svg = document.getElementById('trendChartSvg');
    const viewport = document.getElementById('chartViewport');
    const legend = document.getElementById('chartLegend');
    const emptyState = document.getElementById('chartEmptyState');
    if (!svg || !viewport || !legend || !emptyState) return;

    const processedSeries = new Map();
    chartState.selectedTags.forEach((tag) => {
        const rawPoints = chartState.seriesByTag.get(tag) || [];
        processedSeries.set(tag, processSeriesForRender(rawPoints));
    });

    const visibleSeries = chartState.selectedTags
        .filter((tag) => !chartState.hiddenTags.has(tag))
        .map((tag) => [tag, processedSeries.get(tag) || []])
        .filter(([, points]) => points.length > 0);

    renderChartLegend(processedSeries);
    renderChartMeta(processedSeries);

    if (!visibleSeries.length) {
        svg.innerHTML = '';
        svg.setAttribute('width', '100%');
        svg.setAttribute('height', '520');
        emptyState.style.display = 'flex';
        return;
    }

    emptyState.style.display = 'none';

    const allPoints = visibleSeries.flatMap(([, points]) => points);
    const minTime = Math.min(...allPoints.map((point) => point.timestamp.getTime()));
    const maxTime = Math.max(...allPoints.map((point) => point.timestamp.getTime()));
    const visibleWidth = Math.max(1, viewport.clientWidth - 2);
    const width = visibleWidth;
    const height = 520;
    const padLeft = 82;
    const padRight = chartState.scaleMode === 'per_tag' ? 190 : 44;
    const padTop = 26;
    const padBottom = 76;
    const plotWidth = Math.max(1, width - padLeft - padRight);
    const plotHeight = Math.max(1, height - padTop - padBottom);

    const timeSpan = Math.max(1, maxTime - minTime);
    const baseCenter = minTime + timeSpan / 2;
    const safeZoomX = Math.max(1e-6, chartState.zoomX);
    const visibleSpan = timeSpan / safeZoomX;
    const viewMin = baseCenter + chartState.panX * timeSpan - visibleSpan / 2;
    const viewMax = viewMin + visibleSpan;

    const x = (timestamp) => {
        const denominator = Math.max(1, viewMax - viewMin);
        return padLeft + ((timestamp - viewMin) / denominator) * plotWidth;
    };

    const rangesByTag = new Map();
    const commonRange = computeRange(visibleSeries.flatMap(([, points]) => points.map((point) => point.value)));
    visibleSeries.forEach(([tag, points]) => {
        rangesByTag.set(tag, computeRange(points.map((point) => point.value)));
    });

    const yCommon = buildYMapper(commonRange, chartState.zoomY, chartState.panY, padTop, plotHeight);
    const yForTag = (tag) => {
        if (chartState.scaleMode === 'per_tag') {
            return buildYMapper(rangesByTag.get(tag) || commonRange, chartState.zoomY, chartState.panY, padTop, plotHeight);
        }
        return yCommon;
    };

    const grid = [];
    const yLabels = [];
    const yTicks = 6;
    for (let index = 0; index <= yTicks; index += 1) {
        const ratio = index / yTicks;
        const yLine = padTop + ratio * plotHeight;
        grid.push(
            `<line x1="${padLeft}" y1="${yLine.toFixed(2)}" x2="${(padLeft + plotWidth).toFixed(2)}" y2="${yLine.toFixed(
                2
            )}" class="chart-grid-line" />`
        );

        if (chartState.scaleMode === 'common') {
            const value = yCommon.min + (1 - ratio) * (yCommon.max - yCommon.min);
            yLabels.push(
                `<text x="${padLeft - 10}" y="${(yLine + 4).toFixed(2)}" class="chart-axis-label" text-anchor="end">${formatAxisValue(
                    value
                )}</text>`
            );
        } else {
            yLabels.push(
                `<text x="${padLeft - 10}" y="${(yLine + 4).toFixed(2)}" class="chart-axis-label" text-anchor="end">${Math.round(
                    (1 - ratio) * 100
                )}%</text>`
            );
        }
    }

    const xLabels = [];
    const xTicks = Math.max(4, Math.min(14, Math.floor(width / 170)));
    for (let index = 0; index <= xTicks; index += 1) {
        const ratio = index / xTicks;
        const xLine = padLeft + ratio * plotWidth;
        const timestamp = viewMin + ratio * (viewMax - viewMin);
        xLabels.push(
            `<line x1="${xLine.toFixed(2)}" y1="${padTop}" x2="${xLine.toFixed(2)}" y2="${(padTop + plotHeight).toFixed(
                2
            )}" class="chart-grid-line chart-grid-line-vertical" />`
        );
        xLabels.push(
            `<text x="${xLine.toFixed(2)}" y="${(padTop + plotHeight + 24).toFixed(2)}" class="chart-axis-label" text-anchor="middle">${formatTimeLabel(
                timestamp
            )}</text>`
        );
    }

    const axisX = `<line x1="${padLeft}" y1="${(padTop + plotHeight).toFixed(2)}" x2="${(padLeft + plotWidth).toFixed(
        2
    )}" y2="${(padTop + plotHeight).toFixed(2)}" class="chart-axis-line" />`;
    const axisY = `<line x1="${padLeft}" y1="${padTop}" x2="${padLeft}" y2="${(padTop + plotHeight).toFixed(
        2
    )}" class="chart-axis-line" />`;

    const seriesShapes = [];
    const sideLabels = [];

    visibleSeries.forEach(([tag, points]) => {
        const color = colorForTag(tag);
        const yMapper = yForTag(tag);
        const xyPoints = points.map((point) => ({
            x: x(point.timestamp.getTime()),
            y: yMapper.toY(point.value),
            timestamp: point.timestamp,
            value: point.value
        }));

        const path = buildLinePath(xyPoints, chartState.curveMode);
        seriesShapes.push(
            `<path d="${path}" fill="none" stroke="${color}" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" />`
        );

        if (chartState.showPoints && points.length <= 2500) {
            seriesShapes.push(
                xyPoints
                    .map((point) => {
                        const title = `${tag}\n${formatTimeFull(point.timestamp)}\n${formatAxisValue(point.value)}`;
                        return `<circle cx="${point.x.toFixed(2)}" cy="${point.y.toFixed(2)}" r="2.2" fill="${color}" class="chart-point"><title>${escapeHTML(
                            title
                        )}</title></circle>`;
                    })
                    .join('')
            );
        }

        if (chartState.showValues) {
            const step = Math.max(1, Math.floor(xyPoints.length / Math.max(12, Math.floor(plotWidth / 95))));
            const labels = [];
            for (let index = 0; index < xyPoints.length; index += step) {
                const point = xyPoints[index];
                labels.push(
                    `<text x="${(point.x + 4).toFixed(2)}" y="${(point.y - 6).toFixed(2)}" class="chart-value-label" fill="${color}">${formatAxisValue(
                        point.value
                    )}</text>`
                );
            }
            const last = xyPoints[xyPoints.length - 1];
            if (last) {
                labels.push(
                    `<text x="${(last.x + 6).toFixed(2)}" y="${(last.y - 8).toFixed(2)}" class="chart-value-label chart-value-label-last" fill="${color}">${formatAxisValue(
                        last.value
                    )}</text>`
                );
            }
            seriesShapes.push(labels.join(''));
        }

        if (chartState.scaleMode === 'per_tag') {
            const last = xyPoints[xyPoints.length - 1];
            if (last) {
                const xStart = padLeft + plotWidth + 10;
                sideLabels.push(`
                    <line x1="${(padLeft + plotWidth).toFixed(2)}" y1="${last.y.toFixed(2)}" x2="${(xStart - 3).toFixed(
                    2
                )}" y2="${last.y.toFixed(2)}" stroke="${color}" stroke-width="1.4" />
                    <text x="${xStart}" y="${(last.y - 2).toFixed(2)}" class="chart-side-label" fill="${color}">
                        ${escapeHTML(tag)}: ${formatAxisValue(last.value)}
                    </text>
                `);
            }
        }
    });

    svg.setAttribute('width', String(width));
    svg.setAttribute('height', String(height));
    svg.setAttribute('viewBox', `0 0 ${width} ${height}`);
    const clipId = 'chartPlotClip';
    svg.innerHTML = `
        <defs>
            <clipPath id="${clipId}">
                <rect x="${padLeft}" y="${padTop}" width="${plotWidth}" height="${plotHeight}"></rect>
            </clipPath>
        </defs>
        <rect x="0" y="0" width="${width}" height="${height}" class="chart-bg"></rect>
        <g>${grid.join('')}</g>
        <g>${xLabels.join('')}</g>
        <g>${yLabels.join('')}</g>
        <g>${axisX}${axisY}</g>
        <g clip-path="url(#${clipId})">${seriesShapes.join('')}</g>
        <g>${sideLabels.join('')}</g>
    `;
}

function processSeriesForRender(points) {
    if (!Array.isArray(points) || points.length === 0) return [];

    const target = clamp(chartState.renderTargetPoints, 100, 20000);
    const bucketMs = resolveDownsampleMs(points, target);
    const aggregation = resolveAggregationMode(bucketMs);

    let processed = aggregateByBucket(points, bucketMs, aggregation);
    if (bucketMs <= 0 && processed.length > target * 2) {
        processed = decimateByStride(processed, target);
    }

    return applySmoothing(processed, chartState.smoothingWindow);
}

function resolveDownsampleMs(points, targetPoints) {
    if (!Array.isArray(points) || points.length < 2) return 0;

    const mode = chartState.downsampleInterval;
    if (mode in DOWNSAMPLE_INTERVAL_MS) return DOWNSAMPLE_INTERVAL_MS[mode];

    const minTime = points[0].timestamp.getTime();
    const maxTime = points[points.length - 1].timestamp.getTime();
    const span = Math.max(1, maxTime - minTime);

    if (mode === 'auto' || (mode === 'off' && chartState.aggregationMode !== 'raw')) {
        return normalizeBucketMs(Math.ceil(span / Math.max(2, targetPoints)));
    }

    return 0;
}

function normalizeBucketMs(rawMs) {
    const buckets = [
        1000,
        2000,
        5000,
        10 * 1000,
        15 * 1000,
        30 * 1000,
        60 * 1000,
        2 * 60 * 1000,
        5 * 60 * 1000,
        10 * 60 * 1000,
        15 * 60 * 1000,
        30 * 60 * 1000,
        60 * 60 * 1000,
        2 * 60 * 60 * 1000,
        6 * 60 * 60 * 1000,
        12 * 60 * 60 * 1000,
        24 * 60 * 60 * 1000
    ];

    for (let index = 0; index < buckets.length; index += 1) {
        if (rawMs <= buckets[index]) return buckets[index];
    }
    return buckets[buckets.length - 1];
}

function resolveAggregationMode(bucketMs) {
    if (bucketMs <= 0) return 'raw';
    return chartState.aggregationMode === 'raw' ? 'avg' : chartState.aggregationMode;
}

function aggregateByBucket(points, bucketMs, mode) {
    if (!Array.isArray(points) || points.length === 0) return [];
    if (bucketMs <= 0 || mode === 'raw') return points.slice();

    const result = [];
    let currentBucket = null;
    let stats = null;

    const flush = () => {
        if (!stats) return;
        const timestamp = new Date(stats.tsSum / Math.max(1, stats.count));
        let value = stats.last;
        if (mode === 'avg') value = stats.sum / Math.max(1, stats.count);
        else if (mode === 'min') value = stats.min;
        else if (mode === 'max') value = stats.max;
        result.push({ timestamp, value });
    };

    points.forEach((point) => {
        const ts = point.timestamp.getTime();
        const bucket = Math.floor(ts / bucketMs);

        if (currentBucket === null || bucket !== currentBucket) {
            flush();
            currentBucket = bucket;
            stats = {
                count: 0,
                sum: 0,
                min: Number.POSITIVE_INFINITY,
                max: Number.NEGATIVE_INFINITY,
                last: point.value,
                tsSum: 0
            };
        }

        stats.count += 1;
        stats.sum += point.value;
        stats.min = Math.min(stats.min, point.value);
        stats.max = Math.max(stats.max, point.value);
        stats.last = point.value;
        stats.tsSum += ts;
    });

    flush();
    return result;
}

function decimateByStride(points, target) {
    if (!Array.isArray(points) || points.length <= target) return points.slice();
    const stride = Math.max(2, Math.ceil(points.length / target));
    const decimated = [points[0]];

    for (let index = stride; index < points.length - 1; index += stride) {
        decimated.push(points[index]);
    }

    const last = points[points.length - 1];
    if (decimated[decimated.length - 1] !== last) decimated.push(last);
    return decimated;
}

function applySmoothing(points, windowSize) {
    if (!Array.isArray(points) || points.length < 3 || windowSize <= 1) return points.slice();

    const half = Math.floor(windowSize / 2);
    return points.map((point, index) => {
        const start = Math.max(0, index - half);
        const end = Math.min(points.length - 1, index + half);
        let sum = 0;
        let count = 0;
        for (let pos = start; pos <= end; pos += 1) {
            sum += points[pos].value;
            count += 1;
        }
        return {
            timestamp: point.timestamp,
            value: count > 0 ? sum / count : point.value
        };
    });
}

function renderChartLegend(processedSeries) {
    const legend = document.getElementById('chartLegend');
    if (!legend) return;
    if (chartState.selectedTags.length === 0) {
        legend.innerHTML = '';
        return;
    }

    legend.innerHTML = chartState.selectedTags
        .map((tag) => {
            const points = processedSeries.get(tag) || [];
            const color = colorForTag(tag);
            const hidden = chartState.hiddenTags.has(tag);

            if (points.length === 0) {
                return `
                <button type="button" class="chart-legend-item chart-legend-item-empty" onclick="toggleSeriesVisibility('${escapeJS(tag)}')">
                    <span class="chart-legend-color" style="background:${color};"></span>
                    <span class="chart-legend-name">${escapeHTML(tag)}</span>
                    <span class="chart-legend-value">нет данных</span>
                </button>
            `;
            }

            const values = points.map((point) => point.value);
            const min = Math.min(...values);
            const max = Math.max(...values);
            const last = values[values.length - 1];
            return `
            <button type="button" class="chart-legend-item ${hidden ? 'chart-legend-item-muted' : ''}" onclick="toggleSeriesVisibility('${escapeJS(
                tag
            )}')">
                <span class="chart-legend-color" style="background:${color};"></span>
                <span class="chart-legend-name">${escapeHTML(tag)}</span>
                <span class="chart-legend-value">${formatAxisValue(last)}</span>
                <span class="chart-legend-range">${formatAxisValue(min)} .. ${formatAxisValue(max)}</span>
                <span class="chart-legend-toggle">${hidden ? 'show' : 'hide'}</span>
            </button>
        `;
        })
        .join('');
}

function renderChartMeta(processedSeries) {
    const meta = document.getElementById('chartRenderMeta');
    if (!meta) return;

    const rawPoints = Array.from(chartState.seriesByTag.values()).reduce((sum, points) => sum + points.length, 0);
    const renderedPoints = Array.from(processedSeries.values()).reduce((sum, points) => sum + points.length, 0);

    const downsampleLabel =
        chartState.downsampleInterval === 'off'
            ? 'off'
            : chartState.downsampleInterval === 'auto'
              ? 'auto'
              : chartState.downsampleInterval;

    meta.textContent = `Raw: ${rawPoints} | Rendered: ${renderedPoints} | Agg: ${chartState.aggregationMode} | Downsample: ${downsampleLabel} | Smoothing: ${chartState.smoothingWindow}`;
}

function toggleSeriesVisibility(tag) {
    if (chartState.hiddenTags.has(tag)) chartState.hiddenTags.delete(tag);
    else chartState.hiddenTags.add(tag);
    renderTrendChart();
}

function computeRange(values) {
    const minRaw = Math.min(...values);
    const maxRaw = Math.max(...values);
    const spanRaw = maxRaw - minRaw;
    const padding = spanRaw > 0 ? spanRaw * 0.12 : Math.max(1, Math.abs(maxRaw) * 0.2);
    return { min: minRaw - padding, max: maxRaw + padding };
}

function buildYMapper(baseRange, zoomY, panY, padTop, plotHeight) {
    const baseSpan = Math.max(1e-9, baseRange.max - baseRange.min);
    const zoomedSpan = baseSpan / Math.max(1, zoomY);
    const baseCenter = (baseRange.min + baseRange.max) / 2;
    const maxShift = (baseSpan - zoomedSpan) / 2;
    const shift = clamp(panY, -1, 1) * maxShift;
    const min = baseCenter + shift - zoomedSpan / 2;
    const max = baseCenter + shift + zoomedSpan / 2;

    return {
        min,
        max,
        toY: (value) => padTop + (1 - (value - min) / Math.max(1e-9, max - min)) * plotHeight
    };
}

function buildLinePath(points, curveMode) {
    if (!points.length) return '';
    if (points.length === 1) return `M ${points[0].x.toFixed(2)} ${points[0].y.toFixed(2)}`;
    if (curveMode !== 'smooth' || points.length < 3) {
        return points.map((point, index) => `${index === 0 ? 'M' : 'L'} ${point.x.toFixed(2)} ${point.y.toFixed(2)}`).join(' ');
    }

    let path = `M ${points[0].x.toFixed(2)} ${points[0].y.toFixed(2)}`;
    for (let index = 1; index < points.length - 1; index += 1) {
        const xc = (points[index].x + points[index + 1].x) / 2;
        const yc = (points[index].y + points[index + 1].y) / 2;
        path += ` Q ${points[index].x.toFixed(2)} ${points[index].y.toFixed(2)} ${xc.toFixed(2)} ${yc.toFixed(2)}`;
    }
    const last = points[points.length - 1];
    path += ` T ${last.x.toFixed(2)} ${last.y.toFixed(2)}`;
    return path;
}

function setupChartInteractions() {
    const viewport = document.getElementById('chartViewport');
    if (!viewport || viewport.dataset.interactionsBound === '1') return;
    viewport.dataset.interactionsBound = '1';

    let dragMode = null;
    let startX = 0;
    let startY = 0;
    let startPanX = 0;
    let startPanY = 0;

    viewport.addEventListener('contextmenu', (event) => event.preventDefault());

    viewport.addEventListener('pointerdown', (event) => {
        if (event.pointerType === 'mouse' && event.button !== 0) return;
        dragMode = event.shiftKey ? 'y' : 'x';
        startX = event.clientX;
        startY = event.clientY;
        startPanX = chartState.panX;
        startPanY = chartState.panY;
        viewport.setPointerCapture(event.pointerId);
        viewport.classList.add('dragging');
    });

    viewport.addEventListener('pointermove', (event) => {
        if (!dragMode) return;
        if (dragMode === 'x') {
            const dx = event.clientX - startX;
            const plotWidth = Math.max(1, viewport.clientWidth - 120);
            chartState.panX = startPanX - dx / Math.max(1, plotWidth * Math.max(chartState.zoomX, 1e-6));
            renderTrendChart();
        } else {
            const delta = event.clientY - startY;
            chartState.panY = clamp(startPanY + delta / 280, -1, 1);
            renderTrendChart();
        }
    });

    const stopDragging = () => {
        if (dragMode) {
            writeChartParamsToURL();
            scheduleAutoDataLoad('drag');
        }
        dragMode = null;
        viewport.classList.remove('dragging');
    };
    viewport.addEventListener('pointerup', stopDragging);
    viewport.addEventListener('pointercancel', stopDragging);

    viewport.addEventListener(
        'wheel',
        (event) => {
            if (event.ctrlKey) {
                event.preventDefault();
                setChartZoomX(chartState.zoomX * (event.deltaY > 0 ? 0.9 : 1.1));
                return;
            }
            if (event.altKey) {
                event.preventDefault();
                setChartZoomY(chartState.zoomY * (event.deltaY > 0 ? 0.92 : 1.08));
                return;
            }
            if (event.shiftKey) {
                event.preventDefault();
                chartState.panY = clamp(chartState.panY + event.deltaY / 800, -1, 1);
                writeChartParamsToURL();
                renderTrendChart();
                scheduleAutoDataLoad('pan-y');
                return;
            }
            event.preventDefault();
            chartState.panX += event.deltaY / (1200 * Math.max(chartState.zoomX, 1e-6));
            writeChartParamsToURL();
            renderTrendChart();
            scheduleAutoDataLoad('pan-x-wheel');
        },
        { passive: false }
    );
}

function setChartZoomX(nextValue) {
    const safeValue = Number.isFinite(nextValue) ? nextValue : chartState.zoomX;
    chartState.zoomX = clamp(round3(safeValue), 0.01, 1000000);
    const range = document.getElementById('chartZoomRangeX');
    if (range) range.value = String(chartState.zoomX);
    writeChartParamsToURL();
    updateZoomLabelX();
    renderTrendChart();
    scheduleAutoDataLoad('zoom-x');
}

function setChartZoomY(nextValue) {
    const safeValue = Number.isFinite(nextValue) ? nextValue : chartState.zoomY;
    chartState.zoomY = clamp(round3(safeValue), 0.01, 1000000);
    const range = document.getElementById('chartZoomRangeY');
    if (range) range.value = String(chartState.zoomY);
    writeChartParamsToURL();
    updateZoomLabelY();
    renderTrendChart();
    scheduleAutoDataLoad('zoom-y');
}

function zoomInChartX() {
    setChartZoomX(chartState.zoomX * 1.15);
}

function zoomOutChartX() {
    setChartZoomX(chartState.zoomX / 1.15);
}

function zoomInChartY() {
    setChartZoomY(chartState.zoomY * 1.12);
}

function zoomOutChartY() {
    setChartZoomY(chartState.zoomY / 1.12);
}

function resetChartView() {
    fitAllChartData();
}

function fitAllChartData() {
    chartState.zoomX = 1;
    chartState.zoomY = 1;
    chartState.panX = 0;
    chartState.panY = 0;

    const rangeX = document.getElementById('chartZoomRangeX');
    const rangeY = document.getElementById('chartZoomRangeY');
    if (rangeX) rangeX.value = '1';
    if (rangeY) rangeY.value = '1';
    writeChartParamsToURL();
    updateZoomLabelX();
    updateZoomLabelY();
    renderTrendChart();
}

function updateViewLabels() {
    updateZoomLabelX();
    updateZoomLabelY();
    updateSmoothingLabel();

    setSelectValue('chartScaleMode', chartState.scaleMode);
    setSelectValue('chartCurveMode', chartState.curveMode);
    setSelectValue('chartAggregationMode', chartState.aggregationMode);
    setSelectValue('chartDownsampleInterval', chartState.downsampleInterval);

    const showPoints = document.getElementById('chartShowPoints');
    if (showPoints) showPoints.checked = chartState.showPoints;
    const showValues = document.getElementById('chartShowValues');
    if (showValues) showValues.checked = chartState.showValues;

    const smoothingRange = document.getElementById('chartSmoothingRange');
    if (smoothingRange) smoothingRange.value = String(chartState.smoothingWindow);

    const renderTarget = document.getElementById('chartRenderTargetPoints');
    if (renderTarget) renderTarget.value = String(chartState.renderTargetPoints);
}

function updateZoomLabelX() {
    const label = document.getElementById('chartZoomLabelX');
    if (label) label.textContent = `${round3(chartState.zoomX)}x`;
}

function updateZoomLabelY() {
    const label = document.getElementById('chartZoomLabelY');
    if (label) label.textContent = `${round3(chartState.zoomY)}x`;
}

function updateSmoothingLabel() {
    const label = document.getElementById('chartSmoothingLabel');
    if (label) label.textContent = String(chartState.smoothingWindow);
}

function handleChartHotkeys(event) {
    if (!window.location.pathname.includes('/charts')) return;
    if (isEditableTarget(event.target)) return;
    if (event.defaultPrevented) return;
    if (event.ctrlKey || event.metaKey || event.altKey) return;

    if (event.key === 'f' || event.key === 'F') {
        event.preventDefault();
        fitAllChartData();
        return;
    }
    if (event.key === 'r' || event.key === 'R') {
        event.preventDefault();
        resetChartView();
        return;
    }
    if (event.key === '[') {
        event.preventDefault();
        shiftFetchRangeLeft();
        return;
    }
    if (event.key === ']') {
        event.preventDefault();
        shiftFetchRangeRight();
        return;
    }
    if (event.key === '+' || event.key === '=') {
        event.preventDefault();
        zoomInChartX();
        return;
    }
    if (event.key === '-' || event.key === '_') {
        event.preventDefault();
        zoomOutChartX();
        return;
    }
    if (event.shiftKey && event.key === 'ArrowUp') {
        event.preventDefault();
        zoomInChartY();
        return;
    }
    if (event.shiftKey && event.key === 'ArrowDown') {
        event.preventDefault();
        zoomOutChartY();
        return;
    }
}

function isEditableTarget(target) {
    if (!target || !(target instanceof Element)) return false;
    const tag = target.tagName ? target.tagName.toLowerCase() : '';
    if (tag === 'input' || tag === 'textarea' || tag === 'select') return true;
    return Boolean(target.closest('input, textarea, select, [contenteditable="true"]'));
}

function shiftFetchRangeLeft() {
    if (
        transformFetchRange((from, to, span) => ({
            from: new Date(from.getTime() - span * 0.25),
            to: new Date(to.getTime() - span * 0.25)
        }))
    ) {
        drawChartData();
    }
}

function shiftFetchRangeRight() {
    if (
        transformFetchRange((from, to, span) => ({
            from: new Date(from.getTime() + span * 0.25),
            to: new Date(to.getTime() + span * 0.25)
        }))
    ) {
        drawChartData();
    }
}

function expandFetchRange() {
    if (
        transformFetchRange((from, to, span) => {
            const center = (from.getTime() + to.getTime()) / 2;
            const nextSpan = span * 2;
            return {
                from: new Date(center - nextSpan / 2),
                to: new Date(center + nextSpan / 2)
            };
        })
    ) {
        drawChartData();
    }
}

function shrinkFetchRange() {
    if (
        transformFetchRange((from, to, span) => {
            const center = (from.getTime() + to.getTime()) / 2;
            const nextSpan = Math.max(60 * 1000, span * 0.5);
            return {
                from: new Date(center - nextSpan / 2),
                to: new Date(center + nextSpan / 2)
            };
        })
    ) {
        drawChartData();
    }
}

function transformFetchRange(transformer) {
    const dateFrom = document.getElementById('chartDateFrom');
    const dateTo = document.getElementById('chartDateTo');
    if (!dateFrom || !dateTo || !dateFrom.value || !dateTo.value) {
        showErrorNotification('Сначала укажите диапазон дат');
        return false;
    }

    const from = new Date(dateFrom.value);
    const to = new Date(dateTo.value);
    if (Number.isNaN(from.getTime()) || Number.isNaN(to.getTime()) || from >= to) {
        showErrorNotification('Некорректный текущий диапазон');
        return false;
    }

    const span = to.getTime() - from.getTime();
    const next = transformer(from, to, span);
    if (!next || Number.isNaN(next.from?.getTime()) || Number.isNaN(next.to?.getTime()) || next.from >= next.to) {
        showErrorNotification('Не удалось изменить диапазон');
        return false;
    }

    dateFrom.value = toDateTimeLocalValue(next.from);
    dateTo.value = toDateTimeLocalValue(next.to);
    writeChartParamsToURL();
    return true;
}

function clearChartPage() {
    const tagInput = document.getElementById('chartTagInput');
    const likeInput = document.getElementById('chartLikeInput');
    const countInput = document.getElementById('chartCount');
    const suggestions = document.getElementById('chartSuggestions');
    const dateFrom = document.getElementById('chartDateFrom');
    const dateTo = document.getElementById('chartDateTo');
    const presetName = document.getElementById('chartPresetName');

    if (tagInput) tagInput.value = '';
    if (likeInput) likeInput.value = 'A20*';
    if (countInput) countInput.value = '300';
    if (presetName) presetName.value = '';
    if (suggestions) suggestions.innerHTML = '';

    const now = new Date();
    const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
    if (dateFrom) dateFrom.value = toDateTimeLocalValue(oneHourAgo);
    if (dateTo) dateTo.value = toDateTimeLocalValue(now);

    chartState.seriesByTag = new Map();
    chartState.hiddenTags = new Set();
    chartState.selectedTags = [];
    chartState.scaleMode = 'common';
    chartState.curveMode = 'linear';
    chartState.aggregationMode = 'raw';
    chartState.downsampleInterval = 'off';
    chartState.renderTargetPoints = DEFAULT_RENDER_TARGET;
    chartState.smoothingWindow = 1;
    chartState.showPoints = true;
    chartState.showValues = false;

    resetChartView();
    renderSelectedChartTags();
    updateViewLabels();
    renderTrendChart();
    writeChartParamsToURL();
}

async function copyChartShareLink() {
    const url = `${window.location.origin}/charts/${window.location.search || ''}`;
    try {
        await navigator.clipboard.writeText(url);
        showSuccessNotification('Ссылка на график скопирована');
    } catch (_) {
        const input = document.createElement('input');
        input.value = url;
        document.body.appendChild(input);
        input.select();
        document.execCommand('copy');
        document.body.removeChild(input);
        showSuccessNotification('Ссылка на график скопирована');
    }
}

function openChartDataView() {
    if (!chartState.selectedTags.length) {
        showErrorNotification('Добавьте минимум один тег');
        return;
    }

    const params = new URLSearchParams();
    params.set('tag', chartState.selectedTags.join(','));

    const dateFrom = document.getElementById('chartDateFrom');
    const dateTo = document.getElementById('chartDateTo');
    const countInput = document.getElementById('chartCount');

    if (dateFrom?.value) params.set('from', dateFrom.value);
    if (dateTo?.value) params.set('to', dateTo.value);
    if (countInput?.value) params.set('count', countInput.value);

    if (typeof window.loadPage === 'function') {
        window.loadPage(`/data/?${params.toString()}`);
    }
}

function saveChartPreset() {
    const presetNameInput = document.getElementById('chartPresetName');
    if (!presetNameInput) return;

    const name = String(presetNameInput.value || '').trim();
    if (!name) {
        showErrorNotification('Введите название пресета');
        return;
    }

    const payload = capturePresetPayload();
    const existingIndex = chartState.presets.findIndex((preset) => preset.name === name);
    const savedPreset = {
        name,
        updatedAt: new Date().toISOString(),
        payload
    };

    if (existingIndex >= 0) chartState.presets[existingIndex] = savedPreset;
    else chartState.presets.push(savedPreset);

    chartState.presets.sort((a, b) => a.name.localeCompare(b.name, 'ru'));
    persistChartPresets();
    renderPresetSelect(name);
    showSuccessNotification(`Пресет "${name}" сохранен`);
}

function applyChartPreset() {
    const select = document.getElementById('chartPresetSelect');
    const nameInput = document.getElementById('chartPresetName');
    const presetName = String((select && select.value) || (nameInput && nameInput.value) || '').trim();

    if (!presetName) {
        showErrorNotification('Выберите пресет');
        return;
    }

    const preset = chartState.presets.find((item) => item.name === presetName);
    if (!preset) {
        showErrorNotification(`Пресет "${presetName}" не найден`);
        return;
    }

    applyPresetPayload(preset.payload);
    if (nameInput) nameInput.value = preset.name;
    renderPresetSelect(preset.name);
    showSuccessNotification(`Пресет "${preset.name}" применен`);
}

function deleteChartPreset() {
    const select = document.getElementById('chartPresetSelect');
    const nameInput = document.getElementById('chartPresetName');
    const presetName = String((select && select.value) || (nameInput && nameInput.value) || '').trim();

    if (!presetName) {
        showErrorNotification('Выберите пресет для удаления');
        return;
    }

    const next = chartState.presets.filter((preset) => preset.name !== presetName);
    if (next.length === chartState.presets.length) {
        showErrorNotification(`Пресет "${presetName}" не найден`);
        return;
    }

    chartState.presets = next;
    persistChartPresets();
    renderPresetSelect();
    if (nameInput) nameInput.value = '';
    showSuccessNotification(`Пресет "${presetName}" удален`);
}

function capturePresetPayload() {
    const dateFrom = document.getElementById('chartDateFrom');
    const dateTo = document.getElementById('chartDateTo');
    const countInput = document.getElementById('chartCount');

    return {
        selectedTags: chartState.selectedTags.slice(),
        from: dateFrom ? dateFrom.value : '',
        to: dateTo ? dateTo.value : '',
        count: countInput ? String(countInput.value || '300') : '300',
        scaleMode: chartState.scaleMode,
        curveMode: chartState.curveMode,
        aggregationMode: chartState.aggregationMode,
        downsampleInterval: chartState.downsampleInterval,
        renderTargetPoints: chartState.renderTargetPoints,
        smoothingWindow: chartState.smoothingWindow,
        showPoints: chartState.showPoints,
        showValues: chartState.showValues,
        zoomX: chartState.zoomX,
        zoomY: chartState.zoomY,
        panX: chartState.panX,
        panY: chartState.panY
    };
}

function applyPresetPayload(payload) {
    if (!payload || typeof payload !== 'object') return;

    chartState.selectedTags = Array.from(new Set((payload.selectedTags || []).map((tag) => String(tag).trim()).filter(Boolean)));
    chartState.hiddenTags = new Set();
    chartState.scaleMode = payload.scaleMode === 'per_tag' ? 'per_tag' : 'common';
    chartState.curveMode = payload.curveMode === 'smooth' ? 'smooth' : 'linear';
    chartState.aggregationMode = sanitizeAggregation(payload.aggregationMode || 'raw');
    chartState.downsampleInterval = sanitizeDownsampleInterval(payload.downsampleInterval || 'off');
    chartState.renderTargetPoints = clamp(Number(payload.renderTargetPoints) || DEFAULT_RENDER_TARGET, 100, 20000);
    chartState.smoothingWindow = clamp(Number(payload.smoothingWindow) || 1, 1, 25);
    chartState.showPoints = payload.showPoints === undefined ? true : Boolean(payload.showPoints);
    chartState.showValues = payload.showValues === undefined ? false : Boolean(payload.showValues);
    chartState.zoomX = clamp(Number(payload.zoomX) || 1, 0.01, 1000000);
    chartState.zoomY = clamp(Number(payload.zoomY) || 1, 0.01, 1000000);
    chartState.panX = Number(payload.panX) || 0;
    chartState.panY = clamp(Number(payload.panY) || 0, -1, 1);

    const dateFrom = document.getElementById('chartDateFrom');
    const dateTo = document.getElementById('chartDateTo');
    const countInput = document.getElementById('chartCount');
    if (dateFrom) dateFrom.value = String(payload.from || dateFrom.value || '');
    if (dateTo) dateTo.value = String(payload.to || dateTo.value || '');
    if (countInput) countInput.value = String(payload.count || countInput.value || '300');

    renderSelectedChartTags();
    updateViewLabels();
    writeChartParamsToURL();
    renderTrendChart();

    if (chartState.selectedTags.length > 0 && dateFrom && dateFrom.value && dateTo && dateTo.value) {
        drawChartData();
    }
}

function loadChartPresets() {
    try {
        const raw = localStorage.getItem(PRESET_STORAGE_KEY);
        if (!raw) {
            chartState.presets = [];
            return;
        }
        const parsed = JSON.parse(raw);
        if (!Array.isArray(parsed)) {
            chartState.presets = [];
            return;
        }
        chartState.presets = parsed
            .filter((item) => item && typeof item === 'object' && typeof item.name === 'string')
            .map((item) => ({
                name: item.name,
                updatedAt: item.updatedAt || '',
                payload: item.payload || {}
            }));
    } catch (_) {
        chartState.presets = [];
    }
}

function persistChartPresets() {
    try {
        localStorage.setItem(PRESET_STORAGE_KEY, JSON.stringify(chartState.presets));
    } catch (error) {
        showErrorNotification(`Не удалось сохранить пресеты: ${error.message}`);
    }
}

function renderPresetSelect(selectedName) {
    const select = document.getElementById('chartPresetSelect');
    if (!select) return;

    const selected = String(selectedName || '').trim();
    const options = ['<option value="">--</option>'];
    chartState.presets.forEach((preset) => {
        const selectedAttr = preset.name === selected ? ' selected' : '';
        options.push(`<option value="${escapeHTML(preset.name)}"${selectedAttr}>${escapeHTML(preset.name)}</option>`);
    });
    select.innerHTML = options.join('');
}

function colorForTag(tag) {
    const index = chartState.selectedTags.indexOf(tag);
    return CHART_COLORS[(index >= 0 ? index : 0) % CHART_COLORS.length];
}

function parseBoolParam(rawValue, fallback) {
    if (rawValue === null || rawValue === undefined) return fallback;
    if (rawValue === '1' || rawValue === 'true') return true;
    if (rawValue === '0' || rawValue === 'false') return false;
    return fallback;
}

function sanitizeAggregation(value) {
    return ['raw', 'avg', 'min', 'max', 'last'].includes(value) ? value : 'raw';
}

function sanitizeDownsampleInterval(value) {
    const allowed = ['off', 'auto', '30s', '1m', '5m', '15m', '1h'];
    return allowed.includes(value) ? value : 'off';
}

function setSelectValue(id, value) {
    const select = document.getElementById(id);
    if (!select) return;
    select.value = value;
}

function clamp(value, min, max) {
    return Math.max(min, Math.min(max, value));
}

function round3(value) {
    return Math.round(value * 1000) / 1000;
}

function convertDateTimeLocal(datetimeLocal) {
    if (!datetimeLocal) return '';
    const date = new Date(datetimeLocal);
    if (Number.isNaN(date.getTime())) return datetimeLocal;
    const day = String(date.getDate()).padStart(2, '0');
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const year = date.getFullYear();
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${day}.${month}.${year} ${hours}:${minutes}:00`;
}

function toDateTimeLocalValue(date) {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${year}-${month}-${day}T${hours}:${minutes}`;
}

function formatTimeLabel(timestampMs) {
    const date = new Date(timestampMs);
    const day = String(date.getDate()).padStart(2, '0');
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${day}.${month} ${hours}:${minutes}`;
}

function formatTimeFull(date) {
    const day = String(date.getDate()).padStart(2, '0');
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const year = date.getFullYear();
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    const seconds = String(date.getSeconds()).padStart(2, '0');
    return `${day}.${month}.${year} ${hours}:${minutes}:${seconds}`;
}

function formatAxisValue(value) {
    if (!Number.isFinite(value)) return '0';
    return value.toFixed(2);
}

function escapeHTML(value) {
    return String(value)
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#39;');
}

function escapeJS(value) {
    return String(value)
        .replaceAll('\\', '\\\\')
        .replaceAll("'", "\\'")
        .replaceAll('"', '\\"');
}

export {
    initializeChartsPage,
    addChartTagsFromInput,
    addSuggestedChartTag,
    removeChartTag,
    clearSelectedChartTags,
    findChartTagsByMask,
    drawChartData,
    clearChartPage,
    zoomInChartX,
    zoomOutChartX,
    zoomInChartY,
    zoomOutChartY,
    resetChartView,
    fitAllChartData,
    shiftFetchRangeLeft,
    shiftFetchRangeRight,
    expandFetchRange,
    shrinkFetchRange,
    toggleSeriesVisibility,
    copyChartShareLink,
    openChartDataView,
    saveChartPreset,
    applyChartPreset,
    deleteChartPreset
};
