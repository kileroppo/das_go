package mq

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

var data = `
{
	"code": 1000,
	"status": "hello",
    "id": %d
}
`
var amqpURL = "amqp://wonly:Wl2016822@139.196.221.163:5672/"

func Product() error {
	fmt.Println("hello rabbitmq")
	start := time.Now()
	wg := new(sync.WaitGroup)

	for i := 0; i < 5000; i++ {
		wg.Add(1)
		go productOne(wg, data, i)
	}

	wg.Wait()
	end := time.Now()
	fmt.Printf("run time: %f", end.Sub(start).Seconds())

	return nil
}

func ProductByOneConn() {
	conn, _ := amqp.Dial(amqpURL)
	wg := new(sync.WaitGroup)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go productOneByOneConn(conn, wg, data, i)
	}
}

func ProductByOneChannel() error {

	conn, _ := amqp.Dial(amqpURL)
	srv2dbConf := MQConf{
		Exchange:     "Device2Db_ex2",
		ExchangeType: "fanout",
		RoutingKey:   "Device2Db_queue2",
		Reliable:     false,
		Durable:      false,
	}
	ch := NewMQChannel(srv2dbConf)
	ch.InitbyConn(conn)
	wg := new(sync.WaitGroup)

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		tmp := fmt.Sprintf(data, i)
		go productByOneCh(wg, ch, tmp)
	}
	wg.Wait()
	return nil
}

func BenchmarkProduct(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Product()
	}
}

func BenchmarkProductByOneConn(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProductByOneConn()
	}
}

func BenchmarkProductByOneCh(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProductByOneChannel()
	}
}

func TestProductByOneCh(t *testing.T) {
	if ProductByOneChannel() != nil {
		t.Error("error")
	}
}

func TestProduct(t *testing.T) {
	if Product() != nil {
		t.Error("error")
	}
}

func productOneByOneConn(conn *amqp.Connection, wg *sync.WaitGroup, data string, id int) {
	defer wg.Done()

	data = fmt.Sprintf(data, id)

	srv2dbConf := MQConf{
		Exchange:     "Device2Db_ex2",
		ExchangeType: "fanout",
		RoutingKey:   "Device2Db_queue2",
		Reliable:     false,
		Durable:      false,
	}
	mqCh := NewMQChannel(srv2dbConf)
	mqCh.InitbyConn(conn)

	if err := mqCh.Product2NormalQueue([]byte(data)); err != nil {
		log.Print(err)
	}

	mqCh.Close()
}

func productOne(wg *sync.WaitGroup, data string, id int) {
	defer wg.Done()
	data = fmt.Sprintf(data, id)

	srv2dbConf := MQConf{
		Exchange:     "Device2Db_ex2",
		ExchangeType: "fanout",
		RoutingKey:   "Device2Db_queue2",
		Reliable:     false,
		Durable:      false,
	}

	mqChannel := NewMQChannel(srv2dbConf)

	mqChannel.Init()
	if err := mqChannel.Product2NormalQueue([]byte(data)); err != nil {
		log.Print(err)
	}
	mqChannel.Close()
	CloseChannel(mqChannel)
}

func productByOneCh(wg *sync.WaitGroup, chMq *MQChannel, data string) {
	defer wg.Done()
	chMq.Product2NormalQueue([]byte(data))
}
