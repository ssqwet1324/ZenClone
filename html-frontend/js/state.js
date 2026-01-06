// Состояние приложения
const state = {
    currentUser: null,
    currentProfile: null,
    isOwnProfile: false,
    accessToken: null,
    refreshToken: null,
    username: null,
    userId: null,
    // Пагинация постов в профиле
    postsNextCursor: null,
    postsLoading: false,
    postsHasMore: true,
    currentPostsUserId: null,
    // Пагинация ленты
    feedNextCursor: null,
    feedLoading: false,
    feedHasMore: true
};

