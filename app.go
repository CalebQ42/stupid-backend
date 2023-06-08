package stupid

import "github.com/CalebQ42/stupid-backend/pkg/db"

type App interface {
	Logs() db.LogTable
	Crashes() db.CrashTable
	Extension(*Request) bool
	IgnoreOldVersionCrashes() bool
	LatestVersion() string
}
