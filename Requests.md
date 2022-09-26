# API Requests

Everything here is not concrete.

## Get API Key features

> HTTPS/GET: /?features&key=apiKey

Returns [Api Key](DB.md#api-key-stupid-backendkeys)

## User Count

> HTTPS/GET: /?count&key=apiKey&plat=platform

Platform query value is optional.

Return

```json
{
    "count": 0, //User count
}
```

## Log User

> HTTPS/POST: /?log&key=apiKey&uuid=uuid&plat=platform

## Crash Report

> HTTPS/POST: /?log&key=apiKey

Request Body should be an [single crash report](DB.md#crash-reports-appidcrashes)

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

## Data Request

> HTTPS/POST: /?data&key=apiKey&token=jwt token

Token is only necessary for authenticated requests.
