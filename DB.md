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
        "registeredUsers": true, //Create and authenticate registered users.
        "sendCrash": true,
        "backend": false, //Catch-all for backend site access. Might be removed and replaced with more granular control in the future.
        //Suggested features for data. API Key data features should be checked by the DataApp.
        "appData": true, //Unathenticated requests
        "userData": true, //Authenticated requests
    }
}
```

## Registered Users (stupid-backend/regUsers)

```json
{
    "_id": "uuid",
    "username": "username",
    "password": "hashed password", // argon2ID 32 byte hashed password. Base64 Encoded
    "salt": "password salt", // 16 byte salt. Base64Encoded.
    "email": "email",
    "failed": 0, //Number of failed login attempts. Timeout occurs every 3 failed attempts.
    "lastTimeout": 0 //Unix timestamp of the last timeout.
}
```

Timeout time: 3^((failed/3)-1) minutes. Timeout only occurs every 3 failed attempts. Maxes out at 18 failed attempts at a little over 4 hours of timeout.

## Logged Connection (appID/log)

These records should not be kept for 30 days and will be cleaned every 24 hours.

```json
{
    "_id": "uuid",
    "plat": "Android", //Android, iOS, Web, Linux, Windows, etc...
    "lastConn": 20220922 //YYYYMMDD
}
```

## Crash Reports (appID/crashes)

Single Crash Report:

```json
{
    "_id": "uuid", //to prevent duplicate errors being sent
    "err": "error",
    "plat": "platform", //Android, iOS, Web, Linux, Windows, etc...
    "stack": "stacktrace",
}
```

Grouped Crash Report (what's actually stored):

```json
{
    "_id": "uuid",
    "err": "error",
    "first": "first line of stack", // Better grouping for errors.
    "crashes": [] // An array of single crash reports.
}
```

## Data

Fields required for DefaultDataApp and can also be a suggestion for DataApp's.

```json
{
    "_id": "uuid",
    "owner": "registered user's uuid",
    // Further fields containing all the actual data.
}
```
