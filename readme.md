# Restaurant Platform

На данный момент реализации проект может послужить шаблоном для **своего** проекта.

## Архитектура

- **Gateway** - Обрабатывает каждую ручку
- **Auth** - Работает с токенами: валидация, генерация, рефреш, отзыв etc
- **User** - Работает с пользователями: регистрация, авторизация

## Технологии

 - Go 1.21
 - gRPC + mTLS
 - SQLite3
 - JWT
 - bcrypt
 - Taskfile

## TaskFile

```
# Генерация сертификатов
task gen-certs

# Компиляция proto файлов
task proto SERVICE=auth VERSION=v3

# Запуск всех сервисов (Windows)
task run-all-win

# Запуск определенного сервиса
task run SERVICE=auth

# Тесты
task tests
task test SERVICE=auth
```

## Gateway API Endpoints
| Метод | Ручка | Описание |
|-------|-------|----------|
| GET | /health | Жив ли gateway |
| GET | /metrics | Метрики сервера |
| POST | /register | Регистрирует пользователя |
| POST | /login | Авторизовывает пользователя |
| POST | /refresh | Обновляет пару токенов по истечению access токена |
| POST | /logout | Отзывает refresh токен, удаляет куку |

## Пример пользования

```
# Жив ли сервис?
curl localhost:8080/health

# Зарегистрировать нового пользователя
curl -X POST localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"Name":"NewUser","Password":"Password"}'

# Залогинить нового пользователя
curl -X POST localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"Name":"NewUser","Password":"Password"}' \
  -c cookies.txt

# Обновить пару токенов(по истечению access токена)
curl -X POST localhost:8080/refresh \
  -c cookies.txt \
  -b cookies.txt

# Выйти с сессии (Отзыв токена + удаление куки)
curl -X POST localhost:8080/logout \
  -c cookies.txt \
  -b cookies.txt

# Узнать метрики сервера
curl -X GET localhost:8080/metrics
```

## Установка
git clone https://github.com/absdekty/restaurant-platform.git

Как использовать? [тык](#taskfile)