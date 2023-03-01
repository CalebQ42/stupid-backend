package db

type DBTable interface {
	// Get's the value with the given key then marshals it into v.
	Get(key string, v any) error
	// Add a new value to the table.
	// If the value given does not contain a key value (or _id for mongoDB), then you can get the value's key from the return.
	Add(v any) (key string, err error)
	// Update an existing key. This should do a full replacement of values.
	Update(key string, v any) error
	Has(key string) (bool, error)
}
