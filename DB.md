# DB Data Structure (MongoDB DB/collection)

Everything here is not concrete.

## API Key (stupid-backend/keys)

```json
{
    "_id": "key",
    "appID": "application ID", //Unique identifier for the app. Allows for multiple apps to use the same backend.
    "alias": "Key alias", //Human readable alias describing the key. Might not be set (empty string).
    "death": -1, //unix timestamp of the planned death of the API Key. If -1, the key has no planned expiration. Keys may be expired at any time without notice.
    "features": { //should be parsed as a map[string]bool. More features can be added as needed by the application.
        "log": true,
        "registeredUsers": true,
        "sendCrash": true,
        "getCount": true,
        "backend": false, //Catch-all for backend site access. Might be removed and replaced with more granular control in the future.
    }
}
```

## Registered Users (stupid-backend/regUser)

```json
{
    "_id": "uuid",
    "username": "username",
    "password": "hashed password",
    "salt": "password salt",
    "email": "email"
}
```

## Logged Connection (appID/log)

These records should not be kept for 30 days and will be cleaned every 24 hours.

```json
{
    "_id": "uuid",
    "platform": "Android", //Android, iOS, Web, Linux, Windows, etc...
    "lastConnection": 20220922 //YYYYMMDD
}
```

## Crash Reports (appID/crashes)

```json
{
    "_id": "uuid",
    "firstLine": "first line of error", //Allows errors of the same type to be groupped together
    "errors": [
        {
            "uuid": "uuid", //to prevent duplicate errors being sent
            "platform": "platform", //Android, iOS, Web, Linux, Windows, etc...
            "stack": "stacktrace",
        }
    ]
}
```
