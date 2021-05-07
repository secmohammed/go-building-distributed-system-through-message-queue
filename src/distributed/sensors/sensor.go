package main

import (
    "bytes"
    "encoding/gob"
    "flag"
    "go-building-queueing-distributed-system/src/distributed/dto"
    "go-building-queueing-distributed-system/src/distributed/qutils"
    "log"
    "math/rand"
    "strconv"
    "time"

    "github.com/streadway/amqp"
)

var name = flag.String("name", "sensor", "name of sensor")
var freq = flag.Uint("freq", 5, "update frequency in cycles/sec")
var max = flag.Float64("max", 5., "maximum value for generated readings")
var min = flag.Float64("min", 1., "minimum value for generated readings")
var stepSize = flag.Float64("step", 0.1, "maximum allowable change per measurement")
var r = rand.New(rand.NewSource(time.Now().UnixNano()))
var value = r.Float64()*(*max-*min) + *min
var nom = (*max-*min)/2 + *min
var url = "amqp://guest:guest@localhost:5672"

func publishQueueName(ch *amqp.Channel) {
    msg := amqp.Publishing{
        Body: []byte(*name),
    }
    // Publish the sensor queue name, to identify which queue we are currently at.
    ch.Publish(
        "amq.fanout", // fanout methodolgy
        "",           // if changed to fanout, we don't need a queue name
        false,
        false,
        msg,
    )

}
func main() {
    flag.Parse()
    conn, ch := qutils.GetChannel(url)
    defer conn.Close()
    defer ch.Close()
    dataQueue := qutils.GetQueue(*name, ch, false)
    publishQueueName(ch)
    discoveryQueue := qutils.GetQueue("", ch, true)
    ch.QueueBind(
        discoveryQueue.Name,
        "",
        qutils.SensorDiscoveryExchange,
        false,
        nil,
    )
    go listenForDiscoverRequests(discoveryQueue.Name, ch)
    dur, _ := time.ParseDuration(strconv.Itoa(1000/int(*freq)) + "ms")
    signal := time.Tick(dur)
    // no need to register the buffer for each loop, but don't forget to reset at each loop.
    buf := new(bytes.Buffer)
    enc := gob.NewEncoder(buf)
    for range signal {
        calcValue()
        reading := dto.SensorMessage{
            Name:      *name,
            Value:     value,
            Timestamp: time.Now(),
        }
        buf.Reset()
        enc.Encode(reading)
        msg := amqp.Publishing{
            Body: buf.Bytes(),
        }
        ch.Publish(
            "", // direct methodolgy
            dataQueue.Name,
            false,
            false,
            msg,
        )
        log.Printf("reading message. value: %v\n", value)
    }
}
func listenForDiscoverRequests(name string, ch *amqp.Channel) {
    msgs, _ := ch.Consume(
        name,
        "",
        true,
        false,
        false,
        false,
        nil,
    )
    for range msgs {
        publishQueueName(ch)
    }
}
func calcValue() {
    var maxStep, minStep float64
    if value < nom {
        maxStep = *stepSize
        minStep = -1 * *stepSize * (value - *min) / (nom - *min)
    } else {
        maxStep = *stepSize * (*max - value) / (*max - nom)
        minStep = -1 * *stepSize
    }
    value += r.Float64()*(maxStep-minStep) + minStep
}
