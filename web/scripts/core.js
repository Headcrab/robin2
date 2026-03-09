import { showLoader, hideLoader, initializeShellLayout, closeMobileMenu, showErrorNotification } from './ui.js';
import { updateBreadcrumb } from './navigation.js';

// core page loading functionality
function loadPage(url) {
    saveParams();
    showLoader();
    
    const TIMEOUT = 60000;

    const timeoutPromise = new Promise((resolve, reject) => {
        setTimeout(() => {
            reject(new Error('Превышено время ожидания запроса'));
        }, TIMEOUT);
    });

    return Promise.race([fetch(url), timeoutPromise])
        .then(response => {
            if (!response.ok) {
                throw new Error(`Ошибка сети: ${response.status}`);
            }
            return response.text();
        })
        .then(data => {
            hideLoader();
            document.body.innerHTML = data;
            history.pushState(null, '', url);
            initialize();
            restoreParams();
            closeMobileMenu();
            return data;
        })
        .catch(error => {
            hideLoader();
            console.error('Error:', error);
            showErrorNotification(error.message);
            throw error;
        });
}

// enhanced initialization
function initialize() {
    // initialize robin image
    const robinImage = document.getElementById('robinImage');
    if (robinImage) {
        const season = getSeason();
        const imagePath = `../images/robin_${season}.png`;
        robinImage.src = imagePath;
    }
    
    // initialize shell layout
    initializeShellLayout();
    
    // initialize theme and language switchers
    if (typeof window.initializeThemeAndLanguage === 'function') {
        window.initializeThemeAndLanguage();
    }
    
    // переводы обновляются автоматически в initializeThemeAndLanguage()
    
    // initialize refresh button
    if (typeof window.initializeRefreshButton === 'function') {
        window.initializeRefreshButton();
    }
    
    // fetch initial status
    if (typeof window.fetchStatus === 'function') {
        window.fetchStatus();
        // setup one status refresh interval for the current page
        if (window.statusRefreshIntervalId) {
            clearInterval(window.statusRefreshIntervalId);
        }
        window.statusRefreshIntervalId = setInterval(window.fetchStatus, 60000);
    }
    
    // load home page data if on home page
    if (window.location.pathname === '/' || window.location.pathname === '') {
        if (typeof window.loadHomePageData === 'function') {
            window.loadHomePageData();
        }
    }
    
    // initialize data page if we're on data page
    if (window.location.pathname.includes('/data/')) {
        if (typeof window.initializeDataPage === 'function') {
            window.initializeDataPage();
        }
    }

    // initialize charts page
    if (window.location.pathname.includes('/charts/')) {
        if (typeof window.initializeChartsPage === 'function') {
            window.initializeChartsPage();
        }
    }

    if (window.location.pathname.includes('/tags/')) {
        if (typeof window.setViewMode === 'function') {
            const savedMode = localStorage.getItem('tagsViewMode') || 'list';
            window.setViewMode(savedMode);
        }
    }
    
    // update breadcrumb
    updateBreadcrumb(window.location.pathname);
}

function saveParams() {
    if (document.getElementById("searchInput") != null)
        sessionStorage.setItem("searchInput", document.getElementById("searchInput").value);
    if (document.getElementById("dateFrom") != null)
        sessionStorage.setItem("dateFrom", document.getElementById("dateFrom").value);
    if (document.getElementById("dateTo") != null)
        sessionStorage.setItem("dateTo", document.getElementById("dateTo").value);
    if (document.getElementById("searchCount") != null)
        sessionStorage.setItem("searchCount", document.getElementById("searchCount").value);
    if (document.getElementById("chartTagInput") != null)
        sessionStorage.setItem("chartTagInput", document.getElementById("chartTagInput").value);
    if (document.getElementById("chartLikeInput") != null)
        sessionStorage.setItem("chartLikeInput", document.getElementById("chartLikeInput").value);
    if (document.getElementById("chartDateFrom") != null)
        sessionStorage.setItem("chartDateFrom", document.getElementById("chartDateFrom").value);
    if (document.getElementById("chartDateTo") != null)
        sessionStorage.setItem("chartDateTo", document.getElementById("chartDateTo").value);
    if (document.getElementById("chartCount") != null)
        sessionStorage.setItem("chartCount", document.getElementById("chartCount").value);
}

function restoreParams() {
    if (sessionStorage.getItem("searchInput")) {
        if (document.getElementById("searchInput") != null)
            document.getElementById("searchInput").value = sessionStorage.getItem("searchInput");
        if (document.getElementById("dateFrom") != null)
            document.getElementById("dateFrom").value = sessionStorage.getItem("dateFrom");
        if (document.getElementById("dateTo") != null)
            document.getElementById("dateTo").value = sessionStorage.getItem("dateTo");
        if (document.getElementById("searchCount") != null)
            document.getElementById("searchCount").value = sessionStorage.getItem("searchCount");
    }
    if (sessionStorage.getItem("chartTagInput") && document.getElementById("chartTagInput") != null)
        document.getElementById("chartTagInput").value = sessionStorage.getItem("chartTagInput");
    if (sessionStorage.getItem("chartLikeInput") && document.getElementById("chartLikeInput") != null)
        document.getElementById("chartLikeInput").value = sessionStorage.getItem("chartLikeInput");
    if (sessionStorage.getItem("chartDateFrom") && document.getElementById("chartDateFrom") != null)
        document.getElementById("chartDateFrom").value = sessionStorage.getItem("chartDateFrom");
    if (sessionStorage.getItem("chartDateTo") && document.getElementById("chartDateTo") != null)
        document.getElementById("chartDateTo").value = sessionStorage.getItem("chartDateTo");
    if (sessionStorage.getItem("chartCount") && document.getElementById("chartCount") != null)
        document.getElementById("chartCount").value = sessionStorage.getItem("chartCount");
}

function getSeason() {
    // Получаем сезон из текущей даты
    var date = new Date();
    var month = date.getMonth() + 1;
    var season = Math.floor(month / 3) + 1;
    if (season > 4) {
        season = 1;
    }
    // winter, spring, summer, fall
    var seasons = ['winter', 'spring', 'summer', 'fall'];
    var seasonName = seasons[season - 1];
    return seasonName;
}

// инициализация происходит в global.js

export { loadPage, initialize, saveParams, restoreParams, getSeason }; 
