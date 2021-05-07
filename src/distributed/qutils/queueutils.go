package qutils

import (
    "fmt"
    "log"

    "github.com/streadway/amqp"
)

const SensorListQueue = "SensorList"
const SensorDiscoveryExchange = "SensorDiscover"
const PersistReadingsQueue = "PersistReading"

//GetChannel is used to fetch the channel and connection through dialing the rabbitmq server and check on connectivity then return them.
func GetChannel(url string) (*amqp.Connection, *amqp.Channel) {
    conn, err := amqp.Dial(url)
    failOnError(err, "Failed to establish connection  to message broker.")
    ch, err := conn.Channel()
    failOnError(err, "Failed to open a channel")
    return conn, ch
}

//GetQueue is used to fetch a queue pointer by name and a connection.
func GetQueue(name string, ch *amqp.Channel, autoDelete bool) *amqp.Queue {
    q, err := ch.QueueDeclare(
        name,
        false,
        autoDelete,
        false,
        false,
        nil,
    )
    failOnError(err, "Failed to declare queue")
    return &q
}
func failOnError(err error, msg string) {
    if err != nil {
        log.Fatalf("%s: %s", msg, err)
        panic(fmt.Sprintf("%s: %s", msg, err))
    }
}
