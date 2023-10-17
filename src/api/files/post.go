package files

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"test/fileobject"
	"time"
)

/*func PostFile(obj *prog.ConsumerObject) (*ResultObj, error) {

	if err := obj.Validate(); err != nil {
		return nil, err
	}

	var err error
	res := ResultObj{
		Product:  obj.Product,
		Project:  obj.Project,
		DateTime: obj.Timestamp,
		Success:  false,
	}

	err = prog.ProcessRequest(&obj)
	if err != nil {
		//res.Error = "Internal error"
		fmt.Println(err)
		//continue
	}
	err = rabbitmq.SendRabbitMQReport(&obj)
	if err != nil {
		//res.Error = "Internal error"
		fmt.Println(err)
		//continue
	}
	res.Success = true
	res.UUID = obj.UUID
	res.WMS_url = prog.GetWMSUrl(&obj, host)
	fmt.Println(res.WMS_url)
	fmt.Println("HOST: ", host)
	res.Filename = obj.File.Name

	return ret, nil
}*/

func PostHandler(response http.ResponseWriter, request *http.Request) {
	log.Println("Post request")
	obj := fileobject.ConsumerObject{}
	obj.Geo = fileobject.Geo{}

	/*for k, v := range request.Header {
		fmt.Println("Header: Key= ", k, "   Value = ", v)
	}*/

	resp := map[string]interface{}{}
	resp["success"] = true
	resp["message"] = "projects fetched succesfully"
	retcode := http.StatusOK

	contenttype := request.Header.Get("Content-Type")
	contenttype = strings.Split(contenttype, ";")[0]
	log.Println("Content-Type= ", contenttype)
	switch contenttype {
	case "multipart/form-data":
		//if contenttype == "multipart/form-data" {
		objs, err := ParseFormData(request)
		if err != nil {
			fmt.Println(err)
			resp["success"] = false
			resp["message"] = "unable to get files"
			retcode = http.StatusBadRequest
		}

		err = fileobject.Consume(objs)
		if err != nil {
			fmt.Println(err)
			resp["success"] = false
			resp["message"] = "unable to get files"
			retcode = http.StatusBadRequest
		}

		/*res, err := addObjects(objs, request.Host)
		if err != nil {
			errorResp(fmt.Errorf("Internal error occured."), response)
			fmt.Println(err)
			return
		}

		//response.Write([]byte("Files processed"))

		//json.MarshalIndent(res, "", "\t")
		bytes, err := JSONMarshal(res)
		if err != nil {
			errorResp(fmt.Errorf("internal error occured.\n"), response)
			fmt.Println(err)
			return
		}
		response.WriteHeader(http.StatusCreated)
		response.Header().Set("content-type", "application/json")
		response.Write(bytes)
		return*/

	}

	response.WriteHeader(retcode)
	b, _ := json.Marshal(resp)
	response.Write(b)
	return

}

func ParseFormData(request *http.Request) (*fileobject.ConsumerObject, error) {

	//ret := make([]prog.ConsumerObject, 0)
	//ret := fileobject.ConsumerObject{}
	obj := fileobject.ConsumerObject{}

	request.ParseMultipartForm(10 << 20) // 10 MB

	fmt.Println("PostForm: ", request.PostForm)

	product := ""
	project := ""
	timestr := ""
	for k, v := range request.PostForm {
		if k == "product" && len(v) > 0 {
			product = v[0]
		}
		if k == "project" && len(v) > 0 {
			project = v[0]
		}
		if k == "time" && len(v) > 0 {
			timestr = v[0]
		}
		//fmt.Println("PostFormValueKey: ", k, "  ", v)
	}

	multipartFormData := request.MultipartForm

	if files, ok := multipartFormData.File["image"]; ok && len(files) > 0 {
		log.Println("Processing incoming multipart form data")
		fileheader := files[0]
		//for idx := range files {
		obj.Geo = fileobject.Geo{}

		obj.Meta.Product = product
		obj.Meta.Project = project
		if timestr != "" {
			layout := "2006-01-02T15:04"
			if parsedTime, err := time.Parse(layout, timestr); err == nil {
				obj.Meta.Timestamp = &parsedTime
			}
		}
		obj.Meta.Filename = fileheader.Filename

		file, err := fileheader.Open()
		if err != nil {
			return nil, err
		}

		defer file.Close()
		size, err := getFilesize(file)
		if err != nil {
			return nil, err
		}

		log.Printf("Receiving file:\n\tName: %s\n\tSize: %v\n", obj.Meta.Filename, size)

		buf := make([]byte, size)

		n, err := file.Read(buf)
		if err != nil {
			return nil, err

		}
		//obj.File.Binary = buf
		log.Printf("Read %v bytes\n", n)

		//file_uuid := uuid.New()
		//data_root := utils.GetDataRoot()

		//filename := path.Join(data_root, "raster", obj.Project, obj.Product, file_uuid.String()+".tiff")

		/*err = os.WriteFile(filename, buf, 0644)
		if err != nil {
			return nil, err
		}

		obj.File.Local.Path = filename*/
		tmpfile, err := os.CreateTemp("", "consumer-*")
		if err != nil {
			return nil, err
		}
		_, err = tmpfile.Write(buf)
		if err != nil {
			os.Remove(tmpfile.Name())
			return nil, err
		}
		log.Println("Binary written to:", tmpfile.Name())
		obj.File.Local = &fileobject.Local{
			Path: tmpfile.Name(),
			Size: size,
		}

		//ret = append(ret, obj)
		//}
	}
	return &obj, nil
}

func errorResp(err error, w http.ResponseWriter) { //} request *http.Request) {
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprint(err)))
		//request.Response.StatusCode = 404
		return
	}
}
func getFilesize(file multipart.File) (size int64, err error) {
	size, err = file.Seek(0, os.SEEK_END)
	if err != nil {
		return
	}
	_, err = file.Seek(0, os.SEEK_SET)
	return
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

type ResultObj struct {
	Product  string     `yaml:"product" json:"product"`
	Project  string     `yaml:"project" json:"project"`
	DateTime *time.Time `yaml:"time,omitempty" json:"time,omitempty"`
	WMS_url  string     `yaml:"wms_url" json:"wms_url"`
	WMTS_url string     `yaml:"wmts_url" json:"wmts_url"`
	Error    string     `yaml:"error,omitempty" json:"error,omitempty"`
	Success  bool       `yaml:"success" json:"success"`
	Filename string     `yaml:"filename" json:"filename"`
	UUID     string     `yaml:"uuid" json:"uuid"`
}

/*func addObjects(obj prog.ConsumerObject, host string) (*ResultObj, error) {

}*/
