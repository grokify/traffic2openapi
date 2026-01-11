// Theme toggle functionality
(function() {
    const THEME_KEY = 'traffic2openapi-theme';

    function getPreferredTheme() {
        const stored = localStorage.getItem(THEME_KEY);
        if (stored) {
            return stored;
        }
        return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }

    function setTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem(THEME_KEY, theme);
    }

    // Set initial theme
    setTheme(getPreferredTheme());

    // Theme toggle button
    document.addEventListener('DOMContentLoaded', function() {
        const toggle = document.getElementById('theme-toggle');
        if (toggle) {
            toggle.addEventListener('click', function() {
                const current = document.documentElement.getAttribute('data-theme') || 'light';
                setTheme(current === 'dark' ? 'light' : 'dark');
            });
        }
    });
})();

// Copy to clipboard functionality
(function() {
    document.addEventListener('DOMContentLoaded', function() {
        document.querySelectorAll('.copy-btn').forEach(function(btn) {
            btn.addEventListener('click', function() {
                const targetId = btn.getAttribute('data-copy-target');
                const target = document.getElementById(targetId);
                if (!target) return;

                const text = target.textContent;
                navigator.clipboard.writeText(text).then(function() {
                    btn.classList.add('copied');
                    btn.textContent = 'Copied!';
                    setTimeout(function() {
                        btn.classList.remove('copied');
                        btn.textContent = 'Copy';
                    }, 2000);
                }).catch(function(err) {
                    console.error('Failed to copy:', err);
                });
            });
        });
    });
})();

// View toggle functionality
(function() {
    document.addEventListener('DOMContentLoaded', function() {
        const viewButtons = document.querySelectorAll('.view-btn');

        viewButtons.forEach(function(btn) {
            btn.addEventListener('click', function() {
                const view = btn.getAttribute('data-view');

                // Update button states
                viewButtons.forEach(function(b) {
                    b.classList.remove('active');
                });
                btn.classList.add('active');

                // Update view content
                document.querySelectorAll('.view-content').forEach(function(content) {
                    content.classList.remove('active');
                });
                document.querySelectorAll('.' + view + '-view').forEach(function(content) {
                    content.classList.add('active');
                });
            });
        });
    });
})();

// Simple JSON syntax highlighting
(function() {
    function highlightJSON(text) {
        // Escape HTML first
        text = text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

        // Highlight JSON elements
        return text
            // Strings (property values)
            .replace(/("(?:[^"\\]|\\.)*")(\s*[,\]\}])/g, '<span class="string">$1</span>$2')
            // Property names (keys)
            .replace(/("(?:[^"\\]|\\.)*")(\s*:)/g, '<span class="key">$1</span>$2')
            // Numbers
            .replace(/\b(-?\d+\.?\d*(?:[eE][+-]?\d+)?)\b/g, '<span class="number">$1</span>')
            // Booleans
            .replace(/\b(true|false)\b/g, '<span class="boolean">$1</span>')
            // Null
            .replace(/\b(null)\b/g, '<span class="null">$1</span>');
    }

    document.addEventListener('DOMContentLoaded', function() {
        document.querySelectorAll('code.json').forEach(function(block) {
            const text = block.textContent;
            block.innerHTML = highlightJSON(text);
        });
    });
})();

// Smooth scroll to anchors
(function() {
    document.addEventListener('DOMContentLoaded', function() {
        document.querySelectorAll('a[href^="#"]').forEach(function(anchor) {
            anchor.addEventListener('click', function(e) {
                const href = anchor.getAttribute('href');
                if (href === '#') return;

                const target = document.querySelector(href);
                if (target) {
                    e.preventDefault();
                    target.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                    // Update URL without triggering scroll
                    history.pushState(null, null, href);
                }
            });
        });
    });
})();
