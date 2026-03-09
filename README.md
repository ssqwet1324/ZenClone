# ZenClone

Социальная сеть с микросервисной архитектурой для общения и обмена контентом.

## 📋 Описание

ZenClone - это полнофункциональная социальная сеть, построенная на микросервисной архитектуре. Проект включает систему аутентификации, управление пользователями, создание и управление постами, подписки и поиск пользователей и ленту постов от пользователей, на которых вы подписаны.

## 🏗️ Архитектура

Проект состоит из следующих компонентов:

### Backend сервисы (Go)

- **AuthService** (порт 8080) - Сервис аутентификации и авторизации
  - Регистрация и вход пользователей
  - Управление JWT токенами (access и refresh)
  - Хранение refresh токенов в Redis

- **UsersService** (порт 8081) - Сервис управления пользователями
  - Управление профилями пользователей
  - Подписки и отписки
  - Загрузка аватаров (MinIO)
  - Поиск пользователей по имени/фамилии или username

- **PostService** (порт 8082) - Сервис управления постами
  - Создание, редактирование и удаление постов
  - Получение постов пользователя с пагинацией
  - Интеграция с Kafka для событий

### Frontend

- **HTML Frontend** (порт 3000) - Веб-интерфейс на чистом HTML/JavaScript
  - Регистрация и авторизация
  - Просмотр и управление профилем
  - Создание и редактирование постов
  - Просмотр подписок
  - Поиск пользователей
  - Бесконечная прокрутка постов

### Инфраструктура

- **PostgreSQL** - Основная база данных
- **Redis** - Хранение refresh токенов
- **Kafka + Zookeeper** - Очередь сообщений для событий
- **MinIO** - Хранилище файлов (аватары пользователей)
- **Kafka UI** (порт 8088) - Веб-интерфейс для управления Kafka

## 🚀 Быстрый старт

### Требования

- Docker и Docker Compose
- Go 1.25+ (для локальной разработки)

### Установка и запуск

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd ZenClone
```

2. Настройте переменные окружения для каждого сервиса:
   - `backend/AuthService/.env`
   - `backend/UsersService/.env`
   - `backend/PostService/.env`

   Примеры конфигурационных файлов должны содержать настройки для:
   - Базы данных PostgreSQL
   - Redis
   - Kafka
   - MinIO
   - JWT секреты

3. Создайте топик в Kafka:
```bash
docker exec -it kafka bash
/usr/bin/kafka-topics --create --topic posts --bootstrap-server kafka:9092 --partitions 3 --replication-factor 1
```

4. Настройте MinIO bucket для аватаров:
```bash
# Установите MinIO Client (mc)
mc alias set localminio http://localhost:9000 minioadmin minioadmin
mc mb localminio/avatars
mc anonymous set download localminio/avatars
```

5. Запустите все сервисы через Docker Compose:
```bash
cd backend
docker-compose up -d
```

6. Откройте браузер и перейдите по адресу:
```
http://localhost:3000
```

## 🔧 Конфигурация

### Переменные окружения

Каждый сервис требует свои переменные окружения. Основные параметры:

#### AuthService
- `JWT_SECRET` - Секретный ключ для JWT
- `REDIS_HOST`, `REDIS_PORT` - Настройки Redis
- `USERS_SERVICE_URL` - URL UsersService для внутренних запросов

#### UsersService
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` - Настройки PostgreSQL
- `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY` - Настройки MinIO
- `MINIO_PUBLIC_ENDPOINT` - Публичный URL MinIO
- `BUCKET_NAME` - Имя bucket для аватаров
- `JWT_SECRET` - Секретный ключ для JWT

#### PostService
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` - Настройки PostgreSQL
- `KAFKA_ADDR` - Адрес Kafka broker
- `KAFKA_TOPIC` - Имя топика Kafka
- `JWT_SECRET` - Секретный ключ для JWT

## 📚 API Документация

Каждый сервис имеет Swagger документацию:

- **AuthService**: `http://localhost:8080/swagger/index.html`
- **UsersService**: `http://localhost:8081/swagger/index.html`
- **PostService**: `http://localhost:8082/swagger/index.html`

## 🔑 Основные функции

### Аутентификация
- Регистрация новых пользователей
- Вход в систему
- Обновление JWT токенов
- Автоматическое обновление токенов

### Профили пользователей
- Просмотр профилей
- Редактирование профиля (имя, фамилия, биография, username)
- Загрузка аватаров
- Просмотр подписок

### Подписки
- Подписка на пользователей
- Отписка от пользователей
- Просмотр списка подписок

### Посты
- Создание постов
- Редактирование постов
- Удаление постов
- Просмотр постов пользователя с пагинацией (cursor-based)
- Бесконечная прокрутка

### Поиск
- Поиск по имени и фамилии (два слова)
- Поиск по username (формат: @username)

## 🛠️ Технологии

### Backend
- **Go 1.25+**
- **Gin** - HTTP веб-фреймворк
- **PostgreSQL** - Реляционная база данных
- **Redis** - In-memory хранилище
- **Kafka** - Очередь сообщений
- **MinIO** - Object storage
- **JWT** - Аутентификация
- **Swagger** - API документация
- **Zap** - Логирование

### Frontend
- **HTML5/CSS3/JavaScript (Vanilla)**
- **Nginx** - Веб-сервер

### Infrastructure
- **Docker** - Контейнеризация
- **Docker Compose** - Оркестрация контейнеров

## 📁 Структура проекта

```
ZenClone/
├── backend/
│   ├── AuthService/          # Сервис аутентификации
│   │   ├── internal/
│   │   │   ├── handler/      # HTTP handlers
│   │   │   ├── usecase/      # Бизнес-логика
│   │   │   ├── repository/   # Redis репозиторий
│   │   │   └── config/       # Конфигурация
│   │   └── .env
│   ├── UsersService/         # Сервис пользователей
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── usecase/
│   │   │   ├── repository/   # PostgreSQL + MinIO
│   │   │   └── config/
│   │   ├── migrations/       # Миграции БД
│   │   └── .env
│   ├── PostService/          # Сервис постов
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── usecase/
│   │   │   ├── repository/   # PostgreSQL
│   │   │   ├── kafka/        # Kafka producer
│   │   │   └── config/
│   │   ├── migrations/
│   │   └── .env
│   └── docker-compose.yml    # Docker Compose конфигурация
└── html-frontend/            # Frontend приложение
    ├── js/                   # JavaScript модули
    ├── index.html
    ├── styles.css
    └── nginx.conf
```

## 🔐 Безопасность

- JWT токены для аутентификации
- Хеширование паролей (bcrypt)
- CORS middleware
- Валидация входных данных
- Проверка прав доступа к ресурсам