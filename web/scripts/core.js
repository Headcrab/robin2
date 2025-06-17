import { showLoader, hideLoader, initializeMobileMenu, closeMobileMenu, showErrorNotification } from './ui.js';
import { updateBreadcrumb } from './navigation.js';

// core page loading functionality
function loadPage(url) {
    const p = saveParams();
    showLoader();
    updateBreadcrumb(url);
    
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
            restoreParams(p);
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
    
    // initialize mobile menu
    initializeMobileMenu();
    
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
        // setup status refresh interval
        setInterval(window.fetchStatus, 60000);
    }
    
    // load home page data if on home page
    if (window.location.pathname === '/' || window.location.pathname === '') {
        if (typeof window.loadHomePageData === 'function') {
            window.loadHomePageData();
        }
    }
    
    // initialize data page if we're on data page
    if (window.location.pathname.includes('/data/')) {
        console.log('On data page, calling initializeDataPage');
        if (typeof initializeDataPage === 'function') {
            initializeDataPage();
        } else {
            setTimeout(() => {
                console.log('Data page initialization completed');
            }, 500);
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