# Stupid Backend

This is a project to make a very simple and highly extendable application backend. Uses a RESTful API with MongoDB.

## Features

- Log users
  - Logs an app's UUID, the date (not time) last connected, and the platform. Records are wiped if inactive for 30 days.
- Report Crashes
  - Crashes are grouped together for easy parsing
- User accounts
- Upload and download content
  - Allows for general application data and user specific data
  - Provides a basic, generic data system, but can be extended easily
- Add an extention to extend features
- Allow multiple apps to use the same backend.

## TODO

- [ ] Test all functions (Really should have done this before v0.1.0, but oh well).
- [ ] Create proper tests for said functions.
- [ ] Delete data in DefaultDataApp

## Data structure

See [DB.md](DB.md).

## API

See [OpenAPI document](api.yml).

## Future Plans

These are just musings on what I could potentially do in the future (after the planned features above are complete). Mainly here so I don't forget. Might or might not be implemented or completely adbandoned.

- Allow logins via 3rd parties (such as Google login)
- Change things around so `stupid.App`s don't have to use MongoDB
  - Possible change the actual backend so it doesn't require MongoDB either.
  - Probably would have to change current implementation to `MongoApp` or something similiar.
- Add abilities for backend access without having to look at the DB.
  - Probably build a Flutter app for easy access (and can be deployed as a web app).
- Build a Flutter package for easy integration.
  - Other platforms shouldn't be difficult, I'm just focused on Flutter ATM.
