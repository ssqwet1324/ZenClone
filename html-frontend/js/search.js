// Поиск пользователей

// Выполнить поиск пользователей
async function performUserSearch(query) {
    const searchResults = document.getElementById('searchResults');
    searchResults.style.display = 'block';
    searchResults.innerHTML = '<div class="empty-state"><p>Поиск...</p></div>';

    try {
        // Парсим запрос
        const queryParts = query.trim().split(/\s+/);
        
        let url;
        let searchType;

        if (queryParts.length === 1 && queryParts[0].startsWith('@')) {
            // Поиск по username
            const username = queryParts[0].substring(1); // Убираем @
            if (!username) {
                searchResults.innerHTML = '<div class="empty-state"><p>Введите username после @</p></div>';
                return;
            }
            const urlParams = new URLSearchParams();
            urlParams.set('username', username);
            url = `/api/v1/user/search?${urlParams.toString()}`;
            searchType = 'username';
            // Сохраняем username для отображения в результатах
            window.currentSearchUsername = username;
        } else if (queryParts.length === 2) {
            // Поиск по имени и фамилии
            const firstName = queryParts[0];
            const lastName = queryParts[1];
            const urlParams = new URLSearchParams();
            urlParams.set('first_name', firstName);
            urlParams.set('last_name', lastName);
            url = `/api/v1/user/search?${urlParams.toString()}`;
            searchType = 'name';
        } else {
            searchResults.innerHTML = '<div class="empty-state"><p>Введите два слова (имя и фамилия) или @username</p></div>';
            return;
        }

        const data = await apiRequest(url);

        if (searchType === 'username') {
            // Результат поиска по username - один профиль
            displayUsernameSearchResult(data);
        } else {
            // Результат поиска по имени и фамилии - список пользователей
            displayNameSearchResults(data);
        }
    } catch (error) {
        console.error('Ошибка поиска:', error);
        if (error.message && error.message.includes('USER_NOT_FOUND')) {
            searchResults.innerHTML = '<div class="empty-state"><p>Пользователи не найдены</p></div>';
        } else {
            searchResults.innerHTML = '<div class="empty-state"><p>Ошибка поиска: ' + escapeHtml(error.message) + '</p></div>';
        }
    }
}

// Отображение результатов поиска по username (один профиль)
function displayUsernameSearchResult(profileData) {
    const searchResults = document.getElementById('searchResults');
    
    // Проверяем структуру ответа
    const profile = profileData.data || profileData;
    
    const displayName = `${profile.first_name || ''} ${profile.last_name || ''}`.trim() || 'Пользователь';
    // Используем сохраненный username, так как в ответе его нет
    const username = window.currentSearchUsername || '';
    
    if (!username) {
        searchResults.innerHTML = '<div class="empty-state"><p>Ошибка: username не найден</p></div>';
        return;
    }
    
    searchResults.innerHTML = `
        <div class="search-result-item" data-username="${escapeHtml(username)}">
            <img src="${profile.user_avatar_url || 'data:image/svg+xml,%3Csvg xmlns=\'http://www.w3.org/2000/svg\' width=\'40\' height=\'40\'%3E%3Ccircle cx=\'20\' cy=\'20\' r=\'20\' fill=\'%236366f1\'/%3E%3Ctext x=\'50%25\' y=\'50%25\' text-anchor=\'middle\' dy=\'.3em\' fill=\'white\' font-size=\'20\' font-family=\'Arial\'%3E${username.charAt(0).toUpperCase()}%3C/text%3E%3C/svg%3E'}" 
                 alt="${escapeHtml(username)}" class="search-result-avatar">
            <div class="search-result-info">
                <h4>${escapeHtml(displayName)}</h4>
                <p>@${escapeHtml(username)}</p>
            </div>
        </div>
    `;

    // Очищаем сохраненный username
    window.currentSearchUsername = null;

    // Добавляем обработчик клика
    const resultItem = searchResults.querySelector('.search-result-item');
    if (resultItem) {
        resultItem.addEventListener('click', () => {
            const username = resultItem.dataset.username;
            if (username) {
                searchResults.style.display = 'none';
                document.getElementById('searchInput').value = '';
                showProfile(username);
            }
        });
    }
}

// Отображение результатов поиска по имени и фамилии (список)
function displayNameSearchResults(data) {
    const searchResults = document.getElementById('searchResults');
    
    // Проверяем структуру ответа
    const personsList = data.persons || data.data?.persons || [];
    
    if (!personsList || personsList.length === 0) {
        searchResults.innerHTML = '<div class="empty-state"><p>Пользователи не найдены</p></div>';
        return;
    }

    const resultsHTML = personsList.map(person => {
        const displayName = `${person.name || ''} ${person.last_name || ''}`.trim() || 'Пользователь';
        const username = person.username || '';
        const avatarUrl = person.user_avatar_url || `data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='40' height='40'%3E%3Ccircle cx='20' cy='20' r='20' fill='%236366f1'/%3E%3Ctext x='50%25' y='50%25' text-anchor='middle' dy='.3em' fill='white' font-size='20' font-family='Arial'%3E${(username || displayName).charAt(0).toUpperCase()}%3C/text%3E%3C/svg%3E`;
        
        return `
            <div class="search-result-item" data-username="${escapeHtml(username)}">
                <img src="${avatarUrl}" 
                     alt="${escapeHtml(displayName)}" class="search-result-avatar"
                     onerror="this.src='data:image/svg+xml,%3Csvg xmlns=\'http://www.w3.org/2000/svg\' width=\'40\' height=\'40\'%3E%3Ccircle cx=\'20\' cy=\'20\' r=\'20\' fill=\'%236366f1\'/%3E%3Ctext x=\'50%25\' y=\'50%25\' text-anchor=\'middle\' dy=\'.3em\' fill=\'white\' font-size=\'20\' font-family=\'Arial\'%3E${(username || displayName).charAt(0).toUpperCase()}%3C/text%3E%3C/svg%3E'">
                <div class="search-result-info">
                    <h4>${escapeHtml(displayName)}</h4>
                    ${username ? `<p>@${escapeHtml(username)}</p>` : ''}
                </div>
            </div>
        `;
    }).join('');

    searchResults.innerHTML = resultsHTML;

    // Добавляем обработчики кликов
    const resultItems = searchResults.querySelectorAll('.search-result-item');
    resultItems.forEach(item => {
        item.addEventListener('click', () => {
            const username = item.dataset.username;
            if (username) {
                searchResults.style.display = 'none';
                document.getElementById('searchInput').value = '';
                showProfile(username);
            }
        });
    });
}

// Обработчик поиска
function handleSearch() {
    const searchInput = document.getElementById('searchInput');
    const query = searchInput.value.trim();
    
    if (!query) {
        const searchResults = document.getElementById('searchResults');
        searchResults.style.display = 'none';
        return;
    }
    
    performUserSearch(query);
}

// Закрыть результаты поиска при клике вне
document.addEventListener('click', (e) => {
    const searchContainer = document.getElementById('navSearchContainer');
    const searchResults = document.getElementById('searchResults');
    
    if (searchContainer && searchResults && !searchContainer.contains(e.target)) {
        searchResults.style.display = 'none';
    }
});

