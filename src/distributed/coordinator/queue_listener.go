package coordinator

import (
    "bytes"
    "encoding/gob"
    "fmt"
    "go-building-queueing-distributed-system/src/distributed/dto"
    "go-building-queueing-distributed-system/src/distributed/qutils"

    "github.com/streadway/amqp"
)

const url = "amqp://guest:guest@localhost:5672"

type QueueListener struct {
    conn    *amqp.Connection
    ch      *amqp.Channel
    sources map[string]<-chan amqp.Delivery
    ea      *EventAggregator
}

func NewQueueListener(ea *EventAggregator) *QueueListener {
    ql := QueueListener{
        sources: make(map[string]<-chan amqp.Delivery),
        ea:      ea,
    }
    ql.conn, ql.ch = qutils.GetChannel(url)
    return &ql
}

func (ql *QueueListener) ListenForNewSource() {
    q := qutils.GetQueue("", ql.ch, true)
    ql.ch.QueueBind(
        q.Name,
        "",
        "amq.fanout",
        false,
        nil,
    )
    msgs, _ := ql.ch.Consume(
        q.Name,
        "",
        true,
        false,
        false,
        false,
        nil,
    )
    ql.DiscoverSensors()

    for msg := range msgs {
        ql.ea.PublishEvent("DataSourceDiscovered", string(msg.Body))
        sourceChan, _ := ql.ch.Consume(
            string(msg.Body),
            "",
            true,
            false,
            false,
            false,
            nil,
        )
        if ql.sources[string(msg.Body)] == nil {
            ql.sources[string(msg.Body)] = sourceChan
            go ql.AddListener(sourceChan)
        }
    }
}
func (ql *QueueListener) AddListener(msgs <-chan amqp.Delivery) {
    for msg := range msgs {
        r := bytes.NewReader(msg.Body)
        d := gob.NewDecoder(r)
        sd := new(dto.SensorMessage)
        d.Decode(sd)
        fmt.Printf("Received Message: %v\n", sd)
        ed := EventData{
            Name:      sd.Name,
            Timestamp: sd.Timestamp,
            Value:     sd.Value,
        }
        ql.ea.PublishEvent("MessageReceived_"+msg.RoutingKey, ed)
    }
}

func (ql *QueueListener) DiscoverSensors() {
    ql.ch.ExchangeDeclare(
        qutils.SensorDiscoveryExchange,
        "fanout",
        false,
        false,
        false,
        false,
        nil,
    )
    ql.ch.Publish(
        qutils.SensorDiscoveryExchange,
        "",
        false,
        false,
        amqp.Publishing{},
    )
}
