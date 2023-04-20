package db

import (
	"errors"

	"github.com/CalebQ42/stupid-backend/pkg/crash"
)

var ErrNotFound error = errors.New("not found")

type Table interface {
	// Get's the item with the given key then marshals it into v.
	// If not found, return ErrNotFound
	Get(key string, v any) error
	// Attempt to find an item with the given values.
	// If found, marshall into v.
	// If query is successful, but can't be found, return ErrNotFound.
	Find(values map[string]any, v any) (err error)
	// Similiar to Find, but is expected to return multiple items.
	// v must be a slice.
	FindMany(values map[string]any, v any) (err error)
	// Add a new item to the table.
	// If the value given does not contain a key value (or _id for mongoDB), then you can get the value's key from the return.
	Add(v any) (key string, err error)
	// Update an existing key. This should do a full replacement of values.
	Update(key string, v any) error
	// Check if an item with the given key exists.
	Has(key string) (bool, error)
	// Check if an item with the given values exists.
	Contains(values map[string]any) (bool, error)
	// Delete the given item.
	Delete(key string) error
}

type CrashTable interface {
	Table
	// Add the individual crash to the given crash group.
	// Should be able to parse the crash's error and first line
	// and add it to the appropriate crash group.
	// Should also detect if the given individual crash has already been added.
	AddCrash(c crash.Individual) error
}

type UserTable interface {
	Table
	// Increment the "failed" field of the user with the given id by 1.
	IncrementFailed(id string) error
	// Same as IncrementFailed but also update the "lastTimeout" field with the given time.
	IncrementAndUpdateLastTimeout(id string, t int64) error
}
