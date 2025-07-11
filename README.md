# Posts & Comments Service

Простой GraphQL-сервис для управления постами и древовидными комментариями.

---

## Возможности

- Добавление/получение постов
- Комментарии с неограниченной вложенностью
- Пагинация комментариев
- Пагинация постов(по заданию не требовалось, но я посчитал логичным)
- Ограничение на длину комментария (до 2000 символов)
- Запрет комментариев на уровне поста
- Выбор хранилища: PostgreSQL или In-Memory

---

## Запуск проекта


```bash
git clone https://github.com/your-user/posts-comments-service.git
cd posts-comments-service
make docker.full.run
```
или 
```bash
git clone https://github.com/your-user/posts-comments-service.git
cd posts-comments-service
make docker.full.run STORE_TYPE=memory
```

## Примеры запросов
```graphql
mutation CreatePost {
    createPost(
        title: "Мой первый пост",
        content: "Это содержимое моего первого поста",
        author: "user123",
        allowComments: true
    ) {
        id
        title
        content
        author
        allowComments
        createdAt
    }
}
```
Ответ:

```json
{
  "data": {
    "createPost": {
      "id": "115417bc-59ec-4409-acd9-d6e93dabcb81",
      "title": "Мой первый пост",
      "content": "Это содержимое моего первого поста",
      "author": "user123",
      "allowComments": true,
      "createdAt": "2025-07-11T05:00:21Z"
    }
  }
}
```

Следующий запрос:

```graphql
mutation CreateComments {
    comment1: createComment(
        postId: "115417bc-59ec-4409-acd9-d6e93dabcb81",
        text: "Первый комментарий к посту",
        author: "user456"
    ) {
        id
        text
        author
    }
    comment2: createComment(
        postId: "115417bc-59ec-4409-acd9-d6e93dabcb81",
        text: "Второй комментарий к посту",
        author: "user456"
    ) {
        id
        text
        author
    }
}
```

Ответ:

```json
{
  "data": {
    "comment1": {
      "id": "36a4bba7-6b25-4936-8ca6-9127a90a9565",
      "text": "Первый комментарий к посту",
      "author": "user456"
    },
    "comment2": {
      "id": "efdfe0e6-2b66-421d-aadb-d93a17314e62",
      "text": "Второй комментарий к посту",
      "author": "user456"
    }
  }
}
```

И т.д. исходя из API в graphql