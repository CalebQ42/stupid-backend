# API Requests

Everything here is not concrete. Request method is only enforced for `auth` and `createUser`, but may be enforced in the future.

## Get API Key features

> HTTPS/GET: /?features&key=apiKey

Returns [Api Key](DB.md#api-key-stupid-backendkeys)

## Log User

> HTTPS/POST: /?log&key=apiKey&uuid=uuid&plat=platform

## Crash Report

> HTTPS/POST: /?crash&key=apiKey

Request Body should be an [single crash report](DB.md#crash-reports-appidcrashes)

## Create User

> HTTPS/POST: /?createUser&key=apiKey

Request Body:

```json
{
    "username": "username",
    "password": "password", // Password must be 5 - 32 characters long
    "email": "email"
}
```

Return Body:

```json
{
    "_id": "uuid",
    "token": "jwt token",
    "problem": "", // If problem is populated, _id and token will be empty strings
}
```

Possible problem values:

* `username`: Username is taken. Has priority over `password`.
* `password`: Password is too short or too long.

## Authenticate

> HTTPS/POST: /?auth&key=apiKey

Request Body:

```json
{
    "username": "username",
    "password": "password",
}
```

Return Body:

```json
{
    "_id": "uuid",
    "token": "jwt token",
    "timeout": 0, //Returns minutes until timeout is up if currently timed-out.
}
```

If the username or password was incorrect, `_id` and `token` will be empty strings and timeout will be 0.

## Data Request

> HTTPS/POST: /?data&key=apiKey&token=jwt token

Token is only necessary for authenticated requests.
