import { showSuccessNotification, showErrorNotification } from './ui.js';

// refresh button functionality
function initializeRefreshButton() {
    const refreshBtn = document.getElementById('refresh-btn');
    if (refreshBtn) {
        refreshBtn.addEventListener('click', function() {
            // add spin animation
            const icon = this.querySelector('svg');
            if (icon) {
                icon.classList.add('animate-spin');
                setTimeout(() => {
                    icon.classList.remove('animate-spin');
                }, 1000);
            }
            
            // refresh current page data
            refreshCurrentPageData();
        });
    }
}

function refreshCurrentPageData() {
    const currentPath = window.location.pathname;
    
    // refresh status
    if (typeof window.fetchStatus === 'function') {
        window.fetchStatus();
    }
    
    // refresh page-specific data
    if (currentPath === '/' || currentPath === '') {
        // refresh home page data
        if (typeof window.loadHomePageData === 'function') {
            window.loadHomePageData();
        }
    }
    
    showSuccessNotification('Данные обновлены');
}

function copyToClipboard(text) {
    if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(text).then(() => {
            showSuccessNotification(`Скопировано в буфер: ${text}`);
        }).catch(err => {
            console.error('Ошибка копирования в буфер обмена:', err);
            fallbackCopyToClipboard(text);
        });
    } else {
        fallbackCopyToClipboard(text);
    }
}

function fallbackCopyToClipboard(text) {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    
    try {
        document.execCommand('copy');
        showSuccessNotification(`Скопировано в буфер: ${text}`);
    } catch (err) {
        console.error('Ошибка копирования в буфер обмена:', err);
        showErrorNotification('Не удалось скопировать в буфер обмена');
    }
    
    document.body.removeChild(textArea);
}

export { 
    initializeRefreshButton,
    copyToClipboard
}; 