package files

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"test/database"
	"time"

	"github.com/go-chi/chi/v5"
)

type FileInfo struct {
	UUID      string     `json:"uuid,omitempty"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
	Product   string     `json:"product,omitempty"`
	Filename  string     `json:"filename,omitempty"`
}

func GetFiles(project string, product string) ([]FileInfo, error) {

	files := []FileInfo{}

	db, err := database.GetDB()
	if err != nil {
		return files, err
	}

	schema, err := database.SanitizeSchemaName("project_" + project)
	if err != nil {
		return files, err
	}

	var rows *sql.Rows
	/*
			select f.uuid, f.filename, r.product from project_vedur.raster_geoms r join project_vedur.files f on r.uuid=f.uuid whe
		re r.product='viirs-granule-true-color';
	*/
	if product == "" {
		//cmd := "select uuid, product from " + schema + ".raster_geoms;"
		cmd := "select f.uuid, f.filename, r.product, r.datetime from " + schema + ".raster_geoms r join project_vedur.files f on r.uuid=f.uuid;"
		rows, err = db.Query(cmd)
	} else {
		cmd := "select f.uuid, f.filename, r.product, r.datetime from " + schema + ".raster_geoms r join project_vedur.files f on r.uuid=f.uuid where r.product=$1;"
		rows, err = db.Query(cmd, product)

		//rows, err = db.Query("SELECT uuid, product FROM "+schema+".raster_geoms WHERE product = $1;", product)
	}
	if err != nil {
		fmt.Println(err)
		return files, err
	}
	defer rows.Close()

	for rows.Next() {
		file := FileInfo{}
		err = rows.Scan(&file.UUID, &file.Filename, &file.Product, &file.Timestamp)
		if err != nil {
			return files, err
		}
		fmt.Println("UUID:", file.UUID, "\tProduct:", file.Product, "\tTimestamp:", file.Timestamp)
		files = append(files, file)
	}
	fmt.Println("finished")
	return files, nil
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	project := chi.URLParam(r, "project")
	product := chi.URLParam(r, "product")
	fmt.Println("Project:", project)
	fmt.Println("Product:", product)

	resp := map[string]interface{}{}
	resp["success"] = true
	resp["message"] = "projects fetched succesfully"
	retcode := http.StatusOK

	files, err := GetFiles(project, product)
	if err != nil {
		fmt.Println(err)
		resp["success"] = false
		resp["message"] = "unable to get files"
		retcode = http.StatusBadRequest
	}
	resp["files"] = files
	w.WriteHeader(retcode)
	b, _ := json.Marshal(resp)
	w.Write(b)
}
