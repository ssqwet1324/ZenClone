// Посты

// Загрузка постов пользователя (первая загрузка или перезагрузка)
async function loadUserPosts(username, reset = true) {
    console.log('loadUserPosts вызвана:', { username, isOwnProfile: state.isOwnProfile, userId: state.userId, reset });
    const postsList = document.getElementById('postsList');
    
    // Сбрасываем состояние пагинации при первой загрузке
    if (reset) {
        state.postsNextCursor = null;
        state.postsHasMore = true;
        state.postsLoading = false;
        state.currentPostsUserId = null;
    }
    
    // Для загрузки постов нужен userID
    // Если это профиль текущего пользователя, используем сохраненный ID
    let userId = null;
    if (state.isOwnProfile && state.userId) {
        userId = state.userId;
        console.log('Используем свой userId:', userId);
    } else {
        // Для других пользователей пытаемся найти ID в подписках текущего пользователя
        try {
            if (state.username) {
                const mySubs = await apiRequest(`/api/v1/user/subs/${state.username}`);
                if (mySubs.subs) {
                    const foundUser = mySubs.subs.find(sub => sub.username === username);
                    if (foundUser) {
                        userId = foundUser.id;
                    }
                }
            }
        } catch (error) {
            console.log('Не удалось получить ID из подписок:', error);
        }
        
        if (!userId) {
            if (reset) {
                postsList.innerHTML = '<div class="empty-state"><h3>Посты недоступны</h3><p>Для просмотра постов этого пользователя нужно быть на него подписанным</p></div>';
            }
            return;
        }
    }

    if (!userId) {
        if (reset) {
            postsList.innerHTML = '<div class="empty-state"><h3>Нет постов</h3><p>Не удалось загрузить посты</p></div>';
        }
        return;
    }

    // Проверяем, не идет ли уже загрузка
    if (state.postsLoading) {
        console.log('Загрузка постов уже идет, пропускаем');
        return;
    }

    // Проверяем, есть ли еще посты для загрузки
    if (!reset && !state.postsHasMore) {
        console.log('Больше нет постов для загрузки');
        return;
    }

    // Проверяем, что это тот же пользователь (при прокрутке)
    if (!reset && state.currentPostsUserId && state.currentPostsUserId !== userId) {
        console.log('Загрузка постов другого пользователя, сбрасываем состояние');
        state.postsNextCursor = null;
        state.postsHasMore = true;
    }

    state.currentPostsUserId = userId;
    state.postsLoading = true;

    try {
        // Формируем URL с query параметрами
        const limit = 20;
        const urlParams = new URLSearchParams();
        urlParams.set('limit', limit.toString());
        if (state.postsNextCursor) {
            urlParams.set('cursor', state.postsNextCursor);
        }
        const url = `/api/v1/posts/by-user/${userId}?${urlParams.toString()}`;

        console.log('Запрашиваем посты для userId:', userId, 'cursor:', state.postsNextCursor, 'url:', url);
        const data = await apiRequest(url);

        console.log('Получены посты:', data);
        console.log('isOwnProfile при отображении:', state.isOwnProfile);
        console.log('Количество постов:', data.data?.posts?.length || 0);
        
        const posts = data.data?.posts || [];
        
        if (posts.length > 0) {
            const isOwn = state.isOwnProfile;
            console.log('Отображаем посты, isOwn:', isOwn);
            
            const postsHTML = posts.map(post => {
                const actions = isOwn ? `
                    <div class="post-actions">
                        <button class="btn btn-secondary btn-small edit-post-btn" 
                                data-post-id="${post.id}" 
                                data-post-title="${escapeForDataAttr(post.title)}" 
                                data-post-content="${escapeForDataAttr(post.content)}">
                            ✏️ Редактировать
                        </button>
                        <button class="btn btn-secondary btn-small btn-danger delete-post-btn" 
                                data-post-id="${post.id}">
                            🗑️ Удалить
                        </button>
                    </div>
                ` : '';
                return `
                <div class="post-card" data-post-id="${post.id}">
                    <h3>${escapeHtml(post.title)}</h3>
                    <p>${escapeHtml(post.content)}</p>
                    <div class="post-meta">
                        <span>📅 ${new Date(post.created_at).toLocaleString('ru-RU', { 
                            year: 'numeric', 
                            month: '2-digit', 
                            day: '2-digit', 
                            hour: '2-digit', 
                            minute: '2-digit' 
                        })}</span>
                        ${post.updated_at && post.updated_at !== post.created_at ? 
                            `<span>✏️ Обновлено: ${new Date(post.updated_at).toLocaleString('ru-RU', { 
                                year: 'numeric', 
                                month: '2-digit', 
                                day: '2-digit', 
                                hour: '2-digit', 
                                minute: '2-digit' 
                            })}</span>` : ''}
                    </div>
                    ${actions}
                </div>
            `;
            }).join('');
            
            // Добавляем посты к существующим или заменяем их
            if (reset) {
                postsList.innerHTML = postsHTML;
            } else {
                // Удаляем индикатор загрузки, если он есть
                const loadingIndicator = postsList.querySelector('.posts-loading');
                if (loadingIndicator) {
                    loadingIndicator.remove();
                }
                postsList.insertAdjacentHTML('beforeend', postsHTML);
            }

            // Обновляем состояние пагинации
            state.postsNextCursor = data.data?.next_cursor || data.data?.nextCursor || null;
            state.postsHasMore = !!state.postsNextCursor && state.postsNextCursor !== '';

            // Обработчики для кнопок редактирования и удаления уже настроены через делегирование в app.js
            // Здесь ничего не делаем

            // Если есть еще посты, удаляем индикатор "конец списка"
            const endIndicator = postsList.querySelector('.posts-end');
            if (endIndicator) {
                endIndicator.remove();
            }
        } else {
            // Нет постов
            if (reset) {
                postsList.innerHTML = '<div class="empty-state"><h3>Нет постов</h3><p>У этого пользователя пока нет публикаций</p></div>';
            } else {
                // Удаляем индикатор загрузки
                const loadingIndicator = postsList.querySelector('.posts-loading');
                if (loadingIndicator) {
                    loadingIndicator.remove();
                }
                // Добавляем индикатор конца списка
                if (!postsList.querySelector('.posts-end')) {
                    postsList.insertAdjacentHTML('beforeend', '<div class="posts-end empty-state"><p>Больше нет постов</p></div>');
                }
            }
            state.postsHasMore = false;
        }
    } catch (error) {
        console.error('Ошибка загрузки постов:', error);
        
        // Если это 404 при первой загрузке - это нормально (нет постов)
        // Если это ошибка при прокрутке - просто останавливаем загрузку
        if (reset) {
            if (error.message && (error.message.includes('404') || error.message.includes('POSTS_NOT_FOUND'))) {
                postsList.innerHTML = '<div class="empty-state"><h3>Нет постов</h3><p>У этого пользователя пока нет публикаций</p></div>';
            } else {
                postsList.innerHTML = '<div class="empty-state"><h3>Ошибка загрузки постов</h3><p>' + escapeHtml(error.message) + '</p></div>';
            }
        } else {
            // Удаляем индикатор загрузки при ошибке
            const loadingIndicator = postsList.querySelector('.posts-loading');
            if (loadingIndicator) {
                loadingIndicator.remove();
            }
            state.postsHasMore = false;
        }
    } finally {
        state.postsLoading = false;
    }
}

// Загрузка следующих постов (для прокрутки)
async function loadMoreUserPosts(username) {
    if (state.postsLoading || !state.postsHasMore) {
        return;
    }
    
    const postsList = document.getElementById('postsList');
    
    // Добавляем индикатор загрузки
    if (!postsList.querySelector('.posts-loading')) {
        postsList.insertAdjacentHTML('beforeend', '<div class="posts-loading empty-state"><p>Загрузка постов...</p></div>');
    }
    
    await loadUserPosts(username, false);
}

// Показать модальное окно создания поста
function showCreatePostModal() {
    console.log('showCreatePostModal вызвана');
    const modal = document.getElementById('createPostModal');
    if (modal) {
        console.log('Модальное окно найдено, показываем');
        modal.classList.add('active');
        console.log('Классы модального окна:', modal.className);
        // Очищаем форму
        const titleInput = document.getElementById('postTitle');
        const contentInput = document.getElementById('postContent');
        const errorDiv = document.getElementById('createPostError');
        if (titleInput) titleInput.value = '';
        if (contentInput) contentInput.value = '';
        if (errorDiv) errorDiv.classList.remove('show');
    } else {
        console.error('Модальное окно createPostModal не найдено!');
        alert('Ошибка: модальное окно не найдено. Проверьте консоль браузера.');
    }
}

// Закрыть модальное окно создания поста
function closeCreatePostModal() {
    console.log('closeCreatePostModal вызвана');
    const modal = document.getElementById('createPostModal');
    if (modal) {
        modal.classList.remove('active');
        const errorDiv = document.getElementById('createPostError');
        if (errorDiv) errorDiv.classList.remove('show');
        console.log('Модальное окно закрыто');
    } else {
        console.error('Модальное окно createPostModal не найдено при закрытии!');
    }
}

// Создание поста
async function handleCreatePost(e) {
    e.preventDefault();
    document.getElementById('createPostError').classList.remove('show');

    const title = document.getElementById('postTitle').value.trim();
    const content = document.getElementById('postContent').value.trim();

    if (!title || !content) {
        showError('createPostError', 'Заголовок и содержание обязательны для заполнения');
        return;
    }

    try {
        const data = await apiRequest('/api/v1/posts/create', {
            method: 'POST',
            body: JSON.stringify({ title, content })
        });

        // Закрываем модальное окно
        closeCreatePostModal();

        // Обновляем список постов
        if (state.username) {
            await loadUserPosts(state.username);
        }

        // Показываем сообщение об успехе
        console.log('Пост успешно создан:', data);
    } catch (error) {
        showError('createPostError', error.message || 'Ошибка создания поста');
    }
}

// Показать модальное окно редактирования поста
function showEditPostModal(postId, title, content) {
    console.log('showEditPostModal вызвана', { postId, title, content });
    const modal = document.getElementById('editPostModal');
    if (modal) {
        console.log('Модальное окно редактирования найдено, показываем');
        modal.classList.add('active');
        console.log('Классы модального окна:', modal.className);
        
        const postIdInput = document.getElementById('editPostId');
        const titleInput = document.getElementById('editPostTitle');
        const contentInput = document.getElementById('editPostContent');
        const errorDiv = document.getElementById('editPostError');
        
        if (postIdInput) postIdInput.value = postId;
        if (titleInput) titleInput.value = title || '';
        if (contentInput) contentInput.value = content || '';
        if (errorDiv) errorDiv.classList.remove('show');
    } else {
        console.error('Модальное окно editPostModal не найдено!');
        alert('Ошибка: модальное окно редактирования не найдено');
    }
}

// Закрыть модальное окно редактирования поста
function closeEditPostModal() {
    console.log('closeEditPostModal вызвана');
    const modal = document.getElementById('editPostModal');
    if (modal) {
        modal.classList.remove('active');
        const errorDiv = document.getElementById('editPostError');
        if (errorDiv) errorDiv.classList.remove('show');
    }
}

// Редактирование поста
async function handleUpdatePost(e) {
    e.preventDefault();
    document.getElementById('editPostError').classList.remove('show');

    const postId = document.getElementById('editPostId').value;
    const title = document.getElementById('editPostTitle').value.trim();
    const content = document.getElementById('editPostContent').value.trim();

    if (!title || !content) {
        showError('editPostError', 'Заголовок и содержание обязательны для заполнения');
        return;
    }

    try {
        await apiRequest(`/api/v1/posts/update/${postId}`, {
            method: 'POST',
            body: JSON.stringify({ 
                title: title,
                content: content,
                updated_at: new Date().toISOString()
            })
        });

        // Закрываем модальное окно
        closeEditPostModal();

        // Обновляем список постов
        if (state.username) {
            await loadUserPosts(state.username);
        }

        console.log('Пост успешно обновлен');
    } catch (error) {
        showError('editPostError', error.message || 'Ошибка обновления поста');
    }
}

// Удаление поста
async function handleDeletePost(postId) {
    if (!confirm('Вы уверены, что хотите удалить этот пост?')) {
        return;
    }

    try {
        await apiRequest(`/api/v1/posts/delete/${postId}`, {
            method: 'DELETE'
        });

        // Обновляем список постов
        if (state.username) {
            await loadUserPosts(state.username);
        }

        console.log('Пост успешно удален');
    } catch (error) {
        alert('Ошибка удаления поста: ' + error.message);
    }
}

