package files

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"test/database"

	"github.com/go-chi/chi/v5"
)

func DeleteFile(project string, uuid string) error {
	db, err := database.GetDB()
	if err != nil {
		log.Println(err)
		return err
	}

	schema, err := database.SanitizeSchemaName("project_" + project)
	if err != nil {
		return err
	}
	row := db.QueryRow("DELETE FROM "+schema+".raster_geoms WHERE uuid = $1 RETURNING location;", uuid)
	location := ""
	err = row.Scan(&location)
	if err != nil {
		return err
	}

	fmt.Println("Location:", location)
	return nil
}

func DeleteHandle(response http.ResponseWriter, request *http.Request) {

	/*project := request.Context().Value("project")
	uuid := request.Context().Value("uuid")*/
	project := chi.URLParam(request, "project")
	uuid := chi.URLParam(request, "uuid")

	fmt.Println("test")
	fmt.Println("Project:", project)
	fmt.Println("UUID:", uuid)

	resp := map[string]interface{}{}
	response.Header().Set("Content-Type", "application/json")
	resp["success"] = true
	resp["message"] = "project deleted succesfully"
	retcode := http.StatusOK
	resp["uuid"] = uuid
	resp["project"] = project

	err := DeleteFile(project, uuid)
	if err != nil {
		log.Println(err)
		resp["message"] = fmt.Sprint("could not delete file with uuid '" + uuid + "'")
		resp["success"] = false
		retcode = http.StatusBadRequest
	}

	response.WriteHeader(retcode)
	b, _ := json.Marshal(resp)
	response.Write(b)
	// First You begin a transaction with a call to db.Begin()
	//ctx := context.Background()

	/*tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
		return
	}

	rows, err := tx.QueryContext(ctx, "SELECT location FROM project_"+project+".raster_geoms WHERE uuid='"+uuid+"'")

	defer rows.Close()
	for rows.Next() {
		var location string
		if err := rows.Scan(&location); err != nil {
			//log.Fatal(err)
			log.Println(err)
			tx.Rollback()
			return
		}
		fmt.Println(uuid, location)
	}
	res, err := tx.ExecContext(ctx, "DELETE FROM project_"+project+".raster_geoms WHERE uuid='bc9a9dd2-5151-4b9a-a57b-a7bc08732438'")
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}
	lines, err := res.RowsAffected()
	tx.Commit()*/
	//log.Println("Deleted rows:", lines)
}
