package djinn

import ()

// auth_permission
type Permission struct {
	Id            int64  `db:"id"`
	Name          string `db:"name"`
	ContentTypeId int64  `db:"content_type_id"`
	Codename      string `db:"codename"`
}

func (p *Permission) String() string {
	return p.Name
}
