package dto

import (
    "encoding/gob"
    "time"
)

//SensorMessage struct structure to determine the attributes of a sensor message.
type SensorMessage struct {
    Name      string
    Value     float64
    Timestamp time.Time
}

func init() {
    // using gob encoding as it's purely written in go, if not I would rather to use JSON as it will be sufficient for other programming langauages to deal with.
    // Unless that, gob is more performant.
    gob.Register(SensorMessage{})
}
