// mobile menu functionality
function initializeMobileMenu() {
    const mobileMenuToggle = document.getElementById('mobile-menu-toggle');
    const mobileMenuClose = document.getElementById('mobile-menu-close');
    const mobileNav = document.getElementById('mobile-nav');
    const mobileMenuOverlay = document.getElementById('mobile-menu-overlay');

    if (mobileMenuToggle) {
        mobileMenuToggle.addEventListener('click', openMobileMenu);
    }

    if (mobileMenuClose) {
        mobileMenuClose.addEventListener('click', closeMobileMenu);
    }

    if (mobileMenuOverlay) {
        mobileMenuOverlay.addEventListener('click', closeMobileMenu);
    }
}

function openMobileMenu() {
    const mobileNav = document.getElementById('mobile-nav');
    const mobileMenuOverlay = document.getElementById('mobile-menu-overlay');
    
    if (mobileNav) {
        mobileNav.classList.remove('-translate-x-full');
    }
    
    if (mobileMenuOverlay) {
        mobileMenuOverlay.classList.remove('hidden');
    }
}

function closeMobileMenu() {
    const mobileNav = document.getElementById('mobile-nav');
    const mobileMenuOverlay = document.getElementById('mobile-menu-overlay');
    
    if (mobileNav) {
        mobileNav.classList.add('-translate-x-full');
    }
    
    if (mobileMenuOverlay) {
        mobileMenuOverlay.classList.add('hidden');
    }
}

// loader functions
function showLoader() {
    const loader = document.getElementById('loader');
    if (loader) {
        loader.classList.remove('hidden');
        loader.classList.add('flex');
    }
}

function hideLoader() {
    const loader = document.getElementById('loader');
    if (loader) {
        loader.classList.add('hidden');
        loader.classList.remove('flex');
    }
}

// notification system
function showErrorNotification(message) {
    const notification = createNotification(message, 'error');
    showNotification(notification);
}

function showSuccessNotification(message) {
    const notification = createNotification(message, 'success');
    showNotification(notification);
}

function createNotification(message, type) {
    const notification = document.createElement('div');
    const bgColor = type === 'error' ? 'bg-red-100 border-red-500 text-red-700' : 'bg-green-100 border-green-500 text-green-700';
    
    notification.className = `fixed top-4 right-4 z-50 max-w-sm w-full ${bgColor} border-l-4 p-4 rounded-md shadow-lg transform translate-x-full transition-transform duration-300 ease-in-out`;
    notification.innerHTML = `
        <div class="flex items-center">
            <div class="flex-1">
                <p class="text-sm font-medium">${message}</p>
            </div>
            <button onclick="this.parentElement.parentElement.remove()" class="ml-3 text-current opacity-50 hover:opacity-75">
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        </div>
    `;
    
    return notification;
}

function showNotification(notification) {
    document.body.appendChild(notification);
    
    // trigger animation
    setTimeout(() => {
        notification.classList.remove('translate-x-full');
    }, 100);
    
    // auto remove after 5 seconds
    setTimeout(() => {
        notification.classList.add('translate-x-full');
        setTimeout(() => {
            if (notification.parentElement) {
                notification.remove();
            }
        }, 300);
    }, 5000);
}

function setViewMode(mode) {
    const listView = document.getElementById('list-view');
    const gridView = document.getElementById('grid-view');
    const listBtn = document.querySelector('[onclick="setViewMode(\'list\')"]');
    const gridBtn = document.querySelector('[onclick="setViewMode(\'grid\')"]');
    
    if (mode === 'list') {
        if (listView) listView.classList.remove('hidden');
        if (gridView) gridView.classList.add('hidden');
        if (listBtn) listBtn.classList.add('btn-primary');
        if (listBtn) listBtn.classList.remove('btn-outline');
        if (gridBtn) gridBtn.classList.remove('btn-primary');
        if (gridBtn) gridBtn.classList.add('btn-outline');
    } else if (mode === 'grid') {
        if (gridView) gridView.classList.remove('hidden');
        if (listView) listView.classList.add('hidden');
        if (gridBtn) gridBtn.classList.add('btn-primary');
        if (gridBtn) gridBtn.classList.remove('btn-outline');
        if (listBtn) listBtn.classList.remove('btn-primary');
        if (listBtn) listBtn.classList.add('btn-outline');
    }
    
    // save preference to localStorage
    localStorage.setItem('tagViewMode', mode);
}

// restore view mode preference on page load
document.addEventListener('DOMContentLoaded', function() {
    const savedViewMode = localStorage.getItem('tagViewMode');
    if (savedViewMode && (savedViewMode === 'list' || savedViewMode === 'grid')) {
        setViewMode(savedViewMode);
    }
});

export { 
    initializeMobileMenu, 
    openMobileMenu, 
    closeMobileMenu, 
    showLoader, 
    hideLoader, 
    showErrorNotification, 
    showSuccessNotification,
    setViewMode
}; 