package djinn

import ()

// A generalized manager instance and base class of schemas specific managers
type Manager struct {
	db      *DB
	table   string
	columns []string
	primary string
}

// Does this column exist in the table columns?
// TODO It'd be nice to have a set
func (m *Manager) isValid(column string) bool {
	for _, col := range m.columns {
		if column == col {
			return true
		}
	}
	return false
}
