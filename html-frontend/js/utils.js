// Утилиты

// Безопасное получение из localStorage
function getLocalStorageItem(key) {
    try {
        if (typeof Storage !== 'undefined' && window.localStorage) {
            return window.localStorage.getItem(key);
        }
    } catch (error) {
        console.warn('localStorage недоступен:', error);
    }
    return null;
}

// Безопасная запись в localStorage
function setLocalStorageItem(key, value) {
    try {
        if (typeof Storage !== 'undefined' && window.localStorage) {
            window.localStorage.setItem(key, value);
        }
    } catch (error) {
        console.warn('Не удалось сохранить в localStorage:', error);
    }
}

// Безопасное удаление из localStorage
function removeLocalStorageItem(key) {
    try {
        if (typeof Storage !== 'undefined' && window.localStorage) {
            window.localStorage.removeItem(key);
        }
    } catch (error) {
        console.warn('Не удалось удалить из localStorage:', error);
    }
}

// Экранирование HTML
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Экранирование для data-атрибутов
function escapeForDataAttr(text) {
    if (!text) return '';
    return String(text)
        .replace(/&/g, '&amp;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');
}

// Показать ошибку
function showError(elementId, message) {
    const errorEl = document.getElementById(elementId);
    if (errorEl) {
        errorEl.textContent = message;
        errorEl.classList.add('show');
    }
}

