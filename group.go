package djinn

import ()

// auth_group
type Group struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
}

func (group *Group) String() string {
	return group.Name
}
