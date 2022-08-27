# stupid-backend

A stupid backend to test things out. I don't actually know what I'm doing, but why not :P. Planning on using Go with MongoDB for database duties. Probably will make a dart/Flutter plugin at some point. Made primarily to eventually be used with [SWAssitant](https://github.com/CalebQ42/SWAssistant).

## Goals

- [ ] Provide a http.Handler that can be use on any port or address you feel like
  - Maybe provide a small default runner that just runs on a predefined port
- [X] User count
- [ ] User accounts
  - I need to do some research on security to make sure I'm not leaking passwords everywhere
  - Two types of Users:
    - Global users. Can be used across multiple apps
    - App Users. There to keep track of user count. Links to a global account if wanted
- [ ] User uploaded data
- [ ] Access application data
- [ ] Send crash reports
  - Potentially log the current "page" or action to be sent with the report
- [ ] A really bad looking default website (I'm don't use HTML, CSS, or JS very often).
  - I might just make this a Flutter app that can be deployed as a PWA.
- [ ] Protect unauthorized usage with a secret key. If the key is not present, turn off the features cleanly.
  - Allow open-source to be compiled without borking the entire app.
- [ ] Enable or disable as many (or as few) of the above features.
- [ ] Dart/Flutter plugin
  - Plugins for other languages should be relatively simple to make, I'm just focusing on Flutter right now.
- [ ] Properly use `context.Context`
  - I just haven't had much oportunity to use it as of yet
  - Currently everything is just using `context.TODO()`

## Current Features

- Checks api key against database.
- Get features of an api key.
- Log Connection
- Get user count

## Needed Collections

- API Keys
- Global Users
- App Users
- AppData
- UserData
- Crashes

## Queries

This is just an idea on how the API will be structured. Subject to change (just like everything else). Query keys do not need to be in any particular order.

>! TODO: Move alot to request body instead of query.

### Get Features

> `GET: /?features&key=apiKey`

Features:

```text
c - Get User Count
l - Log App Connections
g - Global Users
u - User Data
a - App Data
r - Crash Reporting
```

Returns:

```JSON
{
  "_id": "api key",
  "features": "clguar", //If the letter above is present, the key can acces that feature.
  "death": -1, //Unix timestamp for when the key will expire. If -1, the key has no planned expiration. Subject to change.
}
```

### Log Connection

> `POST: /?logCon&key=apiKey&id=uuid`

### User Count

> `GET: /?userCount&key=apiKey`

Returns:

```JSON
{
  "count": 0,
}
```

### Authentication

> `GET: /?auth&key=apiKey`

Request Body:

```JSON
{
  "username": "username",
  "password": "password"
}
```

Return:

```JSON
{
  "uuid": "uuid", //If invalid login, unauthenticated, or timed-out an empty string is returned.
  "timeout": 0, //If timed-out, returns seconds remaining in timeout. Otherwise returns 0.
  "token": "jwt token"
}
```

### Create User

> `POST: /?createUser&key=apiKey`

Request Body:

```JSON
{
  "uuid": "uuid",
  "username": "username",
  "password": "password"
}
```

Return:

```JSON
{
  "uuid": "uuid", //If invalid login, unauthenticated, or timed-out an empty string is returned.
  "timeout": 0, //If timed-out, returns seconds remaining in timeout. Otherwise returns 0.
  "token": "jwt token"
}
```

## Data Model

This is all just an idea on how the data will be organized in the DB. Subject to change (just like everything else right now).

API Keys:

```JSON
{
  "key": "api key",
  "features": "clguar", //Look at the api for getting features to see what these mean.
  "death": -1, //Unix timestamp for when the key will expire. If -1, the key has no planned expiration (a key can be revoked at any point in time).
}
```

Global User:

```JSON
{
  "uuid": "uuid",
  "username": "name",
  "password": "hashed password", //I need to do research on security before I really set this part up...
  "salt": "salt",
  "email": "email@email.com", //Probably won't be present or used for a while. Only present to be used in the future for account recovery.
  "failed": 0, //Failed logins. Timout occurs every 3 failed logins.
  "lastTimeout": 0, //Unix timestamp of the last timeout issued. Timout is 3^((failed/3)-1) minutes, maxing out at 18 failed attempts for a little over 4 hours of timeout.
}
```

App User:

```JSON
{
  "uuid": "uuid",
  "hasGlobal": true,
  "lastConnected": 20220808 //Records should be deleted if not connected after 30 days. User data should only be deleted if the global account is deleted.
}
```

Application Data:

```JSON
{
  "uuid": "uuid",
  "displayName": "name to be displayed to user",
  "type": "type", //TBD by application. Suggestions include data, config.
  "data": {} //Determined by the application and type.
}
```

User Data:

```JSON
{
  "uuid": "uuid",
  "owner": "user id", //ID of the global user. App users should NOT have info stored.
  "globalRead": false, //Can anyone see this?
  "readPerm": [ //Other users with permission to read the data
    "user id"
  ],
  "writePerm": [ //Other users with permission to write the data
    "user id"
  ],
  "data": {} //Determined by the application.
}
```

Crash reports:

```JSON
{
  "error": "first line of error", //This is to attempt to group together multiple instances of the same error. Possible could become the _id. Possibly might need to be something different.
  "reports": [
    {
      "_id": "uuid", //This is generated at time of crash. Prevents double sending of crash reports (such as if the report needs to be sent on next app launch)
      "stack": "stacktrace",
      "action": "characters" //What page or activity the user was doing
    }
  ]
}
```
