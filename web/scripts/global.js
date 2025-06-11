// импорты всех модулей
import { loadPage, initialize, getSeason } from './core.js';
import { 
    showErrorNotification, 
    showSuccessNotification, 
    setViewMode 
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

// глобальная инициализация
window.addEventListener('DOMContentLoaded', initialize); 