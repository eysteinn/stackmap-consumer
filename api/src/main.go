package main

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

	"database/sql"

	"github.com/cridenour/go-postgis"

	"github.com/lukeroth/gdal"

	"test/utils"

	"gopkg.in/yaml.v3"
)

const (
	rasterpath = "/data/raster/"
)

/*
insert into raster_geoms (location, src_srs, datetime, product, geom) values ('loc1','somesrc','2022-01-01T10:22','prod', ST_GeomFromText('MULTIPOLYGON (((24.0538073758627 37.7378703845543,24.0539640433116 37.7379275964686,24.0539663216762 37.7378701532806,24.0538073758627 37.7378703845543)))'));
*/

type ConsumerObject struct {
	Product   string    `yaml:"product"`
	Timestamp time.Time `yaml:"timestamp"`
	File      struct {
		Local *struct {
			Path string `yaml:"path"`
			Size int    `yaml:"size"`
		} `yaml:"local"`
		Web *struct {
			Url  string `yaml:"url"`
			Type string `yaml:"type"`
		} `yaml:"web"`
	}
	Geo struct {
		//SpatialRef string  `yaml:"spatialRef"`
		Shape *Shape `yaml:"shape"`
		SRS   string `yaml:"srs"`
	} `yaml:"geo"`
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
		psql.Host, psql.User, psql.Pass, psql.DB, psql.Port)
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

var psql PSQL

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
	//inputFile := "/app/data/viirs-granule-true-color_20221123T142204.tiff"
	ds, err := gdal.Open(inputFile, gdal.ReadOnly)
	if err != nil {
		//log.Fatal(err)
		return nil, err
	}
	defer ds.Close()

	tfm := ds.GeoTransform()

	source := gdal.CreateSpatialReference(ds.Projection()) // .SpatialReference()
	defer source.Release()

	target := gdal.CreateSpatialReference("") //.SpatialReference{}
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

	/*for idx := 0; idx < npts; idx++ {
		fmt.Println("(X, Y) = ", xpoints[idx], " ", ypoints[idx])
	}*/

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
	//fmt.Println(ds.GetGeoTransform())
}

/*func CreateSQLCmd(location string, obj *ConsumerObject) string { // src string, timestamp time.Time, product string, shape *Shape) string {
	shape := obj.Geo.Shape

	npts := shape.npts

	pair := make([]string, npts)

	for idx := 0; idx < npts; idx++ {
		pair[idx] = fmt.Sprintf("%v %v", shape.X_geo[idx], shape.Y_geo[idx])
	}
	pairs := strings.Join(pair[:], ", ")
	//values := fmt.Sprintf("'%s', '%s',)
	cmd := fmt.Sprintf("insert into raster_geoms (location, src_srs, datetime, product, geom) values ('%s','%s','%s','%s', ST_GeomFromText('MULTIPOLYGON (((%s)))'));",
		location, obj.Geo.SRS, obj.Timestamp.UTC().Format(time.RFC3339), obj.Product, pairs)
	//+pairs +")))'));"

	return cmd
}*/

func test1() {
	point := postgis.PointS{4326, -84.5014, 39.1064}
	fmt.Println(">", point)

	/*passw, ok := os.LookupEnv("POSTGRES_PASSWORD")
	if !ok {
		log.Fatal(errors.New("Missing POSTGRES_PASSWORD in environment"))
	}*/
	passw := ""
	conn := fmt.Sprintf("host=localhost user=postgres password=%s dbname=postgres port=9920 sslmode=disable", passw)

	db, err := sql.Open("postgres", conn)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	var newPoint postgis.PointS

	db.QueryRow("SELECT GeomFromEWKB($1);", point).Scan(&newPoint)
	fmt.Printf(">>%v  %v\n", point, newPoint)
}
func GetLocation(obj *ConsumerObject) string {
	filename := filepath.Base(obj.File.Local.Path)
	return filepath.Join(rasterpath, obj.Product, filename)
}

func PushToPSQL(location string, obj *ConsumerObject, passw string) error {
	//cmd := CreateSQLCmd(GetLocation(obj), obj)
	//fmt.Println(cmd)
	fmt.Println("Using password: ", passw)

	//conn := fmt.Sprintf("host=localhost user=postgres password=%s dbname=postgres port=9920 sslmode=disable", passw)
	//conn := fmt.Sprintf("host=localhost user=postgres password=%s dbname=postgres port=9920 sslmode=disable", passw)
	conn := psql.GetConnectionString()
	fmt.Println("Connecting to db: ", conn)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Trying ping")
	/*err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Ping works!!!")*/

	shape := obj.Geo.Shape

	npts := shape.npts

	pair := make([]string, npts)

	//points := make([]postgis.Point, npts)

	for idx := 0; idx < npts; idx++ {
		pair[idx] = fmt.Sprintf("%v %v", shape.X_geo[idx], shape.Y_geo[idx])
		//points[idx] = postgis.Point{X: shape.X_geo[idx], Y: shape.Y_geo[idx]}
	}
	pairs := strings.Join(pair[:], ",")
	//values := fmt.Sprintf("'%s', '%s',)
	//fmt.Println("Pairs: ", pairs)
	//location := GetLocation(obj)

	cmd := fmt.Sprintf("insert into raster_geoms (location, src_srs, datetime, product, geom) values ('%s','%s','%s','%s', ST_GeomFromText('MULTIPOLYGON (((%s)))'));",
		location, obj.Geo.SRS, obj.Timestamp.UTC().Format(time.RFC3339), obj.Product, pairs)
	//cmd := "insert into raster_geoms (location, src_srs, datetime, product, geom) values ($1, $2, $3, $4, ST_GeomFromText('MULTIPOLYGON ((($5)))'))"

	//fmt.Println("Cmd: ", cmd)
	res, err := db.Exec(cmd) //, location, obj.Geo.SRS, obj.Timestamp.UTC().Format(time.RFC3339), obj.Product, pairs)
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
	/*cmd := fmt.Sprintf("insert into raster_geoms (location, src_srs, datetime, product, geom) values ('%s','%s','%s','%s', ST_GeomFromText('MULTIPOLYGON (((%s)))'));",
	location, obj.Geo.SRS, obj.Timestamp.UTC().Format(time.RFC3339), obj.Product, pairs)*/
	//+pairs +")))'));"

	/*	var newPoint postgis.PointS

		db.QueryRow("SELECT GeomFromEWKB($1);", point).Scan(&newPoint)
		fmt.Printf(">>%v  %v\n", point, newPoint)*/
	return nil
}
func downloadFile(out *os.File, url string) (err error) {

	// Create the file
	/*out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()*/

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

func ProcessRequest() error {
	content, err := os.ReadFile("request.yaml")
	if err != nil {
		return err
	}

	//fmt.Println(string(content), "\n\n\n")
	obj := ConsumerObject{}

	err = yaml.Unmarshal(content, &obj)
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
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

		//obj.File.Local.Path = outfile.Name()
		filepath = outfile.Name()
		fmt.Println("Download successful to ", filepath)

		defer os.Remove(outfile.Name())
		/*fmt.Println(outfile.Name())
		os.Exit(0)*/

	}

	if obj.File.Local != nil {
		filepath = obj.File.Local.Path
	}
	fmt.Println("Filepath: ", filepath)
	if obj.Geo.Shape == nil { //|| len(obj.Geo.Shapes) == 0 {
		fmt.Println("Getting shape from:  ", filepath)
		shape, err := NewShape4325(filepath) //obj.File.Local.Path)
		if err != nil {
			return err
		}
		//obj.Geo.Shapes = []Shape{*shape}
		obj.Geo.Shape = shape
	}

	fmt.Println(obj)

	//product := obj.Product
	timestamp := obj.Timestamp
	uuid := uuid.New()

	filename := fmt.Sprintf("%s_%s.%s", timestamp.UTC().Format("20060102T150405"), uuid.String(), fileext)

	raster_root := os.Getenv("RASTER_ROOT")
	if raster_root == "" {
		raster_root = "/data/raster/"
	}
	relpath := path.Join(obj.Product, filename)
	fullpath := path.Join(raster_root, relpath)
	err = PushToPSQL(relpath, &obj, os.Getenv("PGPASSWORD"))
	if err != nil {
		return err
	}
	fmt.Println("Push to SQL successful")

	fmt.Println("Moving " + filepath + " => " + fullpath)
	err = os.MkdirAll(path.Join(raster_root, obj.Product), os.ModePerm)
	if err != nil {
		return err
	}
	/*err = os.Rename(obj.File.Local.Path, outfile)
	if err != nil {
		return err
	}*/
	err = utils.CopyFile(filepath, fullpath)
	if err != nil {
		return err
	}

	fmt.Println(GetUrl(&obj))
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
	//host := fmt.Sprintf("http://localhost:9080/cgi-bin/mapserv?program=mapserv&SERVICE=WMS&VERSION=1.3.0&REQUEST=GetMap&BBOX=-1000000.0,-1000000.0,1600000.0,2000000.0&CRS=AUTO2:42003,1,-18.0,55.0&WIDTH=1024&HEIGHT=768&LAYERS=%s&STYLES=,&CLASSGROUP=black&FORMAT=image/png&TRANSPARENT=true&TIME=%s", obj.Product, obj.Timestamp.UTC().Format(time.RFC3339))
	url := fmt.Sprintf("http://localhost:9080/cgi-bin/mapserv?program=mapserv&SERVICE=WMS&VERSION=1.3.0&REQUEST=GetMap&BBOX=%s&CRS=AUTO2:42003,1,-18.0,55.0&WIDTH=1024&HEIGHT=768&LAYERS=%s&STYLES=,&CLASSGROUP=black&FORMAT=image/png&TRANSPARENT=true&TIME=%s", bbox, obj.Product, obj.Timestamp.UTC().Format(time.RFC3339))
	return url
}
func main() {
	fmt.Println("Hey")
	/*_, err := gdal.GetDriverByName("GTiff")
	if err != nil {
		log.Fatal(err.Error())
	}*/
	/*inputFile := "/app/data/viirs-granule-true-color_20221123T142204.tiff"

	shape, err := NewShape4325(inputFile)
	sql := CreateSQLCmd(shape, time.Now(), "viirs-granule-true-color")
	if err != nil {
		log.Fatal(err)
	}*/
	//test1()

	psql.TryFill()

	err := ProcessRequest()
	if err != nil {
		log.Fatal(err)
	}

}
