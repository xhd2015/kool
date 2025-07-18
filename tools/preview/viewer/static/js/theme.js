// Theme management
export function initializeTheme() {
    const themeToggle = document.getElementById('theme-toggle');
    const savedTheme = localStorage.getItem('theme');

    // Default to light theme
    if (savedTheme === 'dark') {
        document.body.classList.add('dark-theme');
        themeToggle.textContent = 'Light';
    } else {
        themeToggle.textContent = 'Dark';
    }

    themeToggle.addEventListener('click', () => {
        document.body.classList.toggle('dark-theme');
        const isDark = document.body.classList.contains('dark-theme');

        if (isDark) {
            themeToggle.textContent = 'Light';
            localStorage.setItem('theme', 'dark');
        } else {
            themeToggle.textContent = 'Dark';
            localStorage.setItem('theme', 'light');
        }
    });
} 