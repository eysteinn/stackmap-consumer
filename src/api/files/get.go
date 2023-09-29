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

	if product == "" {
		cmd := "select uuid, product from " + schema + ".raster_geoms;"
		rows, err = db.Query(cmd)
	} else {
		rows, err = db.Query("SELECT uuid, product FROM "+schema+".raster_geoms WHERE product = $1;", product)
	}
	if err != nil {
		fmt.Println(err)
		return files, err
	}
	defer rows.Close()

	for rows.Next() {
		file := FileInfo{}
		err = rows.Scan(&file.UUID, &file.Product)
		if err != nil {
			return files, err
		}
		fmt.Println("UUID:", file.UUID, "\tProduct:", file.Product)
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
