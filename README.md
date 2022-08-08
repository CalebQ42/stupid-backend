# stupid-backend

A stupid backend to test things out. I don't actually know what I'm doing, but why not :P. Planning on using Go with MongoDB for database duties. Probably will make a dart/Flutter plugin at some point. Made primarily to eventually be used with [SWAssitant](https://github.com/CalebQ42/SWAssistant).

## Goals

- [ ] Provide a http.Handler that can be use on any port or address you feel like
  - [ ] Maybe provide a small default runner that just runs on a predefined port
- [ ] User count
- [ ] User accounts
  - [ ] I need to do some research on security to make sure I'm not leaking passwords everywhere
- [ ] User uploaded data
- [ ] Access application data
- [ ] Send crash reports
  - [ ] Potentially log the current "page" or action to be sent with the report
- [ ] A really bad looking default website (I'm don't use HTML, CSS, or JS very often).
  - [ ] I might just make this a Flutter app that can be deployed as a PWA.
- [ ] Protect unauthorized usage with a secret key. If the key is not present, turn off the features cleanly.
  - [ ] Allow open-source to be compiled without borking the entire app.
- [ ] Enable or disable as many (or as few) of the above features.
- [ ] Dart/Flutter plugin

## Current Features

LOL

## Data Model

This is all just an idea on how the data will be organized in the DB. Subject to change (just like everything else right now).

User:

```JSON
{
  _id: "uuid",
  lastConnect: 20220807, //Date in YYYYMMDD format. is !hasLogin and hasn't connected for over a month, record is deleted
  hasLogin: true, //If false, username and password will NOT be present, or will be empty. 
  username: "name",
  password: "hashed password", //I need to do research on security before I really set this part up...
  email: "email@email.com" //Probably won't be present or used for a while. Only present to be used in the furture for account recovery.
}
```

Application Data:

```JSON
{
  _id: "uuid",
  displayName: "name to be displayed to user",
  type: "type", //TBD by application. Suggestions include data, config.
  data: {} //Determined by the application and type.
}
```

User Data:

```JSON
{
  _id: "uuid",
  owner: "user id",
  readPerm: [ //Other users with permission to read the data
    "user id"
  ],
  writePerm: [ //Other users with permission to write the data
    "user id"
  ],
  data: {} //Determined by the application.
}
```

Crash reports:

```JSON
{
  _id: "first line of error", //This is to attempt to group together multiple instances of the same error. Possible could become the _id. Possibly might need to be something different.
  reports: [
    {
      _id: "uuid", //This is generated at time of crash. Prevents double sending of crash reports (such as if the report needs to be sent on next app launch)
      stack: "stacktrace",
      action: "characters" //What page or activity the user was doing
    }
  ]
}
```
