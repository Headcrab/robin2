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

// theme and language switcher functions
function createThemeToggle() {
    const themeToggle = document.createElement('button');
    themeToggle.id = 'theme-toggle';
    themeToggle.className = 'p-2 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors duration-200';
    themeToggle.innerHTML = `
        <span class="mr-1">🌙</span>
        <span class="hidden sm:inline text-sm">Темная</span>
    `;
    themeToggle.title = 'Переключить тему';
    themeToggle.onclick = () => {
        if (window.toggleTheme) {
            window.toggleTheme();
        }
    };
    return themeToggle;
}

function createLanguageToggle() {
    const languageContainer = document.createElement('div');
    languageContainer.className = 'relative dropdown';
    
    const languageToggle = document.createElement('button');
    languageToggle.id = 'language-toggle';
    languageToggle.className = 'p-2 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors duration-200 flex items-center';
    languageToggle.setAttribute('data-bs-toggle', 'dropdown');
    languageToggle.setAttribute('aria-expanded', 'false');
    languageToggle.innerHTML = `
        <span class="mr-1">🇷🇺</span>
        <span class="hidden sm:inline text-sm">RU</span>
        <svg class="ml-1 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
    `;
    languageToggle.title = 'Сменить язык';
    
    const dropdown = document.createElement('div');
    dropdown.id = 'language-dropdown';
    dropdown.className = 'absolute right-0 top-full mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50 hidden';
    
    const languages = [
        { code: 'ru', name: 'Русский', flag: '🇷🇺' },
        { code: 'kk', name: 'Қазақша', flag: '🇰🇿' },
        { code: 'en', name: 'English', flag: '🇺🇸' }
    ];
    
    languages.forEach(lang => {
        const a = document.createElement('a');
        a.className = 'block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 cursor-pointer flex items-center';
        a.href = '#';
        a.setAttribute('data-lang', lang.code);
        a.innerHTML = `
            <span class="mr-2">${lang.flag}</span>
            <span>${lang.name}</span>
        `;
        a.onclick = (e) => {
            e.preventDefault();
            if (window.setLanguage) {
                window.setLanguage(lang.code);
            }
            // hide dropdown
            dropdown.classList.add('hidden');
        };
        dropdown.appendChild(a);
    });
    
    // toggle dropdown on click
    languageToggle.onclick = (e) => {
        e.preventDefault();
        dropdown.classList.toggle('hidden');
    };
    
    // hide dropdown when clicking outside
    document.addEventListener('click', (e) => {
        if (!languageContainer.contains(e.target)) {
            dropdown.classList.add('hidden');
        }
    });
    
    languageContainer.appendChild(languageToggle);
    languageContainer.appendChild(dropdown);
    
    return languageContainer;
}

function addThemeAndLanguageSwitchers() {
    // check if switchers already exist
    const existingSwitchers = document.getElementById('theme-language-switchers');
    if (existingSwitchers) {
        console.log('Theme and language switchers already exist, skipping creation');
        return;
    }
    
    // find placeholder in header or fallback to header elements
    const placeholder = document.getElementById('theme-language-switchers-placeholder');
    const headerActions = document.querySelector('header .flex.items-center.space-x-2') ||
                         document.querySelector('.flex.items-center.space-x-4');
    
    const targetContainer = placeholder || headerActions;
    
    if (!targetContainer) {
        console.warn('Header container not found, cannot add theme/language switchers');
        return;
    }
    
    // create container for switchers
    const switchersContainer = document.createElement('div');
    switchersContainer.className = 'flex items-center space-x-2';
    switchersContainer.id = 'theme-language-switchers';
    
    // add theme toggle
    const themeToggle = createThemeToggle();
    switchersContainer.appendChild(themeToggle);
    
    // add language toggle
    const languageToggle = createLanguageToggle();
    switchersContainer.appendChild(languageToggle);
    
    // append to container
    if (placeholder) {
        // replace placeholder
        placeholder.appendChild(switchersContainer);
    } else {
        // append to header actions
        targetContainer.appendChild(switchersContainer);
    }
    
    console.log('Theme and language switchers added to header');
}

function initializeThemeAndLanguage() {
    // wait for DOM to be ready
    const init = () => {
        // add switchers to UI
        addThemeAndLanguageSwitchers();
        
        // переводы обновляются автоматически при setLanguage()
        
        // listen for theme changes to update UI
        window.addEventListener('themeChanged', (event) => {
            console.log('Theme changed:', event.detail.theme);
            // additional UI updates if needed
        });
        
        // listen for language changes to update UI
        window.addEventListener('languageChanged', (event) => {
            console.log('Language changed:', event.detail.language);
            // update translations immediately
            if (window.i18nManager) {
                window.i18nManager.updateTranslations();
                console.log('переводы обновлены при смене языка');
            }
        });
    };
    
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
}

export { 
    initializeMobileMenu, 
    openMobileMenu, 
    closeMobileMenu, 
    showLoader, 
    hideLoader, 
    showErrorNotification, 
    showSuccessNotification,
    setViewMode,
    createThemeToggle,
    createLanguageToggle,
    addThemeAndLanguageSwitchers,
    initializeThemeAndLanguage
}; 