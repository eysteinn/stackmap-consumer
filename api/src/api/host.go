package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"test/prog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQNotification struct {
	Product   string `yaml:"product"`
	Timestamp string `yaml:"timestamp"`
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
func upload(respones http.ResponseWriter, request *http.Request) {
	/*
		Example:
		 http://localhost:3333/upload?timestamp=20230307T003505&product=hrit-ash&file.web.url=https://brunnur.vedur.is/pub/eysteinn/viirs-granule-true-color_20230222T183723.tiff&file.web.type=tiff
	*/
	//request.Response.StatusCode = 404
	//log.Fatal(err)

	obj := prog.ConsumerObject{}
	var err error

	timestr := request.URL.Query().Get("timestamp")
	if timestr != "" {
		obj.Timestamp, err = time.Parse("20060102T150405", timestr)
		if err != nil {
			request.Response.StatusCode = 404
			//log.Fatal(err)
			return
		}
	}
	qval := request.URL.Query()
	obj.Product = qval.Get("product")

	file_web_url := qval.Get("file.web.url")
	file_web_type := qval.Get("file.web.type")
	if file_web_url != "" && file_web_type != "" {
		web := prog.Web{}
		web.Url = file_web_url
		web.Type = file_web_type
		obj.File.Web = &web
	}

	file_local_path := qval.Get("file.local.path")
	if file_local_path != "" {
		local := prog.Local{}
		local.Path = qval.Get("file.local.path")
		obj.File.Local = &local
	}

	geo := prog.Geo{}
	obj.Geo = geo

	b, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
	/*request.URL.Query().Get("product")
	param1 := r.URL.Query().Get("param1")
	*/
	//io.WriteString(respones, string(b))
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

	prog.Psql.TryFill()

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)

	mux.Get("/upload", upload)
	/*mux := http.NewServeMux()
	mux.HandleFunc("/upload", upload)
	mux.HandleFunc("/delete", delete)*/
	fmt.Println("Starting to listen")
	err := http.ListenAndServe(":3333", mux)
	if err != nil {
		log.Panic(err)
	}
}
