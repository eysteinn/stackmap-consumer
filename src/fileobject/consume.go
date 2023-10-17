package fileobject

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"test/database"
	"test/utils"

	"github.com/google/uuid"
)

func Consume(obj *ConsumerObject) error {

	err := obj.FillGeo()
	if err != nil {
		return err
	}

	if obj.Meta.Timestamp == nil {
		log.Println("No time set, attempting to extract time from file tag ")
		tmp, err := ExtractGeotiffTimestampTag(obj.File.Local.Path)
		if err != nil {
			log.Println(fmt.Errorf("Unable to extract time: %w", err))
		}
		if err == nil {
			obj.Meta.Timestamp = &tmp
			log.Println("Time set to:", tmp)
		}
	}

	//file_uuid := uuid.New()
	obj.Meta.UUID = uuid.New()

	filename := obj.Meta.UUID.String() + ".tiff"

	data_root := utils.GetDataRoot()
	/*raster_root := os.Getenv("RASTER_ROOT")
	if raster_root == "" {
		raster_root = "/data/raster/"
	}*/
	//data_root := filepath.Join(data_root, "raster")
	reldir := filepath.Join("raster", obj.Meta.Project, obj.Meta.Product)
	relpath := filepath.Join(reldir, filename)

	fullpath := filepath.Join(data_root, relpath)
	fulldir := filepath.Join(data_root, reldir)

	err = database.WriteToFilesTable(obj.Meta.Project, obj.Meta.UUID.String(), relpath, "{}")
	if err != nil {
		return err
	}
	err = database.WriteToRasterTable(obj.Meta.Project, obj.Meta.Product, obj.Geo.Shape.GetCoordsString(), obj.Meta.UUID, relpath, *obj.Meta.Timestamp, obj.Geo.SRS)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fulldir, os.ModePerm)
	if err != nil {
		return err
	}
	//err = utils.CopyFile(imagefilepath, fullpath)
	//log.Println("Copygin file:", fullpath, "=>", fullpath)

	log.Println("Copyin file:", obj.File.Local.Path, "=>", fullpath)
	err = utils.CopyFile(fullpath, fullpath)
	if err != nil {
		return err
	}
	return nil
}
