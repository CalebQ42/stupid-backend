package stupid

import (
	"github.com/CalebQ42/stupid-backend/crash"
	"github.com/CalebQ42/stupid-backend/db"
)

// A simple API Key restricted app.
// Allows for basic functionality such as logging, crash reporting, and user authentication.
type KeyedApp interface {
	Logs() db.LogTable
	Crashes() db.CrashTable
}

// Extension of KeyedApp that allows filtering of Crash reports
type CrashFilteredApp interface {
	KeyedApp
	// Check to see if this crash should be accepted.
	AcceptCrash(crash.Individual) bool
}

// Extension of KeyedApp that allows for more request then the default.
type ExtendedApp interface {
	KeyedApp
	// Allows for handling non-standard request. If false, status code 400 (bad request) is sent.
	Extension(*Request) bool
}

type UnKeyedApp interface {
	// Handle a request that does not include an API key,
	// but who's Request.Path[0] is this app's name or AlternateName().
	// If false, status code 400 (bad request) is sent.
	HandleReqest(*Request) bool
}

type UnKeyedWithAlternateNameApp interface {
	UnKeyedApp
	// An alternate path name beside the app's name for requests.
	AlternateName() string
}
