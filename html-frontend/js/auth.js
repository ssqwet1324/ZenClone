// Авторизация и регистрация

// Показать модальное окно авторизации
function showAuthModal(tab = 'login') {
    console.log('Показываем модальное окно авторизации, вкладка:', tab);
    const modal = document.getElementById('authModal');
    if (!modal) {
        console.error('Модальное окно authModal не найдено!');
        alert('Ошибка: модальное окно не найдено. Проверьте консоль браузера.');
        return;
    }
    console.log('Модальное окно найдено, добавляем класс active');
    modal.classList.add('active');
    console.log('Классы модального окна:', modal.className);
    switchAuthTab(tab);
}

// Закрыть модальное окно авторизации
function closeAuthModal() {
    document.getElementById('authModal').classList.remove('active');
    clearAuthErrors();
}

// Переключение вкладок авторизации
function switchAuthTab(tab) {
    const loginTab = document.getElementById('loginTab');
    const registerTab = document.getElementById('registerTab');
    const loginForm = document.getElementById('loginForm');
    const registerForm = document.getElementById('registerForm');

    if (tab === 'login') {
        loginTab.classList.add('active');
        registerTab.classList.remove('active');
        loginForm.classList.add('active');
        registerForm.classList.remove('active');
    } else {
        registerTab.classList.add('active');
        loginTab.classList.remove('active');
        registerForm.classList.add('active');
        loginForm.classList.remove('active');
    }
}

// Очистка ошибок
function clearAuthErrors() {
    const loginError = document.getElementById('loginError');
    const registerError = document.getElementById('registerError');
    if (loginError) loginError.classList.remove('show');
    if (registerError) registerError.classList.remove('show');
}

// Вход
async function handleLogin(e) {
    e.preventDefault();
    clearAuthErrors();

    const login = document.getElementById('loginEmail').value;
    const password = document.getElementById('loginPassword').value;

    try {
        const data = await apiRequest('/api/v1/auth/login', {
            method: 'POST',
            body: JSON.stringify({ login, password })
        });

        state.accessToken = data.access_token;
        state.refreshToken = data.refresh_token;
        state.username = data.username;
        state.userId = data.id;
        state.currentUser = { username: data.username, id: data.id };

        setLocalStorageItem('accessToken', data.access_token);
        setLocalStorageItem('refreshToken', data.refresh_token);
        setLocalStorageItem('username', data.username);
        setLocalStorageItem('userId', data.id);

        updateNavbar(true);
        closeAuthModal();
        showProfile(data.username);
    } catch (error) {
        showError('loginError', error.message || 'Ошибка входа');
    }
}

// Регистрация
async function handleRegister(e) {
    e.preventDefault();
    clearAuthErrors();

    const formData = {
        login: document.getElementById('regLogin').value,
        password: document.getElementById('regPassword').value,
        username: document.getElementById('regUsername').value,
        first_name: document.getElementById('regFirstName').value || '',
        last_name: document.getElementById('regLastName').value || '',
        bio: document.getElementById('regBio').value || ''
    };

    try {
        const data = await apiRequest('/api/v1/auth/register', {
            method: 'POST',
            body: JSON.stringify(formData)
        });

        state.accessToken = data.access_token;
        state.refreshToken = data.refresh_token;
        state.username = data.username;
        state.userId = data.id;
        state.currentUser = { username: data.username, id: data.id };

        setLocalStorageItem('accessToken', data.access_token);
        setLocalStorageItem('refreshToken', data.refresh_token);
        setLocalStorageItem('username', data.username);
        setLocalStorageItem('userId', data.id);

        updateNavbar(true);
        closeAuthModal();
        showProfile(data.username);
    } catch (error) {
        showError('registerError', error.message || 'Ошибка регистрации');
    }
}

// Выход
function logout() {
    state.accessToken = null;
    state.refreshToken = null;
    state.username = null;
    state.userId = null;
    state.currentUser = null;
    state.currentProfile = null;

    removeLocalStorageItem('accessToken');
    removeLocalStorageItem('refreshToken');
    removeLocalStorageItem('username');
    removeLocalStorageItem('userId');

    updateNavbar(false);
    showMainPage();
}

// Проверка авторизации
function checkAuth() {
    if (state.accessToken && state.username) {
        state.currentUser = { 
            username: state.username,
            id: state.userId
        };
        updateNavbar(true);
    } else {
        updateNavbar(false);
    }
}

// Обновление навигации
function updateNavbar(isAuthenticated) {
    const loginBtn = document.getElementById('loginBtn');
    const logoutBtn = document.getElementById('logoutBtn');
    const profileBtn = document.getElementById('profileBtn');
    const navSearchContainer = document.getElementById('navSearchContainer');

    if (isAuthenticated) {
        if (loginBtn) loginBtn.style.display = 'none';
        if (logoutBtn) logoutBtn.style.display = 'block';
        if (profileBtn) profileBtn.style.display = 'block';
        if (navSearchContainer) navSearchContainer.style.display = 'block';
    } else {
        if (loginBtn) loginBtn.style.display = 'block';
        if (logoutBtn) logoutBtn.style.display = 'none';
        if (profileBtn) profileBtn.style.display = 'none';
        if (navSearchContainer) navSearchContainer.style.display = 'none';
    }
}

