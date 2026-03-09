// импорты всех модулей
import { loadPage, initialize, getSeason } from './core.js';
import { 
    showErrorNotification, 
    showSuccessNotification, 
    setViewMode,
    initializeThemeAndLanguage,
    closeMobileMenu
} from './ui.js';
import { 
    getTagOnDate,
    getTagList,
    loadSwagger,
    searchTagData,
    loadHomePageData,
    initializeDataPage,
    applyDataRangePreset,
    queryCurrentValue,
    queryAggregate,
    queryEvent,
    decodeCurrentTag,
    openDataInCharts
} from './data.js';
import {
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
} from './charts.js';
import { initializeRefreshButton, copyToClipboard } from './utils.js';
import { 
    fetchStatus, 
    updateSystemStatus, 
    updateHomePageStats,
    loadStatistics,
    loadRecentActivity
} from './status.js';
import { 
    clearSearchForm,
    exportData,
    clearTagSearch,
    exportTags,
    exportLogs,
    clearLogs
} from './export.js';
import { toggleTheme, setTheme, getCurrentTheme, getThemes, themeManager } from './themes.js';
import { setLanguage, getCurrentLanguage, getLanguages, t, updateTranslations, i18nManager } from './i18n.js';

// экспорт функций в глобальную область для совместимости с HTML
window.loadPage = loadPage;
window.getTagOnDate = getTagOnDate;
window.getTagList = getTagList;
window.loadSwagger = loadSwagger;
window.applyDataRangePreset = applyDataRangePreset;
window.queryCurrentValue = queryCurrentValue;
window.queryAggregate = queryAggregate;
window.queryEvent = queryEvent;
window.decodeCurrentTag = decodeCurrentTag;
window.openDataInCharts = openDataInCharts;
window.showErrorNotification = showErrorNotification;
window.showSuccessNotification = showSuccessNotification;
window.setViewMode = setViewMode;
window.closeMobileMenu = closeMobileMenu;
window.copyToClipboard = copyToClipboard;
window.searchTagData = searchTagData;
window.initializeDataPage = initializeDataPage;
window.clearSearchForm = clearSearchForm;
window.exportData = exportData;
window.clearTagSearch = clearTagSearch;
window.exportTags = exportTags;
window.exportLogs = exportLogs;
window.clearLogs = clearLogs;
window.initializeRefreshButton = initializeRefreshButton;
window.fetchStatus = fetchStatus;
window.updateSystemStatus = updateSystemStatus;
window.updateHomePageStats = updateHomePageStats;
window.loadStatistics = loadStatistics;
window.loadRecentActivity = loadRecentActivity;
window.loadHomePageData = loadHomePageData;
window.initializeChartsPage = initializeChartsPage;
window.addChartTagsFromInput = addChartTagsFromInput;
window.addSuggestedChartTag = addSuggestedChartTag;
window.removeChartTag = removeChartTag;
window.clearSelectedChartTags = clearSelectedChartTags;
window.findChartTagsByMask = findChartTagsByMask;
window.drawChartData = drawChartData;
window.clearChartPage = clearChartPage;
window.zoomInChartX = zoomInChartX;
window.zoomOutChartX = zoomOutChartX;
window.zoomInChartY = zoomInChartY;
window.zoomOutChartY = zoomOutChartY;
window.resetChartView = resetChartView;
window.fitAllChartData = fitAllChartData;
window.shiftFetchRangeLeft = shiftFetchRangeLeft;
window.shiftFetchRangeRight = shiftFetchRangeRight;
window.expandFetchRange = expandFetchRange;
window.shrinkFetchRange = shrinkFetchRange;
window.toggleSeriesVisibility = toggleSeriesVisibility;
window.copyChartShareLink = copyChartShareLink;
window.openChartDataView = openChartDataView;
window.saveChartPreset = saveChartPreset;
window.applyChartPreset = applyChartPreset;
window.deleteChartPreset = deleteChartPreset;

// theme and language functions
window.toggleTheme = toggleTheme;
window.setTheme = setTheme;
window.getCurrentTheme = getCurrentTheme;
window.getThemes = getThemes;
window.setLanguage = setLanguage;
window.getCurrentLanguage = getCurrentLanguage;
window.getLanguages = getLanguages;
window.t = t;
window.updateTranslations = updateTranslations;
window.themeManager = themeManager;
window.i18nManager = i18nManager;
window.initializeThemeAndLanguage = initializeThemeAndLanguage;

// глобальная инициализация
window.addEventListener('DOMContentLoaded', async () => {
    // инициализируем менеджеры
    await i18nManager.init();
    await themeManager.init();
    
    // инициализируем страницу
    initialize();
    initializeThemeAndLanguage();
}); 
