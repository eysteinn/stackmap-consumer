package fileobject

import (
	"time"

	"github.com/lukeroth/gdal"
)

func ExtractGeotiffTimestampTag(filepath string) (time.Time, error) {
	t := time.Time{}
	ds, err := gdal.Open(filepath, gdal.ReadOnly)
	if err != nil {
		return t, err
	}
	defer ds.Close()

	dateTimeMetadata := ds.MetadataItem("TIFFTAG_DATETIME", "")
	layout := "2006:01:02 15:04:05"
	t, err = time.Parse(layout, dateTimeMetadata)
	return t, err
}
