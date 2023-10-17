package fileobject

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Web struct {
	Url  string `yaml:"url"`
	Type string `yaml:"type"`
}

type Local struct {
	Path string `yaml:"path"`
	//Name string `yaml:"name"`
	Size int64 `yaml:"size"`
}

type Geo struct {
	//SpatialRef string  `yaml:"spatialRef"`
	Shape *Shape `yaml:"shape"`
	SRS   string `yaml:"srs"`
}

type ConsumerObject struct {
	Meta struct {
		Product   string
		Project   string
		Timestamp *time.Time
		Filename  string
		UUID      uuid.UUID
	}
	File struct {
		Local *Local `yaml:"local"`
		//Web   *Web   `yaml:"web"`
		//Binary []byte `yaml:"-"`
		//Name string `yaml:"name"`
		/*Web *struct {
			Url  string `yaml:"url"`
			Type string `yaml:"type"`
		} `yaml:"web"`*/
	}
	Geo Geo `yaml:"geo"`
}

func (o *ConsumerObject) GetTimeStr() string {
	if o.Meta.Timestamp != nil {
		return o.Meta.Timestamp.UTC().Format(time.RFC3339)
	}
	return ""
}

func (o *ConsumerObject) Validate() error {
	if o.Meta.Product == "" {
		return errors.New("Missing valid product")
	}

	if o.Meta.Project == "" {
		return errors.New("Missing valid project")
	}

	/*if o.File.Web == nil && len(o.File.Binary) == 0 {
		return errors.New("No web file or binary data defined")
	}*/
	return nil
}
