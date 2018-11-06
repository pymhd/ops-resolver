package main

import (
	"sync/atomic"
)

type counter int32

const (
	Total counter = iota
	Success 
	Error
)

var (
        tot, suc, er int32
        StatsKeeper = map[counter]*int32{
                                          Total: &tot,
                                          Success: &suc,
                                          Error: &er,
                                        }
)

func IncQueryCounter(c counter) {
	atomic.AddInt32(StatsKeeper[c], 1)
}

func DecQueryCounter(c counter) {
        atomic.AddInt32(StatsKeeper[c], -1)
}

func GetQueryCounter() (t, s, e int32){
        t = atomic.LoadInt32(StatsKeeper[Total])
        s = atomic.LoadInt32(StatsKeeper[Success]) 
        e = atomic.LoadInt32(StatsKeeper[Error])
        return
}
