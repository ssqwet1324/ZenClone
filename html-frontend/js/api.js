// API базовый URL
const API_BASE = '';

// API запросы
async function apiRequest(url, options = {}) {
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };

    if (state.accessToken) {
        headers['Authorization'] = `Bearer ${state.accessToken}`;
    }

    try {
        const response = await fetch(`${API_BASE}${url}`, {
            ...options,
            headers
        });

        const data = await response.json();

        if (!response.ok) {
            if (response.status === 401 && state.refreshToken) {
                // Попытка обновить токен
                const refreshed = await refreshTokens();
                if (refreshed) {
                    // Повторный запрос
                    headers['Authorization'] = `Bearer ${state.accessToken}`;
                    const retryResponse = await fetch(`${API_BASE}${url}`, {
                        ...options,
                        headers
                    });
                    return await retryResponse.json();
                }
            }
            // Обработка ошибок валидации (массив) и обычных ошибок (объект)
            let errorMessage = 'Ошибка запроса';
            if (data.error) {
                if (Array.isArray(data.error)) {
                    // ErrorResponseValidation - массив ошибок валидации
                    // Маппинг имен полей на русские названия (для регистрации и логина)
                    const fieldNames = {
                        'login': 'Логин',
                        'password': 'Пароль',
                        'username': 'Имя пользователя',
                        'first_name': 'Имя',
                        'last_name': 'Фамилия',
                        'bio': 'О себе'
                    };
                    
                    if (url.includes('/register') || url.includes('/login')) {
                        // Для регистрации и логина показываем имя поля при ошибках валидации
                        errorMessage = data.error.map(err => {
                            const fieldName = fieldNames[err.code] || err.code;
                            return `${fieldName}: ${err.message || err.code}`;
                        }).join(', ');
                    } else {
                        // Для других запросов просто объединяем сообщения
                        errorMessage = data.error.map(err => err.message || err.code).join(', ');
                    }
                } else if (data.error.message) {
                    // ErrorResponse - объект с message (например, INVALID_CREDENTIALS)
                    // Не показываем имя поля для ошибок аутентификации
                    errorMessage = data.error.message;
                } else if (data.error.code) {
                    errorMessage = data.error.code;
                }
            }
            throw new Error(errorMessage);
        }

        return data;
    } catch (error) {
        throw error;
    }
}

// Обновление токенов
async function refreshTokens() {
    try {
        const response = await fetch(`${API_BASE}/api/v1/auth/refresh`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${state.accessToken}`
            },
            body: JSON.stringify({
                refresh_token: state.refreshToken
            })
        });

        const data = await response.json();

        if (response.ok && data.access_token) {
            state.accessToken = data.access_token;
            state.refreshToken = data.refresh_token;
            setLocalStorageItem('accessToken', data.access_token);
            setLocalStorageItem('refreshToken', data.refresh_token);
            return true;
        }
        return false;
    } catch (error) {
        return false;
    }
}

