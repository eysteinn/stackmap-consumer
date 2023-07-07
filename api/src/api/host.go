package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"test/prog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	amqp "github.com/rabbitmq/amqp091-go"
	"gopkg.in/yaml.v3"
)

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
	obj.Product = val.Get("product")
	obj.Project = val.Get("project")

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
func uploadPost(response http.ResponseWriter, request *http.Request) {
	obj := prog.ConsumerObject{}
	obj.Geo = prog.Geo{}

	request.ParseMultipartForm(10 << 20) // 10 MB
	fmt.Println("PostForm: ", request.PostForm)
	for k, v := range request.PostForm {
		fmt.Println(k, "  ", v)
	}

	multipartFormData := request.MultipartForm
	if multipartFormData == nil {
		fmt.Println("Form is nil")
		return
	}
	for k := range multipartFormData.File {
		fmt.Println("Key: ", k)
	}

	if files, ok := multipartFormData.File["request"]; ok && len(files) > 0 {
		file, err := files[0].Open()
		if err != nil {
			errorResp(err, response)
			return
		}
		defer file.Close()
		size, err := getFilesize(file)
		if err != nil {

			errorResp(err, response)
			return
		}
		buf := make([]byte, size)
		n, err := file.Read(buf)
		if err != nil {
			errorResp(err, response)
			return
		}
		fmt.Printf("Read %v bytes of request form", n)

		err = yaml.Unmarshal(buf, obj)
		if err != nil {
			errorResp(err, response)
			return
		}
		fmt.Println("Unmarshal request successfully")
	}

	err := ParseURL(&obj, request.URL.Query())
	if err != nil {
		errorResp(err, response)
		return
	}
	if files, ok := multipartFormData.File["image"]; ok && len(files) > 0 {
		//for k, v := range multipartFormData.File { //["attachments"] {
		//if k == "image" && len(v) > 0 {
		file, err := files[0].Open()
		if err != nil {
			errorResp(err, response)
			return
		}

		//uploadedFile, _ := v[0].Open()
		defer file.Close()
		size, err := getFilesize(file)

		fmt.Println("Filesize: ", size)
		if err != nil {
			errorResp(err, response)
			return
		}
		buf := make([]byte, size)

		fmt.Println(file)
		n, err := file.Read(buf)
		if err != nil {
			errorResp(err, response)
			return
		}
		obj.File.Binary = buf
		fmt.Printf("Read %v bytes", n)
		/*reader := bytes.NewReader(buf)
		_, imgfmt, err := image.Decode(reader)

		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Image format is: ", imgfmt)

		uuid := uuid.New()

		filename := fmt.Sprintf("%s_%s.%s", timestamp.UTC().Format("20060102T150405"), uuid.String(), fileext)
		*/
		// then use the single uploadedFile however you want
		// you may use its read method to get the file's bytes into a predefined slice,
		//here am just using an anonymous slice for the example
		//uploadedFile.Read([]byte{})

		//uploadedFile.Close()

		/*fmt.Println(v.Filename, ":", v.Size)
		uploadedFile, _ := v.Open()
		// then use the single uploadedFile however you want
		// you may use its read method to get the file's bytes into a predefined slice,
		//here am just using an anonymous slice for the example
		uploadedFile.Read([]byte{})
		uploadedFile.Close()*/
		//fmt.Println(k, ": ", v)
	}
	if err = obj.Validate(); err != nil {
		errorResp(err, response)
		return
	}

	//fmt.Println(string(b))
	fmt.Println("Processing request")
	err = prog.ProcessRequest(&obj)
	fmt.Println("Returning with error: ", err)
	if err != nil {
		errorResp(err, response)
		return
	}
	err = SendRabbitMQReport(&obj)
	if err != nil {
		errorResp(err, response)
		return
	}

	//respones.WriteHeader(http.StatusAccepted)
	obj.File.Binary = nil
	b, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		errorResp(err, response)
		return
	}
	response.WriteHeader(http.StatusCreated)
	response.Write(b)
	response.Write([]byte("\n"))
	response.Write([]byte(prog.GetUrl(&obj) + "\n"))
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
func delete(respones http.ResponseWriter, request *http.Request) {
}
func Run() {

	err := prog.Psql.TryFill()
	if err != nil {
		log.Println(err)
	}
	mux := chi.NewRouter()
	mux.Use(middleware.Logger)

	mux.Get("/upload", upload)
	mux.Post("/upload", uploadPost)
	/*mux := http.NewServeMux()
	mux.HandleFunc("/upload", upload)
	mux.HandleFunc("/delete", delete)*/
	fmt.Println("Starting to listen")
	err = http.ListenAndServe(":3333", mux)
	if err != nil {
		log.Panic(err)
	}
}
