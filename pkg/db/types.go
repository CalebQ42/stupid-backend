package db

import "github.com/CalebQ42/stupid-backend/pkg/crash"

type Table interface {
	// Get's the value with the given key then marshals it into v.
	Get(key string, v any) error
	// Add a new value to the table.
	// If the value given does not contain a key value (or _id for mongoDB), then you can get the value's key from the return.
	Add(v any) (key string, err error)
	// Update an existing key. This should do a full replacement of values.
	Update(key string, v any) error
	Has(key string) (bool, error)
}

type CrashTable interface {
	Table
	// Has the individual crash already been reported
	HasIndividualCrash(crashID string) (bool, error)
	// Get the ID for a crash group. If the crash group doesn't exist, a new one should be created.
	// Use crash.Group 
	GroupID(err string, firstLine string) (string, error)
	// Add the individual crash to the given crash group
	AddCrash(groupID string, c crash.Individual) error
}

type App struct {
	Logs    Table
	Crashes CrashTable
}
