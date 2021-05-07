package main

import (
    "fmt"
    "go-building-queueing-distributed-system/src/distributed/coordinator"
)

func main() {
    ql := coordinator.NewQueueListener()
    go ql.ListenForNewSource()
    var a string
    fmt.Scanln(&a)
}
