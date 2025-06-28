package models

import (
	"fmt"

	"github.com/jbrodriguez/mlog"
)

type Group struct {
	ID int64
	// Group_id int8
	Name        string
	Description string
}

func (g *Group) String() string { return g.Name }

func (g *Group) Save() {
	g1 := GetGroup(g.Name)
	DB := GetDB("")
	defer DB.Close()
	if g1 == nil {
		tx, _ := DB.Begin()
		res, e := tx.Exec(`insert into ngroup(
			name,
			description
			) values($1, $2)`, g.Name, g.Description)
		if e != nil {
			tx.Rollback()
			mlog.FatalIfError(fmt.Errorf("can not insert group %s - %s", g.Name, e.Error()))
		}
		g.ID, _ = res.LastInsertId()
		tx.Commit()
	}
}

func GetGroup(name string) *Group {
	DB := GetDB("")
	defer DB.Close()
	g := &Group{}
	if e := DB.QueryRow(`SELECT
	id,
	name,
	description
	FROM ngroup where name = $1`, name).Scan(&g.ID, &g.Name, &g.Description); e != nil {
		mlog.Warning("group '%s' not found - %v\n", name, e)
		return nil
	}
	return g
}

func GetGroupByID(id int64) *Group {
	DB := GetDB("")
	defer DB.Close()
	g := &Group{}
	if e := DB.QueryRow(`SELECT
	id,
	name,
	description
	FROM ngroup where id = $1`, id).Scan(&g.ID, &g.Name, &g.Description); e != nil {
		mlog.Warning("group id %d not found - %v\n", id, e)
		return nil
	}
	return g
}

func GetAllGroups() []*Group {
	DB := GetDB("")
	defer DB.Close()
	rows, _ := DB.Query(`SELECT
		id,
		name,
		description
	FROM ngroup g`)
	defer rows.Close()
	var o []*Group

	for rows.Next() {
		gr := &Group{}
		rows.Scan(&gr.ID, &gr.Name, &gr.Description)
		o = append(o, gr)
	}
	return o
}
