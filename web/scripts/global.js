// импорты всех модулей
import { loadPage, initialize, getSeason } from './core.js';
import { 
    showErrorNotification, 
    showSuccessNotification, 
    setViewMode,
    initializeThemeAndLanguage
} from './ui.js';
import { 
    getTagOnDate,
    getTagList,
    loadSwagger,
    searchTagData,
    loadHomePageData
} from './data.js';
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
import { startDOMObserver, stopDOMObserver, restoreSwitchers } from './dom-observer.js';

// экспорт функций в глобальную область для совместимости с HTML
window.loadPage = loadPage;
window.getTagOnDate = getTagOnDate;
window.getTagList = getTagList;
window.loadSwagger = loadSwagger;
window.showErrorNotification = showErrorNotification;
window.showSuccessNotification = showSuccessNotification;
window.setViewMode = setViewMode;
window.copyToClipboard = copyToClipboard;
window.searchTagData = searchTagData;
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
window.startDOMObserver = startDOMObserver;
window.stopDOMObserver = stopDOMObserver;
window.restoreSwitchers = restoreSwitchers;

// глобальная инициализация
window.addEventListener('DOMContentLoaded', async () => {
    // инициализируем менеджеры
    await i18nManager.init();
    await themeManager.init();
    
    // инициализируем страницу
    initialize();
    initializeThemeAndLanguage();
    startDOMObserver(); // start watching for DOM changes
}); 