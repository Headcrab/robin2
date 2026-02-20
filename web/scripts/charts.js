import { showErrorNotification, showSuccessNotification } from './ui.js';

const CHART_COLORS = [
    '#6366f1',
    '#06b6d4',
    '#10b981',
    '#f59e0b',
    '#ef4444',
    '#8b5cf6',
    '#14b8a6',
    '#f97316',
    '#ec4899',
    '#22c55e',
    '#0ea5e9',
    '#84cc16'
];

const chartState = {
    selectedTags: [],
    seriesByTag: new Map(),
    hiddenTags: new Set(),
    zoomX: 1,
    zoomY: 1,
    panY: 0,
    scaleMode: 'common',
    curveMode: 'linear',
    smoothingWindow: 1,
    showPoints: true,
    showValues: false
};

function initializeChartsPage() {
    const now = new Date();
    const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
    const dateFrom = document.getElementById('chartDateFrom');
    const dateTo = document.getElementById('chartDateTo');
    const likeInput = document.getElementById('chartLikeInput');

    if (dateFrom && !dateFrom.value) {
        dateFrom.value = toDateTimeLocalValue(oneHourAgo);
    }
    if (dateTo && !dateTo.value) {
        dateTo.value = toDateTimeLocalValue(now);
    }
    if (likeInput && !likeInput.value) {
        likeInput.value = 'A20*';
    }

    applyChartParamsFromURL();
    bindChartControlEvents();
    renderSelectedChartTags();
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
    if (window.chartControlsBound) {
        return;
    }
    window.chartControlsBound = true;

    const tagInput = document.getElementById('chartTagInput');
    if (tagInput) {
        tagInput.addEventListener('keydown', (event) => {
            if (event.key === 'Enter') {
                event.preventDefault();
                addChartTagsFromInput();
            }
        });
    }

    const zoomXRange = document.getElementById('chartZoomRangeX');
    if (zoomXRange) {
        zoomXRange.addEventListener('input', (event) => {
            setChartZoomX(Number(event.target.value));
        });
    }

    const zoomYRange = document.getElementById('chartZoomRangeY');
    if (zoomYRange) {
        zoomYRange.addEventListener('input', (event) => {
            setChartZoomY(Number(event.target.value));
        });
    }

    const smoothingRange = document.getElementById('chartSmoothingRange');
    if (smoothingRange) {
        smoothingRange.addEventListener('input', (event) => {
            const next = Number(event.target.value);
            chartState.smoothingWindow = Number.isFinite(next) ? Math.max(1, Math.floor(next)) : 1;
            updateSmoothingLabel();
            renderTrendChart();
        });
    }

    const scaleMode = document.getElementById('chartScaleMode');
    if (scaleMode) {
        scaleMode.addEventListener('change', (event) => {
            chartState.scaleMode = event.target.value === 'per_tag' ? 'per_tag' : 'common';
            renderTrendChart();
        });
    }

    const curveMode = document.getElementById('chartCurveMode');
    if (curveMode) {
        curveMode.addEventListener('change', (event) => {
            chartState.curveMode = event.target.value === 'smooth' ? 'smooth' : 'linear';
            renderTrendChart();
        });
    }

    const showPoints = document.getElementById('chartShowPoints');
    if (showPoints) {
        showPoints.addEventListener('change', (event) => {
            chartState.showPoints = Boolean(event.target.checked);
            renderTrendChart();
        });
    }

    const showValues = document.getElementById('chartShowValues');
    if (showValues) {
        showValues.addEventListener('change', (event) => {
            chartState.showValues = Boolean(event.target.checked);
            renderTrendChart();
        });
    }
}

function applyChartParamsFromURL() {
    const params = new URLSearchParams(window.location.search);
    const tagsParam = params.get('tags') || params.get('tag');
    if (tagsParam) {
        chartState.selectedTags = Array.from(new Set(tagsParam.split(/[;,]+/).map((tag) => tag.trim()).filter(Boolean)));
    }

    const from = params.get('from');
    const to = params.get('to');
    const count = params.get('count');
    const dateFrom = document.getElementById('chartDateFrom');
    const dateTo = document.getElementById('chartDateTo');
    const countInput = document.getElementById('chartCount');

    if (dateFrom && from) dateFrom.value = from;
    if (dateTo && to) dateTo.value = to;
    if (countInput && count) countInput.value = count;
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

    const query = params.toString();
    history.replaceState(null, '', query ? `/charts/?${query}` : '/charts/');
}

function addChartTagsFromInput() {
    const input = document.getElementById('chartTagInput');
    if (!input) return;

    const tags = String(input.value || '').split(/[,\s;]+/).map((tag) => tag.trim()).filter(Boolean);
    let added = 0;

    tags.forEach((tag) => {
        if (addChartTag(tag)) added += 1;
    });

    input.value = '';
    renderSelectedChartTags();
    if (added === 0 && tags.length > 0) {
        showErrorNotification('Теги уже добавлены');
    }
}

function addChartTag(tag) {
    const clean = String(tag || '').trim();
    if (!clean) return false;
    if (chartState.selectedTags.includes(clean)) return false;

    chartState.selectedTags.push(clean);
    chartState.hiddenTags.delete(clean);
    writeChartParamsToURL();
    return true;
}

function removeChartTag(tag) {
    chartState.selectedTags = chartState.selectedTags.filter((curr) => curr !== tag);
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
    const container = document.getElementById('chartSelectedTags');
    if (!container) return;

    if (chartState.selectedTags.length === 0) {
        container.innerHTML = '<span class="tag-empty-hint">Теги не выбраны</span>';
        return;
    }

    container.innerHTML = chartState.selectedTags.map((tag, index) => `
        <span class="tag-chip" style="border-color:${CHART_COLORS[index % CHART_COLORS.length]}44;">
            <span class="tag-chip-dot" style="background:${CHART_COLORS[index % CHART_COLORS.length]};"></span>
            <span class="tag-chip-name">${escapeHTML(tag)}</span>
            <button type="button" class="tag-chip-remove" onclick="removeChartTag('${escapeJS(tag)}')">&times;</button>
        </span>
    `).join('');
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
            // leave text
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

    container.innerHTML = tags.map((tag) => `
        <button type="button" class="tag-suggestion-btn" onclick="addSuggestedChartTag('${escapeJS(tag)}')">
            <span class="tag-suggestion-name">${escapeHTML(tag)}</span>
            <span class="tag-suggestion-action">+ Добавить</span>
        </button>
    `).join('');
}

function addSuggestedChartTag(tag) {
    if (addChartTag(tag)) {
        renderSelectedChartTags();
    }
}
async function drawChartData() {
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
        const totalPoints = Array.from(chartState.seriesByTag.values()).reduce((sum, arr) => sum + arr.length, 0);
        showSuccessNotification(`График построен. Точек: ${totalPoints}`);
    } catch (error) {
        chartState.seriesByTag = new Map();
        renderTrendChart();
        showErrorNotification(`Ошибка построения графика: ${error.message}`);
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

function applySmoothing(points, windowSize) {
    if (!Array.isArray(points) || points.length < 3 || windowSize <= 1) {
        return points.slice();
    }

    const half = Math.floor(windowSize / 2);
    return points.map((point, index) => {
        const start = Math.max(0, index - half);
        const end = Math.min(points.length - 1, index + half);
        let sum = 0;
        let count = 0;
        for (let i = start; i <= end; i += 1) {
            sum += points[i].value;
            count += 1;
        }
        return {
            timestamp: point.timestamp,
            value: count > 0 ? sum / count : point.value
        };
    });
}

function renderTrendChart() {
    const svg = document.getElementById('trendChartSvg');
    const viewport = document.getElementById('chartViewport');
    const legend = document.getElementById('chartLegend');
    const emptyState = document.getElementById('chartEmptyState');
    if (!svg || !viewport || !legend || !emptyState) return;

    const processedSeries = new Map();
    chartState.selectedTags.forEach((tag) => {
        const points = chartState.seriesByTag.get(tag) || [];
        processedSeries.set(tag, applySmoothing(points, chartState.smoothingWindow));
    });

    const visibleSeries = chartState.selectedTags
        .filter((tag) => !chartState.hiddenTags.has(tag))
        .map((tag) => [tag, processedSeries.get(tag) || []])
        .filter(([, points]) => points.length > 0);

    renderChartLegend(processedSeries);

    if (!visibleSeries.length) {
        svg.innerHTML = '';
        svg.setAttribute('width', '100%');
        svg.setAttribute('height', '460');
        emptyState.style.display = 'flex';
        return;
    }

    emptyState.style.display = 'none';

    const allPoints = visibleSeries.flatMap(([, points]) => points);
    const minTime = Math.min(...allPoints.map((point) => point.timestamp.getTime()));
    const maxTime = Math.max(...allPoints.map((point) => point.timestamp.getTime()));
    const pointsMax = Math.max(...visibleSeries.map(([, points]) => points.length));

    const previousScrollRatio = viewport.scrollWidth > viewport.clientWidth
        ? viewport.scrollLeft / Math.max(1, viewport.scrollWidth - viewport.clientWidth)
        : 0;

    const visibleWidth = Math.max(740, viewport.clientWidth - 20);
    const baseWidth = Math.max(visibleWidth, pointsMax * 44);
    const width = Math.round(baseWidth * chartState.zoomX);
    const height = 460;
    const padLeft = 78;
    const padRight = chartState.scaleMode === 'per_tag' ? 140 : 40;
    const padTop = 22;
    const padBottom = 68;
    const plotWidth = Math.max(1, width - padLeft - padRight);
    const plotHeight = Math.max(1, height - padTop - padBottom);

    const x = (ts) => padLeft + ((ts - minTime) / Math.max(1, maxTime - minTime)) * plotWidth;

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

    const yTicks = 6;
    const xTicks = Math.max(4, Math.min(12, Math.floor(width / 180)));
    const grid = [];
    const yLabels = [];

    for (let i = 0; i <= yTicks; i += 1) {
        const ratio = i / yTicks;
        const gy = padTop + ratio * plotHeight;
        grid.push(`<line x1="${padLeft}" y1="${gy.toFixed(2)}" x2="${(padLeft + plotWidth).toFixed(2)}" y2="${gy.toFixed(2)}" class="chart-grid-line" />`);

        if (chartState.scaleMode === 'common') {
            const value = yCommon.min + (1 - ratio) * (yCommon.max - yCommon.min);
            yLabels.push(`<text x="${padLeft - 9}" y="${(gy + 4).toFixed(2)}" class="chart-axis-label" text-anchor="end">${formatAxisValue(value)}</text>`);
        } else {
            yLabels.push(`<text x="${padLeft - 9}" y="${(gy + 4).toFixed(2)}" class="chart-axis-label" text-anchor="end">${Math.round((1 - ratio) * 100)}%</text>`);
        }
    }

    const xLabels = [];
    for (let i = 0; i <= xTicks; i += 1) {
        const ratio = i / xTicks;
        const gx = padLeft + ratio * plotWidth;
        const ts = minTime + ratio * (maxTime - minTime);
        xLabels.push(`<line x1="${gx.toFixed(2)}" y1="${padTop}" x2="${gx.toFixed(2)}" y2="${(padTop + plotHeight).toFixed(2)}" class="chart-grid-line chart-grid-line-vertical" />`);
        xLabels.push(`<text x="${gx.toFixed(2)}" y="${(padTop + plotHeight + 22).toFixed(2)}" class="chart-axis-label" text-anchor="middle">${formatTimeLabel(ts)}</text>`);
    }

    const axisX = `<line x1="${padLeft}" y1="${(padTop + plotHeight).toFixed(2)}" x2="${(padLeft + plotWidth).toFixed(2)}" y2="${(padTop + plotHeight).toFixed(2)}" class="chart-axis-line" />`;
    const axisY = `<line x1="${padLeft}" y1="${padTop}" x2="${padLeft}" y2="${(padTop + plotHeight).toFixed(2)}" class="chart-axis-line" />`;

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

        const d = buildLinePath(xyPoints, chartState.curveMode);
        seriesShapes.push(`<path d="${d}" fill="none" stroke="${color}" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />`);

        if (chartState.showPoints && points.length <= 2500) {
            seriesShapes.push(xyPoints.map((point) => {
                const title = `${tag}\n${formatTimeFull(point.timestamp)}\n${formatAxisValue(point.value)}`;
                return `<circle cx="${point.x.toFixed(2)}" cy="${point.y.toFixed(2)}" r="2.2" fill="${color}" class="chart-point"><title>${escapeHTML(title)}</title></circle>`;
            }).join(''));
        }

        if (chartState.showValues) {
            const step = Math.max(1, Math.floor(xyPoints.length / Math.max(12, Math.floor(plotWidth / 85))));
            const labels = [];
            for (let i = 0; i < xyPoints.length; i += step) {
                const point = xyPoints[i];
                labels.push(`<text x="${(point.x + 4).toFixed(2)}" y="${(point.y - 6).toFixed(2)}" class="chart-value-label" fill="${color}">${formatAxisValue(point.value)}</text>`);
            }
            const last = xyPoints[xyPoints.length - 1];
            if (last) {
                labels.push(`<text x="${(last.x + 6).toFixed(2)}" y="${(last.y - 8).toFixed(2)}" class="chart-value-label chart-value-label-last" fill="${color}">${formatAxisValue(last.value)}</text>`);
            }
            seriesShapes.push(labels.join(''));
        }

        if (chartState.scaleMode === 'per_tag') {
            const last = xyPoints[xyPoints.length - 1];
            if (last) {
                const xStart = padLeft + plotWidth + 8;
                sideLabels.push(`
                    <line x1="${(padLeft + plotWidth).toFixed(2)}" y1="${last.y.toFixed(2)}" x2="${(xStart - 2).toFixed(2)}" y2="${last.y.toFixed(2)}" stroke="${color}" stroke-width="1.4" />
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
    svg.innerHTML = `
        <rect x="0" y="0" width="${width}" height="${height}" class="chart-bg"></rect>
        <g>${grid.join('')}</g>
        <g>${xLabels.join('')}</g>
        <g>${yLabels.join('')}</g>
        <g>${axisX}${axisY}</g>
        <g>${seriesShapes.join('')}</g>
        <g>${sideLabels.join('')}</g>
    `;

    if (previousScrollRatio > 0 && viewport.scrollWidth > viewport.clientWidth) {
        viewport.scrollLeft = previousScrollRatio * (viewport.scrollWidth - viewport.clientWidth);
    }
}

function renderChartLegend(processedSeries) {
    const legend = document.getElementById('chartLegend');
    if (!legend) return;

    if (chartState.selectedTags.length === 0) {
        legend.innerHTML = '';
        return;
    }

    legend.innerHTML = chartState.selectedTags.map((tag) => {
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
            <button type="button" class="chart-legend-item ${hidden ? 'chart-legend-item-muted' : ''}" onclick="toggleSeriesVisibility('${escapeJS(tag)}')">
                <span class="chart-legend-color" style="background:${color};"></span>
                <span class="chart-legend-name">${escapeHTML(tag)}</span>
                <span class="chart-legend-value">${formatAxisValue(last)}</span>
                <span class="chart-legend-range">${formatAxisValue(min)} .. ${formatAxisValue(max)}</span>
                <span class="chart-legend-toggle">${hidden ? 'show' : 'hide'}</span>
            </button>
        `;
    }).join('');
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
    const pad = spanRaw > 0 ? spanRaw * 0.12 : Math.max(1, Math.abs(maxRaw) * 0.2);
    return { min: minRaw - pad, max: maxRaw + pad };
}

function buildYMapper(baseRange, zoomY, panY, padTop, plotHeight) {
    const baseSpan = Math.max(1e-9, baseRange.max - baseRange.min);
    const zoomedSpan = baseSpan / Math.max(1, zoomY);
    const baseCenter = (baseRange.min + baseRange.max) / 2;
    const maxShift = (baseSpan - zoomedSpan) / 2;
    const shift = Math.max(-1, Math.min(1, panY)) * maxShift;
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

    let d = `M ${points[0].x.toFixed(2)} ${points[0].y.toFixed(2)}`;
    for (let i = 1; i < points.length - 1; i += 1) {
        const xc = (points[i].x + points[i + 1].x) / 2;
        const yc = (points[i].y + points[i + 1].y) / 2;
        d += ` Q ${points[i].x.toFixed(2)} ${points[i].y.toFixed(2)} ${xc.toFixed(2)} ${yc.toFixed(2)}`;
    }
    const last = points[points.length - 1];
    d += ` T ${last.x.toFixed(2)} ${last.y.toFixed(2)}`;
    return d;
}
function setupChartInteractions() {
    const viewport = document.getElementById('chartViewport');
    if (!viewport || viewport.dataset.interactionsBound === '1') return;
    viewport.dataset.interactionsBound = '1';

    let dragMode = null;
    let startX = 0;
    let startY = 0;
    let startScroll = 0;
    let startPanY = 0;

    viewport.addEventListener('contextmenu', (event) => event.preventDefault());

    viewport.addEventListener('pointerdown', (event) => {
        if (event.pointerType === 'mouse' && event.button !== 0) return;
        dragMode = event.shiftKey ? 'y' : 'x';
        startX = event.clientX;
        startY = event.clientY;
        startScroll = viewport.scrollLeft;
        startPanY = chartState.panY;
        viewport.setPointerCapture(event.pointerId);
        viewport.classList.add('dragging');
    });

    viewport.addEventListener('pointermove', (event) => {
        if (!dragMode) return;
        if (dragMode === 'x') {
            viewport.scrollLeft = startScroll - (event.clientX - startX);
        } else {
            const delta = event.clientY - startY;
            chartState.panY = Math.max(-1, Math.min(1, startPanY + delta / 280));
            renderTrendChart();
        }
    });

    const stopDragging = () => {
        dragMode = null;
        viewport.classList.remove('dragging');
    };
    viewport.addEventListener('pointerup', stopDragging);
    viewport.addEventListener('pointercancel', stopDragging);

    viewport.addEventListener('wheel', (event) => {
        if (event.ctrlKey) {
            event.preventDefault();
            const delta = event.deltaY > 0 ? -0.2 : 0.2;
            setChartZoomX(chartState.zoomX + delta);
            return;
        }
        if (event.altKey) {
            event.preventDefault();
            const delta = event.deltaY > 0 ? -0.1 : 0.1;
            setChartZoomY(chartState.zoomY + delta);
            return;
        }
        if (event.shiftKey) {
            event.preventDefault();
            chartState.panY = Math.max(-1, Math.min(1, chartState.panY + event.deltaY / 800));
            renderTrendChart();
        }
    }, { passive: false });
}

function setChartZoomX(next) {
    chartState.zoomX = Math.min(20, Math.max(1, round2(next)));
    const range = document.getElementById('chartZoomRangeX');
    if (range) range.value = String(chartState.zoomX);
    updateZoomLabelX();
    renderTrendChart();
}

function setChartZoomY(next) {
    chartState.zoomY = Math.min(10, Math.max(1, round2(next)));
    const range = document.getElementById('chartZoomRangeY');
    if (range) range.value = String(chartState.zoomY);
    updateZoomLabelY();
    renderTrendChart();
}

function zoomInChartX() {
    setChartZoomX(chartState.zoomX + 0.25);
}

function zoomOutChartX() {
    setChartZoomX(chartState.zoomX - 0.25);
}

function zoomInChartY() {
    setChartZoomY(chartState.zoomY + 0.1);
}

function zoomOutChartY() {
    setChartZoomY(chartState.zoomY - 0.1);
}

function resetChartView() {
    chartState.zoomX = 1;
    chartState.zoomY = 1;
    chartState.panY = 0;

    const rangeX = document.getElementById('chartZoomRangeX');
    const rangeY = document.getElementById('chartZoomRangeY');
    if (rangeX) rangeX.value = '1';
    if (rangeY) rangeY.value = '1';

    const viewport = document.getElementById('chartViewport');
    if (viewport) viewport.scrollLeft = 0;

    updateViewLabels();
    renderTrendChart();
}

function updateViewLabels() {
    updateZoomLabelX();
    updateZoomLabelY();
    updateSmoothingLabel();

    const scaleMode = document.getElementById('chartScaleMode');
    if (scaleMode) scaleMode.value = chartState.scaleMode;
    const curveMode = document.getElementById('chartCurveMode');
    if (curveMode) curveMode.value = chartState.curveMode;
    const showPoints = document.getElementById('chartShowPoints');
    if (showPoints) showPoints.checked = chartState.showPoints;
    const showValues = document.getElementById('chartShowValues');
    if (showValues) showValues.checked = chartState.showValues;
    const smoothingRange = document.getElementById('chartSmoothingRange');
    if (smoothingRange) smoothingRange.value = String(chartState.smoothingWindow);
}

function updateZoomLabelX() {
    const label = document.getElementById('chartZoomLabelX');
    if (label) label.textContent = `${Math.round(chartState.zoomX * 100)}%`;
}

function updateZoomLabelY() {
    const label = document.getElementById('chartZoomLabelY');
    if (label) label.textContent = `${Math.round(chartState.zoomY * 100)}%`;
}

function updateSmoothingLabel() {
    const label = document.getElementById('chartSmoothingLabel');
    if (label) label.textContent = String(chartState.smoothingWindow);
}

function shiftFetchRangeLeft() {
    if (transformFetchRange((from, to, span) => ({
        from: new Date(from.getTime() - span * 0.25),
        to: new Date(to.getTime() - span * 0.25)
    }))) {
        drawChartData();
    }
}

function shiftFetchRangeRight() {
    if (transformFetchRange((from, to, span) => ({
        from: new Date(from.getTime() + span * 0.25),
        to: new Date(to.getTime() + span * 0.25)
    }))) {
        drawChartData();
    }
}

function expandFetchRange() {
    if (transformFetchRange((from, to, span) => {
        const center = (from.getTime() + to.getTime()) / 2;
        const nextSpan = span * 2;
        return {
            from: new Date(center - nextSpan / 2),
            to: new Date(center + nextSpan / 2)
        };
    })) {
        drawChartData();
    }
}

function shrinkFetchRange() {
    if (transformFetchRange((from, to, span) => {
        const center = (from.getTime() + to.getTime()) / 2;
        const nextSpan = Math.max(60 * 1000, span * 0.5);
        return {
            from: new Date(center - nextSpan / 2),
            to: new Date(center + nextSpan / 2)
        };
    })) {
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

    if (tagInput) tagInput.value = '';
    if (likeInput) likeInput.value = 'A20*';
    if (countInput) countInput.value = '300';
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
    chartState.smoothingWindow = 1;
    chartState.showPoints = true;
    chartState.showValues = false;
    resetChartView();
    renderSelectedChartTags();
    renderTrendChart();
    writeChartParamsToURL();
}

function colorForTag(tag) {
    const index = chartState.selectedTags.indexOf(tag);
    return CHART_COLORS[(index >= 0 ? index : 0) % CHART_COLORS.length];
}

function round2(value) {
    return Math.round(value * 100) / 100;
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
    shiftFetchRangeLeft,
    shiftFetchRangeRight,
    expandFetchRange,
    shrinkFetchRange,
    toggleSeriesVisibility
};
