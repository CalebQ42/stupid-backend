package db

import (
	"strings"

	"github.com/CalebQ42/stupid-backend/pkg/crash"
	"github.com/google/uuid"
)

type fromTable struct {
	Table
}

func (f fromTable) AddCrash(c crash.Individual) error {
	first, _, _ := strings.Cut(c.Stack, "\n")
	var g crash.Group
	err := f.Find(map[string]any{
		"error": c.Error,
		"first": first,
	}, &g)
	if err == ErrNotFound {
		g = crash.Group{
			ID:        uuid.NewString(),
			Error:     c.Error,
			FirstLine: first,
			Crashes:   []crash.Individual{c},
		}
		_, err = f.Add(g)
		return err
	} else if err != nil {
		return err
	}
	for i := range g.Crashes {
		if g.Crashes[i].ID == c.ID {
			return nil
		}
	}
	g.Crashes = append(g.Crashes, c)
	return f.Update(g.ID, g)
}

// Crates a CrashTable from a table. Highly suggested to NOT use this.
// The returned CrashTable updates the entire crash group every time.
func TableToCrashTable(t Table) CrashTable {
	return fromTable{Table: t}
}
