// Профили и подписки

// Показать главную страницу
function showMainPage() {
    document.getElementById('mainContainer').style.display = 'block';
    document.getElementById('profileContainer').style.display = 'none';
}

// Показать профиль
async function showProfile(username) {
    if (!state.accessToken) {
        showAuthModal('login');
        return;
    }

    document.getElementById('mainContainer').style.display = 'none';
    document.getElementById('profileContainer').style.display = 'block';

    state.isOwnProfile = username === state.username;

    try {
        // Загрузка профиля
        const response = await apiRequest(`/api/v1/get-user-profile/${username}`);
        
        // Проверяем структуру ответа - может быть обернут в data или другой ключ
        const profileData = response.data || response;
        
        state.currentProfile = { ...profileData, username };

        // Отображение профиля
        displayProfile(profileData, username);

        // Настройка кнопок (делаем это сразу после загрузки профиля, передаем is_subscribed)
        // Обрабатываем разные форматы: boolean, строка, число
        const isSubscribed = profileData.is_subscribed === true || 
                            profileData.is_subscribed === 'true' || 
                            profileData.is_subscribed === 1 || 
                            profileData.is_subscribed === '1';
        setupProfileButtons(username, isSubscribed);

        // Загрузка подписок (делаем это первым, чтобы потенциально получить ID из списка)
        await loadSubscriptions(username);

        // Загрузка постов
        await loadUserPosts(username);
    } catch (error) {
        console.error('Ошибка загрузки профиля:', error);
        alert('Ошибка загрузки профиля: ' + error.message);
    }
}

// Отображение профиля
function displayProfile(profileData, username) {
    document.getElementById('profileName').textContent = 
        `${profileData.first_name || ''} ${profileData.last_name || ''}`.trim() || username;
    
    document.getElementById('profileBio').textContent = profileData.bio || 'Нет описания';

    const avatarImg = document.getElementById('profileAvatar');
    if (profileData.user_avatar_url) {
        avatarImg.src = profileData.user_avatar_url;
        avatarImg.style.display = 'block';
    } else {
        avatarImg.src = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="150" height="150"%3E%3Ccircle cx="75" cy="75" r="75" fill="%236366f1"/%3E%3Ctext x="50%25" y="50%25" text-anchor="middle" dy=".3em" fill="white" font-size="60" font-family="Arial"%3E' + 
            (username.charAt(0).toUpperCase()) + '%3C/text%3E%3C/svg%3E';
    }
}

// Настройка кнопок профиля
function setupProfileButtons(username, isSubscribed = false) {
    const editBtn = document.getElementById('editProfileBtn');
    const subscribeBtn = document.getElementById('subscribeBtn');
    const unsubscribeBtn = document.getElementById('unsubscribeBtn');
    const avatarUploadLabel = document.getElementById('avatarUploadLabel');
    const createPostBtn = document.getElementById('createPostBtn');
    const postsHeader = document.getElementById('postsHeader');

    if (state.isOwnProfile) {
        if (editBtn) editBtn.style.display = 'block';
        if (subscribeBtn) subscribeBtn.style.display = 'none';
        if (unsubscribeBtn) unsubscribeBtn.style.display = 'none';
        if (avatarUploadLabel) avatarUploadLabel.style.display = 'flex';
        if (createPostBtn) createPostBtn.style.display = 'inline-block';
        if (postsHeader) postsHeader.style.display = 'flex';
    } else {
        if (editBtn) editBtn.style.display = 'none';
        if (avatarUploadLabel) avatarUploadLabel.style.display = 'none';
        if (createPostBtn) createPostBtn.style.display = 'none';
        if (postsHeader) postsHeader.style.display = 'none';
        
        // Показываем правильную кнопку в зависимости от статуса подписки
        // Обрабатываем разные форматы: boolean, строка, число
        const isActuallySubscribed = isSubscribed === true || 
                                     isSubscribed === 'true' || 
                                     isSubscribed === 1 || 
                                     isSubscribed === '1';
        
        if (isActuallySubscribed) {
            if (subscribeBtn) subscribeBtn.style.display = 'none';
            if (unsubscribeBtn) unsubscribeBtn.style.display = 'block';
        } else {
            if (subscribeBtn) subscribeBtn.style.display = 'block';
            if (unsubscribeBtn) unsubscribeBtn.style.display = 'none';
        }
    }
}

// Загрузка подписок
async function loadSubscriptions(username) {
    try {
        const data = await apiRequest(`/api/v1/user/subs/${username}`);
        
        const subscriptionsList = document.getElementById('subscriptionsList');
        
        if (data.subs && data.subs.length > 0) {
            subscriptionsList.innerHTML = data.subs.map(sub => {
                const displayName = `${sub.first_name || ''} ${sub.last_name || ''}`.trim() || sub.username;
                return `
                <div class="subscription-card" data-username="${escapeHtml(sub.username)}">
                    <img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='80' height='80'%3E%3Ccircle cx='40' cy='40' r='40' fill='%236366f1'/%3E%3Ctext x='50%25' y='50%25' text-anchor='middle' dy='.3em' fill='white' font-size='30' font-family='Arial'%3E${sub.username.charAt(0).toUpperCase()}%3C/text%3E%3C/svg%3E" 
                         alt="${escapeHtml(sub.username)}" class="subscription-avatar">
                    <h3>${escapeHtml(displayName)}</h3>
                    <p>@${escapeHtml(sub.username)}</p>
                </div>
            `;
            }).join('');
        } else {
            subscriptionsList.innerHTML = '<div class="empty-state"><h3>Нет подписок</h3><p>Этот пользователь ни на кого не подписан</p></div>';
        }
    } catch (error) {
        console.error('Ошибка загрузки подписок:', error);
        document.getElementById('subscriptionsList').innerHTML = '<div class="empty-state"><h3>Ошибка загрузки подписок</h3><p>' + escapeHtml(error.message) + '</p></div>';
    }
}

// Подписка на пользователя
async function handleSubscribe() {
    if (!state.currentProfile) return;

    try {
        await apiRequest(`/api/v1/user/subscribe/${state.currentProfile.username}`, {
            method: 'POST'
        });

        // Обновляем состояние подписки
        state.currentProfile.is_subscribed = true;
        
        // Обновляем кнопки
        document.getElementById('subscribeBtn').style.display = 'none';
        document.getElementById('unsubscribeBtn').style.display = 'block';
    } catch (error) {
        alert('Ошибка подписки: ' + error.message);
    }
}

// Отписка от пользователя
async function handleUnsubscribe() {
    if (!state.currentProfile) return;

    try {
        await apiRequest(`/api/v1/user/unsubscribe/${state.currentProfile.username}`, {
            method: 'POST'
        });

        // Обновляем состояние подписки
        state.currentProfile.is_subscribed = false;
        
        // Обновляем кнопки
        document.getElementById('unsubscribeBtn').style.display = 'none';
        document.getElementById('subscribeBtn').style.display = 'block';
    } catch (error) {
        alert('Ошибка отписки: ' + error.message);
    }
}

// Переключение вкладок профиля
function switchProfileTab(tabName) {
    document.querySelectorAll('.tab-btn[data-tab]').forEach(btn => {
        btn.classList.remove('active');
    });
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });

    document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');
    document.getElementById(`${tabName}Tab`).classList.add('active');
}

// Показать модальное окно редактирования профиля
function showEditProfileModal() {
    if (!state.currentProfile) return;

    const profile = state.currentProfile;
    document.getElementById('editUsername').value = profile.username || '';
    document.getElementById('editFirstName').value = profile.first_name || '';
    document.getElementById('editLastName').value = profile.last_name || '';
    document.getElementById('editBio').value = profile.bio || '';

    document.getElementById('editProfileModal').classList.add('active');
}

// Закрыть модальное окно редактирования профиля
function closeEditProfileModal() {
    document.getElementById('editProfileModal').classList.remove('active');
    document.getElementById('editProfileError').classList.remove('show');
}

// Обновление профиля
async function handleUpdateProfile(e) {
    e.preventDefault();
    document.getElementById('editProfileError').classList.remove('show');

    const formData = {};
    const username = document.getElementById('editUsername').value;
    const firstName = document.getElementById('editFirstName').value;
    const lastName = document.getElementById('editLastName').value;
    const bio = document.getElementById('editBio').value;
    const passwordOld = document.getElementById('editPasswordOld').value;
    const passwordNew = document.getElementById('editPasswordNew').value;

    if (username) formData.username = username;
    if (firstName) formData.first_name = firstName;
    if (lastName) formData.last_name = lastName;
    if (bio) formData.bio = bio;
    if (passwordOld && passwordNew) {
        formData.password_old = passwordOld;
        formData.password_new = passwordNew;
    }

    try {
        await apiRequest('/api/v1/update-user-info', {
            method: 'POST',
            body: JSON.stringify(formData)
        });

        closeEditProfileModal();
        // Обновляем профиль
        await showProfile(state.username);
    } catch (error) {
        showError('editProfileError', error.message || 'Ошибка обновления профиля');
    }
}

// Загрузка аватара
async function handleAvatarUpload(e) {
    const file = e.target.files[0];
    if (!file) return;

    const formData = new FormData();
    formData.append('avatar', file);

    try {
        const response = await fetch(`${API_BASE}/api/v1/user/upload-avatar`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${state.accessToken}`
            },
            body: formData
        });

        if (!response.ok) {
            const data = await response.json();
            throw new Error(data.error?.message || 'Ошибка загрузки аватара');
        }

        // Обновляем профиль
        await showProfile(state.username);
    } catch (error) {
        alert('Ошибка загрузки аватара: ' + error.message);
    }
}

