package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"test/api/files"
	"test/prog"
	"time"

	_ "embed"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	amqp "github.com/rabbitmq/amqp091-go"
)

//go:embed post.html
var post_html []byte

type RabbitMQNotification struct {
	Product   string `yaml:"product"`
	Timestamp string `yaml:"timestamp"`
	Project   string `yaml:"project"`
}

func SendRabbitMQReport(obj *prog.ConsumerObject) error {

	// DEBUG THIS METHOD!!!!!!!
	fmt.Println("Sending to rabbitmq")
	notification := RabbitMQNotification{}
	notification.Product = obj.Product
	notification.Timestamp = obj.Timestamp.Format("20060102T150405")
	passw := os.Getenv("RABBITMQ_PASS")
	host := os.Getenv("RABBITMQ_HOST")
	url := fmt.Sprintf("amqp://user:%s@%s:5672/", passw, host)
	fmt.Println("Connecting to: ", url)
	/*dialer := amqp.DefaultDial(time.Duration(time.Second * 1))
	conn, err := dialer("tcp", url)*/
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"sat-stream", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	routing_key := "mapviewer.consume.success." + obj.Product

	body, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(ctx,
		"sat-stream", // exchange
		routing_key,  // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return err
	}
	log.Printf(" [x] Sent %s", string(body))
	return nil
}

func ParseURL(obj *prog.ConsumerObject, val url.Values) error {
	var err error

	timestr := val.Get("timestamp")
	if timestr != "" {
		obj.Timestamp, err = time.Parse("20060102T150405", timestr)
		if err != nil {
			return err
		}
	}
	if val.Has("product") {
		obj.Product = val.Get("product")
	}
	if val.Has("project") {
		obj.Project = val.Get("project")
	}

	file_web_url := val.Get("file.web.url")
	file_web_type := val.Get("file.web.type")
	if file_web_url != "" && file_web_type != "" {
		web := prog.Web{}
		web.Url = file_web_url
		web.Type = file_web_type
		obj.File.Web = &web
	}

	file_local_path := val.Get("file.local.path")
	if file_local_path != "" {
		local := prog.Local{}
		local.Path = val.Get("file.local.path")
		obj.File.Local = &local
	}
	return nil
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

func PostMultipartFormData(response http.ResponseWriter, request *http.Request) ([]prog.ConsumerObject, error) {

	ret := make([]prog.ConsumerObject, 0)

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
		for idx := range files {
			obj := prog.ConsumerObject{}
			obj.Geo = prog.Geo{}

			obj.Product = product
			obj.Project = project
			if timestr != "" {
				layout := "2006-01-02T15:04"
				if parsedTime, err := time.Parse(layout, timestr); err == nil {
					obj.Timestamp = parsedTime
				}
			}
			obj.File.Name = files[idx].Filename

			fmt.Println("Receiving file: ")
			fmt.Println("\tName: ", obj.File.Name)
			fmt.Println("\tTime: ", timestr, "  (", obj.Timestamp, ")")

			file, err := files[idx].Open()
			if err != nil {
				errorResp(err, response)
				return nil, err
			}

			defer file.Close()
			size, err := getFilesize(file)

			if err != nil {
				errorResp(err, response)
				return nil, err
			}
			fmt.Println("\tSize: ", size)

			buf := make([]byte, size)

			n, err := file.Read(buf)
			if err != nil {
				errorResp(err, response)
				return nil, err

			}
			obj.File.Binary = buf
			fmt.Printf("Read %v bytes\n", n)

			ret = append(ret, obj)
		}
	}
	return ret, nil
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

func addObjects(objs []prog.ConsumerObject, host string) ([]*ResultObj, error) {
	var err error
	ret := make([]*ResultObj, 0, len(objs))
	for _, obj := range objs {
		res := ResultObj{
			Product:  obj.Product,
			Project:  obj.Project,
			DateTime: &obj.Timestamp,
			Success:  false,
		}
		ret = append(ret, &res)
		if err = obj.Validate(); err != nil {
			res.Error = fmt.Sprint("%v", err)
			fmt.Println(err)
			continue
		}
		err = prog.ProcessRequest(&obj)
		if err != nil {
			//res.Error = "Internal error"
			fmt.Println(err)
			//continue
		}
		err = SendRabbitMQReport(&obj)
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

	}
	return ret, nil
}
func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}
func extractContentType(request *http.Request) string {
	contenttype := request.Header.Get("Content-Type")
	contenttype = strings.Split(contenttype, ";")[0]
	return contenttype
}

func uploadPost(response http.ResponseWriter, request *http.Request) {
	obj := prog.ConsumerObject{}
	obj.Geo = prog.Geo{}

	/*for k, v := range request.Header {
		fmt.Println("Header: Key= ", k, "   Value = ", v)
	}*/

	contenttype := request.Header.Get("Content-Type")
	contenttype = strings.Split(contenttype, ";")[0]
	fmt.Println("Content-Type= ", contenttype)
	switch contenttype {
	case "multipart/form-data":
		//if contenttype == "multipart/form-data" {
		fmt.Println("Received 'multipart/form-data' request")
		objs, err := PostMultipartFormData(response, request)
		if err != nil {
			errorResp(err, response)
			return
		}
		res, err := addObjects(objs, request.Host)
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
		return

	}

	errorResp(fmt.Errorf("Content type not supported. Use 'multipart/formdata'\n"), response)
	return

}

func html(respones http.ResponseWriter, request *http.Request) {
	fmt.Println("Serving POST HTML")
	// set header
	respones.Header().Set("Content-type", "text/html")
	respones.Write(post_html)
	//http.ServeFile(respones, request, "./post.html")
}
func upload(respones http.ResponseWriter, request *http.Request) {
	/*
		Example:
		 http://localhost:3333/upload?timestamp=20230307T003505&product=hrit-ash&file.web.url=https://brunnur.vedur.is/pub/eysteinn/viirs-granule-true-color_20230222T183723.tiff&file.web.type=tiff
	*/
	//request.Response.StatusCode = 404
	//log.Fatal(err)

	obj := prog.ConsumerObject{}
	err := ParseURL(&obj, request.URL.Query())
	if err != nil {
		request.Response.StatusCode = 404
		return
	}

	geo := prog.Geo{}
	obj.Geo = geo

	b, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))

	fmt.Println("Processing request")
	err = prog.ProcessRequest(&obj)
	fmt.Println("Returning with error: ", err)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sending to RabbitMQ")
	err = SendRabbitMQReport(&obj)
	if err != nil {
		log.Fatal(err)
	}
	//respones.WriteHeader(http.StatusAccepted)
	respones.Write(b)
}

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		next.ServeHTTP(w, r)
	})
}
func filesRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Post("/upload", uploadPost)
	router.Post("/projects/{project}/files", uploadPost)
	router.Get("/projects/{project}/files", files.GetHandler)
	router.Delete("/projects/{project}/files/{uuid}", files.DeleteHandle)
	return router
}

func Run() {

	err := prog.Psql.TryFill()
	if err != nil {
		log.Println(err)
	}
	router := chi.NewRouter()
	router.Use(Cors)
	router.Use(middleware.Logger)
	router.Mount("/api/v1/", filesRoutes())
	//mux.Get("/html/simplepost", html)

	/*router.Get("/upload", html)
	router.Post("/upload", uploadPost)*/
	/*router.Get("/projects/{project}/files", files.GetHandler)
	router.Delete("/projects/{project}/files/{uuid}", files.DeleteHandle)*/
	/*mux.Route("/projects/{project}/files", func(r chi.Router) {
		r.Delete("/{uuid}", files.DeleteHandle)
	})*/
	//mux.Delete("/files", deleteFile) make this dynamic /files/{uuid}
	/*mux := http.NewServeMux()
	mux.HandleFunc("/upload", upload)
	mux.HandleFunc("/delete", delete)*/
	hostaddress := "0.0.0.0:3333"
	fmt.Println("Starting to listen: " + hostaddress)
	err = http.ListenAndServe(hostaddress, router)
	if err != nil {
		log.Panic(err)
	}
}
