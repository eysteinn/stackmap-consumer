package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"test/fileobject"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SendRabbitMQReport(obj *fileobject.ConsumerObject) error {

	// DEBUG THIS METHOD!!!!!!!
	fmt.Println("Sending to rabbitmq")
	notification := RabbitMQNotification{}
	notification.Product = obj.Meta.Product
	notification.Timestamp = obj.GetTimeStr() //obj.Timestamp.Format("20060102T150405")
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

	routing_key := "mapviewer.consume.success." + obj.Meta.Product

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

type RabbitMQNotification struct {
	Product   string `yaml:"product"`
	Timestamp string `yaml:"timestamp"`
	Project   string `yaml:"project"`
}
