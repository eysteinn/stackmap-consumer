package database

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func WriteToRasterTable(project string, product string, coordinates string, file_uuid uuid.UUID, location string, timestamp time.Time, srs string) error {
	//func WriteToRasterTable(obj *fileobject.ConsumerObject) error {

	timestr := timestamp.UTC().Format(time.RFC3339)

	cmd := fmt.Sprintf("insert into project_%s.raster_geoms (uuid, location, src_srs, datetime, product, geom) values ('%s','%s','%s','%s','%s', ST_GeomFromText('MULTIPOLYGON (((%s)))'));",
		project, file_uuid.String(), location, srs, timestr, product, coordinates)
	//fmt.Println("Executing query")
	//fmt.Println(cmd)
	_, err := db.Exec(cmd)

	return err
}
