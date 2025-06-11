// debug script to check if all modules load correctly
console.log('Debug: Starting module check...');

// Check if global.js loaded correctly
setTimeout(() => {
    const requiredFunctions = [
        'loadPage', 'getTagOnDate', 'getTagList', 'loadSwagger',
        'showErrorNotification', 'showSuccessNotification', 'setViewMode',
        'copyToClipboard', 'searchTagData', 'clearSearchForm',
        'exportData', 'clearTagSearch', 'exportTags', 'exportLogs',
        'clearLogs', 'initializeRefreshButton', 'fetchStatus',
        'updateSystemStatus', 'updateHomePageStats', 'loadStatistics',
        'loadRecentActivity', 'loadHomePageData'
    ];
    
    const missing = [];
    const available = [];
    
    requiredFunctions.forEach(fn => {
        if (typeof window[fn] === 'function') {
            available.push(fn);
        } else {
            missing.push(fn);
        }
    });
    
    console.log(`Debug: ${available.length}/${requiredFunctions.length} functions available`);
    
    if (missing.length > 0) {
        console.error('Debug: Missing functions:', missing);
    } else {
        console.log('Debug: All modules loaded successfully! ✅');
    }
    
    if (available.length > 0) {
        console.log('Debug: Available functions:', available);
    }
}, 1000);

export {}; // make this a module 