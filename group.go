package djinn

import ()

// auth_group
type Group struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
}

// TODO Many to many link to permissions

func (group *Group) String() string {
	return group.Name
}
