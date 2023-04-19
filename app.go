package stupid

import "github.com/CalebQ42/stupid-backend/pkg/db"

type App struct {
	Logs    db.Table
	Crashes db.CrashTable
	//Extends this App's capabilities beyond the default. Returns whether the request was handled.
	Extension func(*Request) bool
}
