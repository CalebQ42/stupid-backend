# Stupid Backend

This is a project to make a very simple and highly extendable application backend. Uses a RESTful API with MongoDB.

## Planned Features

[ ] Log user connection (for user count within the last 30 days)
[ ] Get user count
[ ] (Maybe) Log features use. (So you know what features users are using)
[ ] Crash reporting
[ ] User accounts
[ ] Basic Data Storage without setup.
[X] Extend basic functionality via interfaces
[X] Very basic server as example

## Data structure

See [DB.md](DB.md)

## API

See [Requests.md](Requests.md)

## Future Plans

These are just musings on what I could potentially do in the future (after the planned features above are complete). Mainly here so I don't forget.

* Change things around so stupid.Apps don't have to use MongoDB
  * Possible change the actual backend so it doesn't require MongoDB either.
* Add abilities for backend access without having to look at the DB.
  * Probably build a Flutter app for easy access.
* Build a Flutter package for easy integration
  * Other platforms shouldn't be difficult, I'm just focused on Flutter ATM.
