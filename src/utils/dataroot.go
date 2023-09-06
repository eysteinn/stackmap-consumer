package utils

import "os"

func GetDataRoot() string {
	data_root := os.Getenv("MAPDATA_ROOT")
	if data_root == "" {
		data_root = "/data/"
	}
	return data_root
}
