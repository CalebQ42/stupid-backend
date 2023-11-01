package stupid

import (
	"github.com/CalebQ42/stupid-backend/crash"
	"github.com/CalebQ42/stupid-backend/db"
)

type App interface {
	Logs() db.LogTable
	Crashes() db.CrashTable
}

type CrashFilteredApp interface {
	App
	AcceptCrash(crash.Individual) bool
}

type ExtendedApp interface {
	App
	Extension(*Request) bool
}
