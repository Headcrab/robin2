// theme management module

class ThemeManager {
    constructor() {
        this.currentTheme = 'light';
        this.themes = {
            light: {
                name: 'Светлая',
                icon: '🌙'
            },
            dark: {
                name: 'Темная', 
                icon: '☀️'
            }
        };
        this.init();
    }

    init() {
        // load saved theme from localStorage
        const savedTheme = localStorage.getItem('theme') || 'light';
        this.setTheme(savedTheme);
        
        // listen for system theme changes
        if (window.matchMedia) {
            const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
            mediaQuery.addListener(() => {
                if (!localStorage.getItem('theme')) {
                    this.setTheme(mediaQuery.matches ? 'dark' : 'light');
                }
            });
        }
    }

    setTheme(theme) {
        if (!this.themes[theme]) {
            console.warn(`Theme ${theme} not found`);
            return;
        }

        this.currentTheme = theme;
        
        // update document class
        document.documentElement.classList.remove('light', 'dark');
        document.documentElement.classList.add(theme);
        
        // update data attribute for CSS
        document.documentElement.setAttribute('data-theme', theme);
        
        // save to localStorage
        localStorage.setItem('theme', theme);
        
        // update theme toggle button
        this.updateThemeButton();
        
        // dispatch event for other components
        window.dispatchEvent(new CustomEvent('themeChanged', {
            detail: { theme: theme }
        }));
        
        console.log(`Theme changed to: ${theme}`);
    }

    toggleTheme() {
        const newTheme = this.currentTheme === 'light' ? 'dark' : 'light';
        this.setTheme(newTheme);
    }

    updateThemeButton() {
        const themeBtn = document.getElementById('theme-toggle');
        if (themeBtn) {
            const currentThemeData = this.themes[this.currentTheme];
            const nextTheme = this.currentTheme === 'light' ? 'dark' : 'light';
            const nextThemeData = this.themes[nextTheme];
            
            themeBtn.innerHTML = `
                <span class="mr-1">${nextThemeData.icon}</span>
                <span class="hidden sm:inline text-sm">${nextThemeData.name}</span>
            `;
            
            // use translation if available
            const titleText = window.t ? window.t('theme.toggle') : `Переключить на ${nextThemeData.name.toLowerCase()} тему`;
            themeBtn.title = titleText;
        }
    }

    getCurrentTheme() {
        return this.currentTheme;
    }

    getThemes() {
        return this.themes;
    }
}

// create global instance
const themeManager = new ThemeManager();

// export functions for global access
function toggleTheme() {
    themeManager.toggleTheme();
}

function setTheme(theme) {
    themeManager.setTheme(theme);
}

function getCurrentTheme() {
    return themeManager.getCurrentTheme();
}

function getThemes() {
    return themeManager.getThemes();
}

export { 
    toggleTheme, 
    setTheme, 
    getCurrentTheme, 
    getThemes,
    themeManager
}; 