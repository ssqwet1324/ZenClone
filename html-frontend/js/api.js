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
            throw new Error(data.error?.message || 'Ошибка запроса');
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

