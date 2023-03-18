package prog

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"

	"database/sql"

	"github.com/lukeroth/gdal"

	"test/utils"
)

const (
	rasterpath = "/data/raster/"
)

/*
insert into raster_geoms (location, src_srs, datetime, product, geom) values ('loc1','somesrc','2022-01-01T10:22','prod', ST_GeomFromText('MULTIPOLYGON (((24.0538073758627 37.7378703845543,24.0539640433116 37.7379275964686,24.0539663216762 37.7378701532806,24.0538073758627 37.7378703845543)))'));
*/
type Web struct {
	Url  string `yaml:"url"`
	Type string `yaml:"type"`
}

type Local struct {
	Path string `yaml:"path"`
	Size int    `yaml:"size"`
}

type Geo struct {
	//SpatialRef string  `yaml:"spatialRef"`
	Shape *Shape `yaml:"shape"`
	SRS   string `yaml:"srs"`
}

type ConsumerObject struct {
	Product   string    `yaml:"product"`
	Timestamp time.Time `yaml:"timestamp"`
	File      struct {
		Local *Local `yaml:"local"`
		Web   *Web   `yaml:"web"`
		/*Web *struct {
			Url  string `yaml:"url"`
			Type string `yaml:"type"`
		} `yaml:"web"`*/
	}
	Geo Geo `yaml:"geo"`
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

func (p *PSQL) TryFill() {
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
		log.Fatal(errors.New("Unable to grap credentials for PSQL"))
	}
	if p.Port == "" {
		p.Port = "5432"
	}
}

var Psql PSQL

func transf(gt [6]float64, x int, y int) (X_geo float64, Y_geo float64) {

	X_geo = gt[0] + float64(x)*gt[1] + float64(y)*gt[2]
	Y_geo = gt[3] + float64(x)*gt[4] + float64(y)*gt[5]
	return X_geo, Y_geo
}

type Shape struct {
	X_geo []float64
	Y_geo []float64
	npts  int
}

func NewShape4325(inputFile string) (*Shape, error) {
	ds, err := gdal.Open(inputFile, gdal.ReadOnly)
	if err != nil {
		return nil, err
	}
	defer ds.Close()

	tfm := ds.GeoTransform()

	source := gdal.CreateSpatialReference(ds.Projection())
	defer source.Release()

	target := gdal.CreateSpatialReference("")
	target.FromEPSG(4326)
	defer target.Release()
	target.SetAxisMappingStrategy(gdal.OAMS_TraditionalGisOrder)

	transform := gdal.CreateCoordinateTransform(source, target)
	defer transform.Destroy()

	const npts = 5
	xpoints := [npts]float64{}
	ypoints := [npts]float64{}
	xpoints[0], ypoints[0] = transf(tfm, 0, 0)
	xpoints[1], ypoints[1] = transf(tfm, ds.RasterXSize(), 0)
	xpoints[2], ypoints[2] = transf(tfm, ds.RasterXSize(), ds.RasterYSize())
	xpoints[3], ypoints[3] = transf(tfm, 0, ds.RasterYSize())
	xpoints[4], ypoints[4] = transf(tfm, 0, 0)

	zpoints := [npts]float64{}
	for idx := 0; idx < npts; idx++ {
		zpoints[idx] = 0
	}

	transform.Transform(npts, xpoints[:], ypoints[:], zpoints[:])
	return &Shape{
			X_geo: xpoints[:],
			Y_geo: ypoints[:],
			npts:  npts,
		},
		nil
}

func GetLocation(obj *ConsumerObject) string {
	filename := filepath.Base(obj.File.Local.Path)
	return filepath.Join(rasterpath, obj.Product, filename)
}

func PushToPSQL(location string, obj *ConsumerObject, passw string) error {
	fmt.Println("Using password: ", passw)

	conn := Psql.GetConnectionString()
	fmt.Println("Connecting to db: ", conn)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		log.Fatal(err)
	}

	shape := obj.Geo.Shape

	npts := shape.npts

	pair := make([]string, npts)

	for idx := 0; idx < npts; idx++ {
		pair[idx] = fmt.Sprintf("%v %v", shape.X_geo[idx], shape.Y_geo[idx])
	}
	pairs := strings.Join(pair[:], ",")

	cmd := fmt.Sprintf("insert into raster_geoms (location, src_srs, datetime, product, geom) values ('%s','%s','%s','%s', ST_GeomFromText('MULTIPOLYGON (((%s)))'));",
		location, obj.Geo.SRS, obj.Timestamp.UTC().Format(time.RFC3339), obj.Product, pairs)
	res, err := db.Exec(cmd)
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}
	fmt.Println("Exec worked")
	ra, err := res.RowsAffected()
	fmt.Println("Rows affected: ", ra)

	if err != nil {
		return err
	}
	return nil
}
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
}

func ProcessRequest(obj *ConsumerObject) error {

	var filepath string
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

		filepath = outfile.Name()
		fmt.Println("Download successful to ", filepath)

		defer os.Remove(outfile.Name())

	}

	if obj.File.Local != nil {
		filepath = obj.File.Local.Path
	}
	fmt.Println("Filepath: ", filepath)
	if obj.Geo.Shape == nil {
		fmt.Println("Getting shape from:  ", filepath)
		shape, err := NewShape4325(filepath)
		if err != nil {
			return err
		}
		obj.Geo.Shape = shape

		ds, err := gdal.Open(filepath, gdal.ReadOnly)
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

	timestamp := obj.Timestamp
	uuid := uuid.New()

	filename := fmt.Sprintf("%s_%s.%s", timestamp.UTC().Format("20060102T150405"), uuid.String(), fileext)

	raster_root := os.Getenv("RASTER_ROOT")
	if raster_root == "" {
		raster_root = "/data/raster/"
	}
	relpath := path.Join(obj.Product, filename)
	fullpath := path.Join(raster_root, relpath)
	err := PushToPSQL(relpath, obj, os.Getenv("PGPASSWORD"))
	if err != nil {
		return err
	}
	fmt.Println("Push to SQL successful")

	fmt.Println("Moving " + filepath + " => " + fullpath)
	err = os.MkdirAll(path.Join(raster_root, obj.Product), os.ModePerm)
	if err != nil {
		return err
	}
	err = utils.CopyFile(filepath, fullpath)
	if err != nil {
		return err
	}

	fmt.Println(GetUrl(obj))
	return nil
}

func GetUrl(obj *ConsumerObject) string {

	tmpX := append(make([]float64, 0, len(obj.Geo.Shape.X_geo)), obj.Geo.Shape.X_geo...)
	tmpY := append(make([]float64, 0, len(obj.Geo.Shape.Y_geo)), obj.Geo.Shape.Y_geo...)

	sort.Float64s(tmpX)
	sort.Float64s(tmpY)

	ll_lat := tmpY[0]
	ll_lon := tmpX[0]

	ur_lat := tmpY[len(tmpY)-1]
	ur_lon := tmpX[len(tmpX)-1]

	bbox := fmt.Sprintf("%v,%v,%v,%v", ll_lat, ll_lon, ur_lat, ur_lon)
	url := fmt.Sprintf("http://localhost:9080/cgi-bin/mapserv?program=mapserv&SERVICE=WMS&VERSION=1.3.0&REQUEST=GetMap&BBOX=%s&CRS=EPSG:4326&WIDTH=1024&HEIGHT=768&LAYERS=%s&STYLES=,&CLASSGROUP=black&FORMAT=image/png&TRANSPARENT=true&TIME=%s", bbox, obj.Product, obj.Timestamp.UTC().Format(time.RFC3339))
	return url
}
func Run() error {

	filename := os.Args[1]
	fmt.Println("Processing file: ", filename)
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	obj := ConsumerObject{}
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
	}
	return nil
}
