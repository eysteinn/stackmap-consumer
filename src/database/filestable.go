package database

import (
	"strings"
)

func WriteToFilesTable(project string, uuid string, filename string, metadata string) error {

	db, err := GetDB()
	if err != nil {
		return err
	}
	schema, err := SanitizeSchemaName("project_" + project)
	if err != nil {
		return err
	}

	cmd := strings.ReplaceAll("INSERT INTO {{SCHEMA}}.files (uuid, filename, metadata) VALUES ($1, $2, $3) RETURNING uuid;", "{{SCHEMA}}", schema)
	//fmt.Println("Inserting into files:", cmd)
	_, err = db.Exec(cmd, uuid, filename, "{}")

	return err
}
