// Инициализация состояния из localStorage
function initState() {
    try {
        state.accessToken = getLocalStorageItem('accessToken');
        state.refreshToken = getLocalStorageItem('refreshToken');
        state.username = getLocalStorageItem('username');
        state.userId = getLocalStorageItem('userId');
    } catch (error) {
        console.error('Ошибка при инициализации состояния:', error);
    }
}

// Инициализируем состояние сразу
initState();

// Инициализация
document.addEventListener('DOMContentLoaded', () => {
    console.log('DOM загружен, инициализация...');
    try {
        initEventListeners();
        checkAuth();
        console.log('Инициализация завершена');
    } catch (error) {
        console.error('Ошибка при инициализации:', error);
    }
});

// Инициализация обработчиков событий
function initEventListeners() {
    try {
        // Навигация
        const loginBtn = document.getElementById('loginBtn');
        const logoutBtn = document.getElementById('logoutBtn');
        const profileBtn = document.getElementById('profileBtn');
        const getStartedBtn = document.getElementById('getStartedBtn');

        if (loginBtn) {
            console.log('Найден loginBtn, добавляем обработчик');
            loginBtn.addEventListener('click', (e) => {
                console.log('Клик на loginBtn!');
                e.preventDefault();
                showAuthModal('login');
            });
        } else {
            console.error('loginBtn не найден!');
        }
        
        if (logoutBtn) logoutBtn.addEventListener('click', logout);
        if (profileBtn) profileBtn.addEventListener('click', () => showProfile(state.username));
        
        if (getStartedBtn) {
            console.log('Найден getStartedBtn, добавляем обработчик');
            getStartedBtn.addEventListener('click', (e) => {
                console.log('Клик на getStartedBtn!');
                e.preventDefault();
                showAuthModal('login');
            });
        } else {
            console.error('getStartedBtn не найден!');
        }

        // Модальные окна
        const closeModal = document.getElementById('closeModal');
        const closeEditModal = document.getElementById('closeEditModal');
        const closeCreatePostModalBtn = document.getElementById('closeCreatePostModal');
        const closeEditPostModalBtn = document.getElementById('closeEditPostModal');
        const loginTab = document.getElementById('loginTab');
        const registerTab = document.getElementById('registerTab');

        if (closeModal) closeModal.addEventListener('click', closeAuthModal);
        if (closeEditModal) closeEditModal.addEventListener('click', closeEditProfileModal);
        if (closeCreatePostModalBtn) {
            closeCreatePostModalBtn.addEventListener('click', () => {
                console.log('Клик на закрытие модального окна создания поста');
                closeCreatePostModal();
            });
        }
        if (closeEditPostModalBtn) {
            closeEditPostModalBtn.addEventListener('click', () => {
                console.log('Клик на закрытие модального окна редактирования поста');
                closeEditPostModal();
            });
        }
        if (loginTab) loginTab.addEventListener('click', () => switchAuthTab('login'));
        if (registerTab) registerTab.addEventListener('click', () => switchAuthTab('register'));

        // Формы
        const loginForm = document.getElementById('loginForm');
        const registerForm = document.getElementById('registerForm');
        const editProfileForm = document.getElementById('editProfileForm');
        const createPostForm = document.getElementById('createPostForm');
        const editPostForm = document.getElementById('editPostForm');

        if (loginForm) loginForm.addEventListener('submit', handleLogin);
        if (registerForm) registerForm.addEventListener('submit', handleRegister);
        if (editProfileForm) editProfileForm.addEventListener('submit', handleUpdateProfile);
        if (createPostForm) createPostForm.addEventListener('submit', handleCreatePost);
        if (editPostForm) editPostForm.addEventListener('submit', handleUpdatePost);

        // Профиль
        const editProfileBtn = document.getElementById('editProfileBtn');
        const subscribeBtn = document.getElementById('subscribeBtn');
        const unsubscribeBtn = document.getElementById('unsubscribeBtn');
        const avatarInput = document.getElementById('avatarInput');
        const createPostBtn = document.getElementById('createPostBtn');

        if (editProfileBtn) editProfileBtn.addEventListener('click', showEditProfileModal);
        if (subscribeBtn) subscribeBtn.addEventListener('click', handleSubscribe);
        if (unsubscribeBtn) unsubscribeBtn.addEventListener('click', handleUnsubscribe);
        if (avatarInput) avatarInput.addEventListener('change', handleAvatarUpload);
        if (createPostBtn) {
            console.log('Обработчик для createPostBtn добавлен');
            createPostBtn.addEventListener('click', (e) => {
                console.log('Клик на кнопку создания поста!');
                e.preventDefault();
                showCreatePostModal();
            });
        } else {
            console.error('createPostBtn не найден при инициализации!');
        }

        // Вкладки профиля
        document.querySelectorAll('[data-tab]').forEach(tab => {
            tab.addEventListener('click', (e) => {
                const tabName = e.target.dataset.tab;
                if (tabName) switchProfileTab(tabName);
            });
        });

        // Делегирование событий для кнопки создания поста (на случай если она создается динамически)
        const profileContainer = document.getElementById('profileContainer');
        if (profileContainer) {
            profileContainer.addEventListener('click', (e) => {
                if (e.target && e.target.id === 'createPostBtn') {
                    console.log('Клик на createPostBtn через делегирование');
                    e.preventDefault();
                    showCreatePostModal();
                }
            });
        }

        // Делегирование событий для подписок (динамически создаваемых элементов)
        const subscriptionsList = document.getElementById('subscriptionsList');
        if (subscriptionsList) {
            subscriptionsList.addEventListener('click', (e) => {
                const card = e.target.closest('.subscription-card');
                if (card) {
                    const username = card.dataset.username;
                    if (username) {
                        showProfile(username);
                    }
                }
            });
        }

        // Делегирование событий для кнопок редактирования и удаления постов
        const postsList = document.getElementById('postsList');
        if (postsList) {
            postsList.addEventListener('click', (e) => {
                const editBtn = e.target.closest('.edit-post-btn');
                const deleteBtn = e.target.closest('.delete-post-btn');
                
                if (editBtn) {
                    const postId = editBtn.dataset.postId;
                    const title = editBtn.dataset.postTitle;
                    const content = editBtn.dataset.postContent;
                    console.log('Клик на редактирование поста через делегирование:', { postId, title, content });
                    showEditPostModal(postId, title, content);
                }
                
                if (deleteBtn) {
                    const postId = deleteBtn.dataset.postId;
                    console.log('Клик на удаление поста через делегирование:', postId);
                    handleDeletePost(postId);
                }
            });
        }

        // Закрытие модальных окон по клику вне их
        window.addEventListener('click', (e) => {
            const authModal = document.getElementById('authModal');
            const editModal = document.getElementById('editProfileModal');
            const createPostModal = document.getElementById('createPostModal');
            const editPostModal = document.getElementById('editPostModal');
            if (e.target === authModal) closeAuthModal();
            if (e.target === editModal) closeEditProfileModal();
            if (e.target === createPostModal) closeCreatePostModal();
            if (e.target === editPostModal) closeEditPostModal();
        });
    } catch (error) {
        console.error('Ошибка инициализации обработчиков событий:', error);
    }
}
