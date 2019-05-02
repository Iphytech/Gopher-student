

package main

import (
    "fmt"
    "math/rand"
    "sync/atomic"
    "time"
)


type readOp struct {
    key  int
    resp chan int
}
type writeOp struct {
    key  int
    val  int
    resp chan bool
}

func main() {

    // As before we'll count how many operations we perform.
    var readOps uint64
    var writeOps uint64

    // The `reads` and `writes` channels will be used by
    // other goroutines to issue read and write requests,
    // respectively.
    reads := make(chan *readOp)
    writes := make(chan *writeOp)

    // Here is the goroutine that owns the `state`, which
    // is a map as in the previous example but now private
    // to the stateful goroutine. This goroutine repeatedly
    // selects on the `reads` and `writes` channels,
    // responding to requests as they arrive. A response
    // is executed by first performing the requested
    // operation and then sending a value on the response
    // channel `resp` to indicate success (and the desired
    // value in the case of `reads`).
    go func() {
        var state = make(map[int]int)
        for {
            select {
            case read := <-reads:
                read.resp <- state[read.key]
            case write := <-writes:
                state[write.key] = write.val
                write.resp <- true
            }
        }
    }()

    // This starts 100 goroutines to issue reads to the
    // state-owning goroutine via the `reads` channel.
    // Each read requires constructing a `readOp`, sending
    // it over the `reads` channel, and the receiving the
    // result over the provided `resp` channel.
    for r := 0; r < 100; r++ {
        go func() {
            for {
                read := &readOp{
                    key:  rand.Intn(5),
                    resp: make(chan int)}
                reads <- read
                <-read.resp
                atomic.AddUint64(&readOps, 1)
                time.Sleep(time.Millisecond)
            }
        }()
    }

    // We start 10 writes as well, using a similar
    // approach.
    for w := 0; w < 10; w++ {
        go func() {
            for {
                write := &writeOp{
                    key:  rand.Intn(5),
                    val:  rand.Intn(100),
                    resp: make(chan bool)}
                writes <- write
                <-write.resp
                atomic.AddUint64(&writeOps, 1)
                time.Sleep(time.Millisecond)
            }
        }()
    }

    // Let the goroutines work for a second.
    time.Sleep(time.Second)

    // Finally, capture and report the op counts.
    readOpsFinal := atomic.LoadUint64(&readOps)
    fmt.Println("readOps:", readOpsFinal)
    writeOpsFinal := atomic.LoadUint64(&writeOps)
    fmt.Println("writeOps:", writeOpsFinal)
}
