package prog

import (
	"fmt"
	"sort"

	_ "github.com/lib/pq"

	"test/fileobject"
)

const (
	rasterpath = "/data/raster/"
)

/*
Images with GCP (corner values?) need to be transformed into a image with valid projection.
https://stackoverflow.com/questions/63874859/convert-png-to-geotiff-with-gdal
*/

/*
insert into raster_geoms (location, src_srs, datetime, product, geom) values ('loc1','somesrc','2022-01-01T10:22','prod', ST_GeomFromText('MULTIPOLYGON (((24.0538073758627 37.7378703845543,24.0539640433116 37.7379275964686,24.0539663216762 37.7378701532806,24.0538073758627 37.7378703845543)))'));
*/

/*
type Web struct {
	Url  string `yaml:"url"`
	Type string `yaml:"type"`
}

type Local struct {
	Path string `yaml:"path"`
	Size int64  `yaml:"size"`
}

type Geo struct {
	//SpatialRef string  `yaml:"spatialRef"`
	Shape *fileobject.Shape `yaml:"shape"`
	SRS   string            `yaml:"srs"`
}

type PSQL struct {
	Host string
	User string
	Pass string
	DB   string
	Port string
}

func (p *PSQL) GetConnectionString() string {
	conn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		Psql.Host, Psql.User, Psql.Pass, Psql.DB, Psql.Port)
	return conn
}

func (p *PSQL) TryFill() error {
	p.Host = os.Getenv("PSQL_HOST")
	p.User = os.Getenv("PSQL_USER")
	p.DB = os.Getenv("PSQL_DB")
	p.Pass = os.Getenv("PSQL_PASS")
	p.Port = os.Getenv("PSQL_PORT")

	if p.Host == "" {
		p.Host = "postgresql.default.svc.cluster.local"
	}
	if p.User == "" {
		p.User = "postgres"
	}
	if p.DB == "" {
		p.DB = "postgres"
	}

	if p.Host == "" || p.User == "" || p.DB == "" || p.Pass == "" {
		return errors.New("Unable to grap credentials for PSQL")
	}
	if p.Port == "" {
		p.Port = "5432"
	}
	return nil
}

var Psql PSQL

func GetLocation(obj *fileobject.ConsumerObject) string {
	filename := filepath.Base(obj.File.Local.Path)
	return filepath.Join(rasterpath, obj.Meta.Product, filename)
}

func PushToPSQL(location string, obj *fileobject.ConsumerObject, passw string) error {
	fmt.Println("Using password: ", passw)

	conn := Psql.GetConnectionString()
	fmt.Println("Connecting to db: ", conn)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Conenction established")

	shape := obj.Geo.Shape

	npts := shape.Npts

	pair := make([]string, npts)

	for idx := 0; idx < npts; idx++ {
		pair[idx] = fmt.Sprintf("%v %v", shape.X_geo[idx], shape.Y_geo[idx])
	}
	pairs := strings.Join(pair[:], ",")

	schema, err := database.SanitizeSchemaName("project_" + obj.Meta.Project)
	if err != nil {
		return err
	}

	uuid := obj.Meta.UUID.String()
	{
		cmd := strings.ReplaceAll("INSERT INTO {{SCHEMA}}.files (filename, metadata) VALUES ($1, $2) RETURNING uuid;", "{{SCHEMA}}", schema)
		fmt.Println("Inserting into files:", cmd)
		row := db.QueryRow(cmd, obj.File.Local.Path, "{}")
		err = row.Scan(&uuid)
		if err != nil {
			return err
		}
	}

	timestr := obj.GetTimeStr()
	//cmd := fmt.Sprintf("insert into project_%s.raster_geoms (location, src_srs, datetime, product, geom) values ('%s','%s','%s','%s', ST_GeomFromText('MULTIPOLYGON (((%s)))')) returning uuid;",
	cmd := fmt.Sprintf("insert into project_%s.raster_geoms (uuid, location, src_srs, datetime, product, geom) values ('%s','%s','%s','%s','%s', ST_GeomFromText('MULTIPOLYGON (((%s)))'));",
		obj.Meta.Project, uuid, location, obj.Geo.SRS, timestr, obj.Meta.Product, pairs)
	fmt.Println("Executing query")
	fmt.Println(cmd)
	//res, err := db.Exec(cmd)

	//row := db.QueryRow(cmd)
	_, err = db.Exec(cmd)

	//return row.Scan(&obj.UUID)
	return err

}
*/

/*
func downloadFile(out *os.File, url string) (err error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	fmt.Println("Writing the body to file: ", out.Name())
	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}*/

/*
func ProcessRequest(obj *fileobject.ConsumerObject) error {

		var imagefilepath string
		var fileext string
		if obj.File.Web != nil {

			outfile, err := ioutil.TempFile("", "consumer-*")
			if err != nil {
				return err
			}
			fmt.Println("Downloading file")
			err = downloadFile(outfile, obj.File.Web.Url)
			if err != nil {
				return err
			}
			fileext = obj.File.Web.Type

			imagefilepath = outfile.Name()
			fmt.Println("Download successful to ", imagefilepath)

			defer os.Remove(outfile.Name())

		} else if len(obj.File.Binary) > 0 {
			buf := obj.File.Binary

			outfile, err := ioutil.TempFile("", "consumer-*")
			if err != nil {
				return err
			}
			outfile.Write(buf)
			fmt.Println("Written to file")

			obj.File.Binary = nil
			imagefilepath = outfile.Name()
		}

		if obj.File.Local != nil {
			imagefilepath = obj.File.Local.Path
		}
		fmt.Println("Filepath: ", imagefilepath)
		if obj.Timestamp == nil {
			fmt.Println("Trying to get time from file:", imagefilepath)
			ds, err := gdal.Open(imagefilepath, gdal.ReadOnly)
			if err != nil {
				return err
			}
			defer ds.Close()

			dateTimeMetadata := ds.MetadataItem("TIFFTAG_DATETIME", "")
			layout := "2006:01:02 15:04:05"
			tmp, err := time.Parse(layout, dateTimeMetadata)
			if err == nil {
				obj.Timestamp = &tmp
				fmt.Println("Extracted time: ", obj.Timestamp)
			} else {
				fmt.Println("Error parsing time:", err)
			}
		}
		if obj.Geo.Shape == nil {
			fmt.Println("Getting shape from:  ", imagefilepath)
			shape, err := NewShape4325(imagefilepath)
			if err != nil {
				return err
			}
			obj.Geo.Shape = shape

			ds, err := gdal.Open(imagefilepath, gdal.ReadOnly)
			if err != nil {
				return err
			}
			defer ds.Close()

			source := gdal.CreateSpatialReference(ds.Projection())
			obj.Geo.SRS, err = source.ToProj4()
			if err != nil {
				return err
			}
		}

		//timestamp := obj.Timestamp
		uuid := uuid.New()

		//filename := fmt.Sprintf("%s_%s.%s", timestamp.UTC().Format("20060102T150405"), uuid.String(), fileext)
		filename := fmt.Sprintf("%s.%s", uuid.String(), fileext)

		data_root := utils.GetDataRoot()

		//data_root := filepath.Join(data_root, "raster")
		reldir := filepath.Join("raster", obj.Project, obj.Product)
		relpath := filepath.Join(reldir, filename)
		//relpath := path.Join(obj.Project, obj.Product, filename)
		//fullpath := path.Join(raster_root, relpath)

		err := PushToPSQL(relpath, obj, os.Getenv("PGPASSWORD"))
		if err != nil {
			return err
		}
		fmt.Println("Push to SQL successful")

		fmt.Println("Moving " + imagefilepath + " => " + filepath.Join(data_root, relpath))
		//err = os.MkdirAll(path.Join(raster_root, obj.Project, obj.Product), os.ModePerm)
		err = os.MkdirAll(filepath.Join(data_root, reldir), os.ModePerm)
		if err != nil {
			return err
		}
		//err = utils.CopyFile(imagefilepath, fullpath)
		err = utils.CopyFile(imagefilepath, filepath.Join(data_root, relpath))
		if err != nil {
			return err
		}

		//fmt.Println(GetUrl(obj))
		return nil
	}
*/
func GetWMSUrl(obj *fileobject.ConsumerObject, host string) string {

	tmpX := append(make([]float64, 0, len(obj.Geo.Shape.X_geo)), obj.Geo.Shape.X_geo...)
	tmpY := append(make([]float64, 0, len(obj.Geo.Shape.Y_geo)), obj.Geo.Shape.Y_geo...)

	sort.Float64s(tmpX)
	sort.Float64s(tmpY)

	ll_lat := tmpY[0]
	ll_lon := tmpX[0]

	ur_lat := tmpY[len(tmpY)-1]
	ur_lon := tmpX[len(tmpX)-1]

	bbox := fmt.Sprintf("%v,%v,%v,%v", ll_lat, ll_lon, ur_lat, ur_lon)
	//http://localhost:9080/?map=/mapfiles/vedur/rasters.map&SERVICE=WMS&VERSION=1.3.0&REQUEST=GetMap&BBOX=44.67992280194039,-48.1384265209119,57.90268443794766,-2.745601176530886&CRS=EPSG:4326&WIDTH=1024&HEIGHT=768&LAYERS=viirs-granule-true-color&STYLES=,&CLASSGROUP=black&FORMAT=image/png&TRANSPARENT=true
	//url := fmt.Sprintf("http://localhost:9080/cgi-bin/mapserv?program=mapserv&SERVICE=WMS&VERSION=1.3.0&REQUEST=GetMap&BBOX=%s&CRS=EPSG:4326&WIDTH=1024&HEIGHT=768&LAYERS=%s&STYLES=,&CLASSGROUP=black&FORMAT=image/png&TRANSPARENT=true&TIME=%s", bbox, obj.Product, obj.Timestamp.UTC().Format(time.RFC3339))

	ret := fmt.Sprintf("%s/services/wms?map=/mapfiles/%s/rasters.map&SERVICE=WMS&VERSION=1.3.0&REQUEST=GetMap&BBOX=%s&CRS=EPSG:4326&WIDTH=1024&HEIGHT=768&LAYERS=%s&STYLES=,&CLASSGROUP=black&FORMAT=image/png&TRANSPARENT=true&TIME=%s", host, obj.Meta.Project, bbox, obj.Meta.Product, obj.GetTimeStr())
	return ret
}
func Run() error {

	/*filename := os.Args[1]
	fmt.Println("Processing file: ", filename)
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	obj := fileobject.ConsumerObject{}
	fmt.Println("Unmarshalling")
	err = yaml.Unmarshal(content, &obj)
	if err != nil {
		return err
	}
	fmt.Println("Filling psql struct")
	Psql.TryFill()
	fmt.Println("running: process request")
	err = ProcessRequest(&obj)
	if err != nil {
		log.Fatal(err)
	}*/
	return nil
}
