package rabbitmq

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/streadway/amqp"

	"das/core/log"
)

var (
	ErrInvalidMQ  = errors.New("baseMQ was invalid")
)

var (
	seqSli    []uint64
	cacheKey  = "pmsMqCache%d_%d"

	consumerNum = 3
	publishNum  = 7
)

type exchangeCfg struct {
	name       string
	kind       string
	durable    bool
	autoDelete bool
	internal   bool
	noWait     bool
}

type queueCfg struct {
	name       string
	key        string
	exchange   string
	durable    bool
	autoDelete bool
	exclusive  bool
	noWait     bool
}

type baseMq struct {
	connection *amqp.Connection
	mqUri      string

	publishCh []*amqp.Channel
	consumeCh []*amqp.Channel

	currReConnNum int32
	mu            sync.Mutex
	reconnFlag    uint32

	ctx    context.Context
	cancel context.CancelFunc

	ctxReconn context.Context
	cancelReconn context.CancelFunc
}

func (bmq *baseMq) initConn() (err error) {
	bmq.connection, err = amqp.Dial(bmq.mqUri)
	if err != nil {
		log.Errorf("baseMq.initConn > amqp.Dial > %s", err)
		return
	}

	bmq.ctx, bmq.cancel = context.WithCancel(context.Background())
	return nil
}

func (bmq *baseMq) initChannels(publishNum, consumerNum int) (err error) {
	bmq.publishCh = make([]*amqp.Channel, publishNum)
	bmq.consumeCh = make([]*amqp.Channel, consumerNum)

	init := func(chs []*amqp.Channel, sendFlag bool) (err error) {
		for i := 0; i < len(chs); i++ {
			chs[i], err = bmq.connection.Channel()
			if err != nil {
				return
			}
			err = chs[i].Confirm(false)
			if err != nil {
				return
			}
		}
		return nil
	}

	err = init(bmq.publishCh, true)
	if err != nil {
		return
	}

	err = init(bmq.consumeCh, false)
	if err != nil {
		return
	}

	return nil
}

func (bmq *baseMq) initQueue(ch *amqp.Channel, cfg *queueCfg) error {
	queue, err := ch.QueueDeclare(
		cfg.name, // name, leave empty to generate a unique name
		true,
		false, // delete when usused
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)
	if nil != err {
		log.Errorf("baseMq.initQueue > QueueDeclare %s", err)
		return err
	}
	err = ch.QueueBind(
		queue.Name,   // name of the queue
		cfg.key,      // bindingKey
		cfg.exchange, // sourceExchange
		false,        // noWait
		nil,          // arguments
	)
	if nil != err {
		log.Errorf("baseMq.initQueue > QueueBind > %s", err)
		return err
	}

	return nil
}

//func (bmq *baseMq) initConsumer(cfg *queueCfg) (err error) {
//	queue, err := bmq.channel.QueueDeclare(
//		cfg.name, // name, leave empty to generate a unique name
//		true,
//		false, // delete when usused
//		false, // exclusive
//		false, // noWait
//		nil,   // arguments
//	)
//	if nil != err {
//		log.Errorf("baseMq.initConsumer > channelConsume > QueueDeclare %s", err)
//		bmq.channel.Close()
//		return err
//	}
//	err = bmq.channel.QueueBind(
//		queue.Name,                // name of the queue
//		bmq.channelCtx.RoutingKey, // bindingKey
//		bmq.channelCtx.Exchange,   // sourceExchange
//		false,                     // noWait
//		nil,                       // arguments
//	)
//	if nil != err {
//		log.Errorf("baseMq.initConsumer > channelConsume > QueueBind > %s", err)
//		bmq.channel.Close()
//		return err
//	}
//
//	return nil
//}

func (bmq *baseMq) initExchange(ch *amqp.Channel, cfg *exchangeCfg) error {
	err := ch.ExchangeDeclare(
		cfg.name,
		cfg.kind,
		true,
		false,
		false,
		false,
		amqp.Table{},
	)

	if err != nil {
		log.Errorf("baseMq.initExchange > %s", err)
		return err
	}

	return nil
}

func (bmq *baseMq) publishSafe(index int, exchange, routingKey string, data []byte) (err error) {
	select {
	case <-bmq.ctx.Done():
		return ErrInvalidMQ
	default:
		err = bmq.publishCh[index].Publish(
			exchange,   // publish to an exchange
			routingKey, // routing to 0 or more queues
			false,      // mandatory
			false,      // immediate
			amqp.Publishing{
				Headers:         amqp.Table{},
				ContentType:     "text/plain",
				ContentEncoding: "",
				Body:            data,
				DeliveryMode:    amqp.Persistent, // 1=non-persistent, 2=persistent
				Priority:        0,               // 0-9
				// a bunch of application/implementation-specific fields
			},
		)
		if err != nil {
			go bmq.reConn()
			return
		}
	}
	return nil
}

//func (bmq *baseMq) startACKHandle(index int) {
//	ackCh := make(chan amqp.Confirmation, 100)
//	confirmCh := bmq.publishCh[index].NotifyPublish(ackCh)
//
//	for {
//		select {
//		case v := <-confirmCh:
//			ch, ok := ackChMaps[index].Load(v.DeliveryTag)
//			if ok {
//				cch, ok := ch.(chan bool)
//				if ok {
//					cch <- v.Ack
//					close(cch)
//				}
//			}
//		case <-bmq.ctx.Done():
//			atomic.StoreUint64(&seqSli[index], 0)
//			return
//		}
//	}
//}

//func (bmq *baseMq) asyncAckHandle(index int) {
//	ackCh := make(chan amqp.Confirmation, 1000)
//	confirmCh := bmq.publishCh[index].NotifyPublish(ackCh)
//
//	for {
//		select {
//		case v := <-confirmCh:
//			if !v.Ack {
//				key := fmt.Sprintf(cacheKey, index, v.DeliveryTag)
//				val,err := db.RedisUserPool.Get(2, key)
//				if err != nil {
//					continue
//				}
//				log.Errorf("asyncAckHandle > nack [%s] > msg: %s", key, val)
//				sli := strings.Split(val, "_")
//				if len(sli) != 2 {
//					continue
//				}
//				publish(index, bmq, sli[0], util.Str2Bytes(sli[1]))
//			}
//			//else {
//			//	log.Infof("cache[%d] msg[%d] send success", index, v.DeliveryTag)
//			//}
//		case <-bmq.ctx.Done():
//			atomic.StoreUint64(&seqSli[index], 0)
//			return
//		}
//	}
//}

func (bmq *baseMq) consume(index, prefetchCount int, queueName, consumerName string) (ch <-chan amqp.Delivery, err error) {
	channel := bmq.consumeCh[index]
	err = channel.Qos(prefetchCount, 0, false)
	if err != nil {
		log.Errorf("channel.Qos > %s", err)
	}
	ch, err = channel.Consume(
		queueName,    // queue
		consumerName, // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)

	if err != nil {
		log.Errorf("baseMq.consume > Consume > %s", err)
		return nil, err
	}

	return
}

func (bmq *baseMq) reConn() error {
	log.Info("MQ Start connecting...")
	var err error
	bmq.mu.Lock()
	if !atomic.CompareAndSwapUint32(&bmq.reconnFlag, 0, 1) {
		bmq.mu.Unlock()
		<- bmq.ctxReconn.Done()
		return nil
	}
	bmq.cancel()
	bmq.ctxReconn, bmq.cancelReconn = context.WithCancel(context.Background())
	bmq.mu.Unlock()
	for {
		log.Infof("MQ %dth Reconnecting...", bmq.currReConnNum+1)
		if bmq.connection == nil || bmq.connection.IsClosed() {
			err = bmq.initConn()
			if err != nil {
				log.Errorf("baseMq.reConn > initConn > %s", err)
				bmq.currReConnNum++
				bmq.cancel()
				time.Sleep(time.Second * 10)
				continue
			}
		}
		err = bmq.initChannels(publishNum, consumerNum)
		if err != nil {
			log.Errorf("baseMq.reConn > initChannels > %s", err)
			bmq.currReConnNum++
			bmq.cancel()
			time.Sleep(time.Second * 10)
			continue
		}
		bmq.cancelReconn()
		bmq.currReConnNum = 0
		atomic.StoreUint32(&bmq.reconnFlag, 0)
		log.Info("MQ Reconnected Successfully")
		bmq.ctx, bmq.cancel = context.WithCancel(context.Background())
		return nil
	}
}

func (bmq *baseMq) Close() {
	bmq.cancel()

	for _, i := range bmq.publishCh {
		if !isNil(i) {
			i.Close()
		}
	}

	for _, i := range bmq.consumeCh {
		if !isNil(i) {
			i.Close()
		}
	}

	if !isNil(bmq.connection) {
		err := bmq.connection.Close()
		if err != nil {
			return
		}
	}

	return
}

func isNil(i interface{}) bool {
	val := reflect.ValueOf(i)
	return val.IsNil()
}