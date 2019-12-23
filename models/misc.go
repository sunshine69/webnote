package models

import (
	"log"
)

type Group struct {
	ID int64
	Group_id int8
	Name string
	Description string
}

func (g *Group) String() string { return g.Name }

func (g *Group) Save() {
	g1 := GetGroup(g.Name)
	DB := GetDB(""); defer DB.Close()
	if g1 == nil {
		tx, _ := DB.Begin()
		res, e := tx.Exec(`insert into ngroup(
			group_id,
			name,
			description
			) values(int8($1), $2, $3)`, g.Group_id, g.Name, g.Description)
		if e != nil {
			tx.Rollback()
			log.Fatalf("ERROR can not insert group %s - %v\n", g.Name, e)
		}
		g.ID, _ = res.LastInsertId()
		tx.Commit()
	}
}

func GetGroup(name string) *Group {
	DB := GetDB(""); defer DB.Close()
	g := Group{}
	if e := DB.QueryRow(`SELECT
	id() as id,
	group_id,
	name,
	description
	FROM ngroup where name = $1`, name).Scan(&g.ID, &g.Group_id,  &g.Name,  &g.Description); e != nil {
		log.Printf("WARN group %s not found - %v\n", name, e)
		return nil
	}
	return &g
}

func GetGroupByID(id int8) *Group {
	DB := GetDB(""); defer DB.Close()
	g := Group{}
	if e := DB.QueryRow(`SELECT
	id() as id,
	group_id,
	name,
	description
	FROM ngroup where group_id = int8($1)`, id).Scan(&g.ID, &g.Group_id,  &g.Name,  &g.Description); e != nil {
		log.Printf("WARN group ID %d not found - %v\n", id, e)
		return nil
	}
	return &g
}