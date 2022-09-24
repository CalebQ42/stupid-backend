# API Requests

Everything here is not concrete.

## Get API Key features

> HTTPS/GET: /?features&key=apiKey

Returns [Api Key](DB.md#api-key)

## User Count

> HTTPS/GET: /?count&key=apiKey

Return

```json
{
    "count": 0, //User count
}
```

## Log User

> HTTPS/POST: /?log&key=apiKey&uuid=uuid

## Authenticate

> HTTPS/POST: /?auth&key=apiKey

Request Body:

```json
{
    "username": "username",
    "password": "password",
}
```

Return Body (If successful):

```json
{
    "_id": "uuid",
    "token": "jwt token"
}
```
