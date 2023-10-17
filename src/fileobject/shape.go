package fileobject

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/lukeroth/gdal"
)

type Shape struct {
	X_geo []float64
	Y_geo []float64
	Npts  int
}

func (s *Shape) GetCoordsString() string {
	pair := make([]string, s.Npts)

	for idx := 0; idx < s.Npts; idx++ {
		pair[idx] = fmt.Sprintf("%v %v", s.X_geo[idx], s.Y_geo[idx])
	}
	pairs := strings.Join(pair[:], ",")
	return pairs
}

func transf(gt [6]float64, x int, y int) (X_geo float64, Y_geo float64) {

	X_geo = gt[0] + float64(x)*gt[1] + float64(y)*gt[2]
	Y_geo = gt[3] + float64(x)*gt[4] + float64(y)*gt[5]
	return X_geo, Y_geo
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

	xpts := []float64{}
	ypts := []float64{}
	xmrg := int(0.05 * float64(ds.RasterXSize()))
	ymrg := int(0.05 * float64(ds.RasterYSize()))
	parts := 5

	var xtmp, ytmp float64

	xtmp, ytmp = transf(tfm, 0-xmrg, 0-ymrg)
	xpts = append(xpts, xtmp)
	ypts = append(ypts, ytmp)

	for i := 1; i < parts; i++ {
		dx := ds.RasterXSize() * i / parts
		xtmp, ytmp = transf(tfm, dx, 0-ymrg)
		xpts = append(xpts, xtmp)
		ypts = append(ypts, ytmp)
	}

	xtmp, ytmp = transf(tfm, ds.RasterXSize()+xmrg, 0-ymrg)
	xpts = append(xpts, xtmp)
	ypts = append(ypts, ytmp)

	for i := 1; i < parts; i++ {
		dy := ds.RasterYSize() * i / parts
		xtmp, ytmp = transf(tfm, ds.RasterXSize()+xmrg, dy)
		xpts = append(xpts, xtmp)
		ypts = append(ypts, ytmp)
	}

	xtmp, ytmp = transf(tfm, ds.RasterXSize()+xmrg, ds.RasterYSize()+ymrg)
	xpts = append(xpts, xtmp)
	ypts = append(ypts, ytmp)

	for i := 1; i < parts; i++ {
		dx := ds.RasterXSize() - ds.RasterXSize()*i/parts
		xtmp, ytmp = transf(tfm, dx, ds.RasterYSize()+ymrg)
		xpts = append(xpts, xtmp)
		ypts = append(ypts, ytmp)
	}

	xtmp, ytmp = transf(tfm, 0-xmrg, ds.RasterYSize()+ymrg)
	xpts = append(xpts, xtmp)
	ypts = append(ypts, ytmp)

	for i := 1; i < parts; i++ {
		dy := ds.RasterYSize() - ds.RasterYSize()*i/parts
		xtmp, ytmp = transf(tfm, 0-xmrg, dy)
		xpts = append(xpts, xtmp)
		ypts = append(ypts, ytmp)
	}

	xtmp, ytmp = transf(tfm, 0-xmrg, 0-ymrg)
	xpts = append(xpts, xtmp)
	ypts = append(ypts, ytmp)

	npts := len(xpts)
	xpoints := xpts
	ypoints := ypts

	zpoints := make([]float64, npts)
	/*const npts = 5
	xpoints := [npts]float64{}
	ypoints := [npts]float64{}

	// TODO:More points!
	xpoints[0], ypoints[0] = transf(tfm, 0, 0)
	xpoints[1], ypoints[1] = transf(tfm, ds.RasterXSize(), 0)
	xpoints[2], ypoints[2] = transf(tfm, ds.RasterXSize(), ds.RasterYSize())
	xpoints[3], ypoints[3] = transf(tfm, 0, ds.RasterYSize())
	xpoints[4], ypoints[4] = transf(tfm, 0, 0)

	zpoints := [npts]float64{}*/
	for idx := 0; idx < npts; idx++ {
		zpoints[idx] = 0
	}

	transform.Transform(npts, xpoints[:], ypoints[:], zpoints[:])
	return &Shape{
			X_geo: xpoints[:],
			Y_geo: ypoints[:],
			Npts:  npts,
		},
		nil
}

func (o *ConsumerObject) FillGeo() error {
	if o.File.Local == nil {
		return errors.New("Local file not available")
	}
	if o.File.Local.Path == "" {
		return errors.New("File is not existing locally")
	}

	filepath := o.File.Local.Path
	if o.Geo.Shape == nil {
		log.Println("Getting shape from:  ", filepath)
		shape, err := NewShape4325(filepath)
		if err != nil {
			return err
		}
		o.Geo.Shape = shape
	}

	if o.Geo.SRS == "" {
		ds, err := gdal.Open(filepath, gdal.ReadOnly)
		if err != nil {
			return err
		}
		defer ds.Close()

		source := gdal.CreateSpatialReference(ds.Projection())
		defer source.Destroy()
		o.Geo.SRS, err = source.ToProj4()
		if err != nil {
			return err
		}
	}
	return nil
}
