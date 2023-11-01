# Stupid Backend

A purposely simple and "stupid" backend. Primarily created for [SWAssistant](https://github.com/CalebQ42/SWAssistant) with that specific implementation found at [swassistant-backend](https://github.com/CalebQ42/swassistant-backend). Though made for SWAssistant, this is a barebones that provides some common use cases I (and potentially others) that can be quickly and easily deployed.

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
  - Users will be shared between multiple apps.

## Base Authenticed URLs

These are the available functions in the core setup for `KeyedApp`s. These are meant to be added to in a specific implementation.

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
    "count": true, // Get user count; Total user and users per platform. Based on Log.
    "log": true, // Log a user connecting.
    "auth": true, // Authenticate and create user accounts.
    "crash": true // Send crash reports
    // Additional permissions should be added by specific implementations.
  },
  "death": -1 // Unix timestamp (seconds) of the planned death of the key. Keys can be expired at any time without warning. -1 indicates no intended death time.
}
```

### User Count

> GET: /count?key={api_key}&platform={platform}

Required the count permission. Platform query is optional.

```json
{
  "platform": "platform", // If no platform is given, will be "all".
  "count": 0
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
  "stack": "stacktrace",
  "version": "app version",
}
```

If successful, returns 201.

### Create User

> POST: /createUser?key={api_key}

Requires the userAuth permission. If not using stupid-server, make sure user creation is only allowed when using TLS.

Request Body:

```JSON
{
  // Must not be empty or just spaces and cannot end or begin with spaces.
  // Must be less then 64 characters.
  // Spaces will be trimmed.
  "username": "username",
  // Passwords must be between 5-32 characters.
  "password": "password",
  "email": "email"
}
```

Response:

```JSON
{
  "token": "jwt token",
  // Only populated if there's some problem with creating the user.
  // username - Username is already taken or invalid.
  // email - email is (probably) invalid.
  // password - Password is invalid.
  "problem": "username"
}
```

### Authenticate

> POST: /auth?key={api_key}

Requires the userAuth permission. If not using stupid-server, make sure user authentication is only allowed when using TLS.

Requests will timout every 3 failed attempts for `3^((failed/3)-1)` minutes with a maximum of 60 minutes of timeout.

Request Body:

```JSON
{
  "username": "username",
  "password": "password"
}
```

Response:

If username is not found, returns 404 with no body.

```JSON
{
  "token": "jwt token",
  "timeout": -1 // Minutes remaining until timeout is done. -1 if no timeout.
}
```

### Authenticated requests

There are no requests provided by stupid-backend that requires authentication, but will check authentications if the `token` query is given and extension on stupid-backend will have access to some basic info about the user. Ex:

> GET: /getdata?key={api_key}&token={jwt_token}

If a token is present, but the token is invalid (expired or otherwise), returns 401.

## Unauthenticated Apps

You can add an app as an `UnKeyedApp` that doesn't require an API key. This does not have an default functions, but requests will be forwarded to the app. When using both `KeyedApp` and `UnKeyedApp`, the `KeyedApp`s will have priority.

### Requests

> ANY: /{appID}/
> ANY: /{alternateName}/

## TODO

This libary is yet unfinished, and still needs a couple things.

- Change passwords.
  - De-authorize JWT tokens when this is done.
- Allow for third-party logins
  - Primarily want Google 0Auth, but implemented in a way that others could be added later on.
- Provide more pre-made db's
- Proper tests
  - Could cause less headaches for me in the future.
- Build a dashboard
  - Add the necessary APIs for access to this info.
