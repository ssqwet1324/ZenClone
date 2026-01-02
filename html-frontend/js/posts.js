// Посты

// Загрузка постов пользователя
async function loadUserPosts(username) {
    console.log('loadUserPosts вызвана:', { username, isOwnProfile: state.isOwnProfile, userId: state.userId });
    const postsList = document.getElementById('postsList');
    
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
            postsList.innerHTML = '<div class="empty-state"><h3>Посты недоступны</h3><p>Для просмотра постов этого пользователя нужно быть на него подписанным</p></div>';
            return;
        }
    }

    if (!userId) {
        postsList.innerHTML = '<div class="empty-state"><h3>Нет постов</h3><p>Не удалось загрузить посты</p></div>';
        return;
    }

    try {
        console.log('Запрашиваем посты для userId:', userId);
        const data = await apiRequest(`/api/v1/posts/by-user/${userId}`);

        console.log('Получены посты:', data);
        console.log('isOwnProfile при отображении:', state.isOwnProfile);
        console.log('Количество постов:', data.data?.posts?.length || 0);
        
        if (data.data && data.data.posts && data.data.posts.length > 0) {
            const isOwn = state.isOwnProfile;
            console.log('Отображаем посты, isOwn:', isOwn, 'state.isOwnProfile:', state.isOwnProfile);
            console.log('username:', username, 'state.username:', state.username);
            
            const postsHTML = data.data.posts.map(post => {
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
                console.log('Пост:', post.id, 'actions:', actions ? 'есть' : 'нет');
                return `
                <div class="post-card" data-post-id="${post.id}">
                    <h3>${escapeHtml(post.title)}</h3>
                    <p>${escapeHtml(post.content)}</p>
                    <div class="post-meta">
                        <span>📅 ${new Date(post.created_at).toLocaleDateString('ru-RU')}</span>
                        ${post.updated_at && post.updated_at !== post.created_at ? 
                            `<span>✏️ Обновлено: ${new Date(post.updated_at).toLocaleDateString('ru-RU')}</span>` : ''}
                    </div>
                    ${actions}
                </div>
            `;
            }).join('');
            
            console.log('HTML постов сгенерирован, длина:', postsHTML.length);
            console.log('HTML содержит post-actions:', postsHTML.includes('post-actions'));
            postsList.innerHTML = postsHTML;

            // Проверяем что кнопки действительно в DOM
            setTimeout(() => {
                const testActions = postsList.querySelectorAll('.post-actions');
                console.log('Проверка после вставки: найдено .post-actions:', testActions.length);
                if (testActions.length > 0) {
                    console.log('Первая кнопка действий:', testActions[0].innerHTML.substring(0, 100));
                    const editBtns = testActions[0].querySelectorAll('.edit-post-btn');
                    const deleteBtns = testActions[0].querySelectorAll('.delete-post-btn');
                    console.log('В первой карточке найдено кнопок редактирования:', editBtns.length);
                    console.log('В первой карточке найдено кнопок удаления:', deleteBtns.length);
                } else {
                    console.error('КРИТИЧЕСКАЯ ОШИБКА: .post-actions не найдены в DOM!');
                    console.log('Первые 500 символов HTML:', postsList.innerHTML.substring(0, 500));
                }
            }, 100);

            // Добавляем обработчики для кнопок редактирования и удаления
            if (isOwn) {
                console.log('Добавляем обработчики для кнопок редактирования/удаления, isOwn:', isOwn);
                const editButtons = postsList.querySelectorAll('.edit-post-btn');
                const deleteButtons = postsList.querySelectorAll('.delete-post-btn');
                console.log('Найдено кнопок редактирования:', editButtons.length);
                console.log('Найдено кнопок удаления:', deleteButtons.length);
                
                if (editButtons.length === 0 && deleteButtons.length === 0) {
                    console.error('КРИТИЧЕСКАЯ ОШИБКА: Кнопки не найдены в DOM!');
                    console.log('Содержимое postsList:', postsList.innerHTML.substring(0, 500));
                }
                
                editButtons.forEach(btn => {
                    btn.addEventListener('click', (e) => {
                        const postId = e.target.closest('.edit-post-btn').dataset.postId;
                        const title = e.target.closest('.edit-post-btn').dataset.postTitle;
                        const content = e.target.closest('.edit-post-btn').dataset.postContent;
                        console.log('Клик на редактирование поста:', { postId, title, content });
                        showEditPostModal(postId, title, content);
                    });
                });

                deleteButtons.forEach(btn => {
                    btn.addEventListener('click', (e) => {
                        const postId = e.target.closest('.delete-post-btn').dataset.postId;
                        console.log('Клик на удаление поста:', postId);
                        handleDeletePost(postId);
                    });
                });
            } else {
                console.log('Кнопки редактирования/удаления не добавляются (не свой профиль)');
            }
        } else {
            postsList.innerHTML = '<div class="empty-state"><h3>Нет постов</h3><p>У этого пользователя пока нет публикаций</p></div>';
        }
    } catch (error) {
        console.error('Ошибка загрузки постов:', error);
        postsList.innerHTML = '<div class="empty-state"><h3>Ошибка загрузки постов</h3><p>' + escapeHtml(error.message) + '</p></div>';
    }
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

