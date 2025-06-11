// DOM observer module for automatic switcher restoration

class DOMObserver {
    constructor() {
        this.observer = null;
        this.isObserving = false;
        this.init();
    }

    init() {
        // create MutationObserver to watch for DOM changes
        this.observer = new MutationObserver((mutations) => {
            this.handleMutations(mutations);
        });

        // start observing when DOM is ready
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', () => {
                this.startObserving();
            });
        } else {
            this.startObserving();
        }
    }

    startObserving() {
        if (this.isObserving || !this.observer) return;

        // observe changes to the entire document
        this.observer.observe(document.body, {
            childList: true,
            subtree: true,
            attributes: false
        });

        this.isObserving = true;
        console.log('DOM observer started - watching for header changes');
    }

    stopObserving() {
        if (!this.isObserving || !this.observer) return;

        this.observer.disconnect();
        this.isObserving = false;
        console.log('DOM observer stopped');
    }

    handleMutations(mutations) {
        let shouldCheckSwitchers = false;

        mutations.forEach((mutation) => {
            // check if header or navigation related elements were added/removed
            mutation.addedNodes.forEach((node) => {
                if (node.nodeType === Node.ELEMENT_NODE) {
                    // check if it's a header or contains header elements
                    if (node.matches && (
                        node.matches('header') ||
                        node.matches('.header') ||
                        node.querySelector('header') ||
                        node.querySelector('.header') ||
                        node.matches('[class*="header"]') ||
                        node.matches('[id*="header"]')
                    )) {
                        shouldCheckSwitchers = true;
                    }
                }
            });

            mutation.removedNodes.forEach((node) => {
                if (node.nodeType === Node.ELEMENT_NODE) {
                    // check if theme/language switchers were removed
                    if (node.id === 'theme-language-switchers' ||
                        node.id === 'theme-toggle' ||
                        node.id === 'language-toggle') {
                        shouldCheckSwitchers = true;
                    }
                }
            });
        });

        if (shouldCheckSwitchers) {
            // debounce the check to avoid multiple calls
            this.debouncedCheckSwitchers();
        }
    }

    debouncedCheckSwitchers() {
        // clear previous timeout
        if (this.checkTimeout) {
            clearTimeout(this.checkTimeout);
        }

        // set new timeout
        this.checkTimeout = setTimeout(() => {
            this.checkAndRestoreSwitchers();
        }, 250);
    }

    checkAndRestoreSwitchers() {
        // check if switchers exist
        const themeToggle = document.getElementById('theme-toggle');
        const languageToggle = document.getElementById('language-toggle');
        const switchersContainer = document.getElementById('theme-language-switchers');

        // check if header exists but switchers don't
        const header = document.querySelector('header') || document.querySelector('.header');
        
        if (header && !switchersContainer) {
            console.log('Header found but switchers missing - attempting to restore');
            
            // attempt to restore switchers
            if (window.initializeThemeAndLanguage) {
                window.initializeThemeAndLanguage();
            } else if (window.addThemeAndLanguageSwitchers) {
                window.addThemeAndLanguageSwitchers();
            }
        }
    }

    // manual trigger for restoration
    manualRestore() {
        console.log('Manual restoration triggered');
        this.checkAndRestoreSwitchers();
    }
}

// create global instance
const domObserver = new DOMObserver();

// export functions
function startDOMObserver() {
    domObserver.startObserving();
}

function stopDOMObserver() {
    domObserver.stopObserving();
}

function restoreSwitchers() {
    domObserver.manualRestore();
}

// export to window for debugging
window.domObserver = domObserver;
window.restoreSwitchers = restoreSwitchers;

export { 
    startDOMObserver, 
    stopDOMObserver, 
    restoreSwitchers,
    domObserver 
}; 