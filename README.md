# Stupid Backend

A purposely simple and "stupid" backend. Primarily created for [SWAssistant](https://github.com/CalebQ42/SWAssistant) with that specific implementation found at [swassistant-backend](https://github.com/CalebQ42/swassistant-backend).

## Functions

- Disable or enable any capabilities based on an API key or server configuration.
- Log anonymously
  - Deleted after 30 days. Purely for user count purposes.
- Crash reports
  - Anonymouse with optional extra data.
- User accounts
  - Authentication provided, but specific uses are left up to the implementation.
- All of the above, but with multiple apps using the same backend.
  - Each app will have a seperate App ID.

## Base URLs

These are the available functions in the core setup. These are meant to be added to in a specific implementation.

### API Key Info

> GET: /key/{api_key}

Requires the key permission.

```json
{
  "id": "uuid string",
  "appID": "myApp",
  "alias": "Human readable description of the key",
  "permissions": {
    "key": true, // Get info about this key.
    "count": true, // Get user count; total user and users per platform
    "log": true, // Log a user connecting.
    "userAuth": true, // Authenticate a user account. Includes creating new users.
    "crash": true, // Send crash reports
    // Additional permissions should be added by specific implementations.
  },
  "death": 0 // Unix timestamp (seconds) of the planned death of the key. Keys can be expired at any time without warning. -1 indicates no intended death time.
}
```

### Log Connection

> POST: /log?key={api_key}&id={uuid}&platform={platform}

Requires the log permission.

Logs that a user connected to the API. This is purely meant to get a rough amount of active users. Should be opt-in and IDs are removed if they haven't logged within 30 days.

If successful, returns 201.

### Report Crash

> POST: /crash?key={api_key}

Requires the crash permission.

Request body:

```JSON
{
  "id": "uuid string",
  "error": "error",
  "platform": "platform",
  "stack": "stacktrace"
}
```

If successful, returns 201.

### Authenticate

> POST: /auth?key={api_key}

Requires the userAuth permission

Request Body:

```JSON
{
  "username": "username",
  "password": "password"
}
```

Response:

```JSON
{
  "token": "jwt token",
  "timout": 0 // Minutes remaining until timeout is done.
}
```
