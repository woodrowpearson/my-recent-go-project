package simulator

import "sync/atomic"


type Statistics struct {
	hotDiscarded uint64
	coldDiscarded uint64
	frozenDiscarded uint64
	hotDecayed uint64
	coldDecayed uint64
	frozenDecayed uint64
	hotSuccesses uint64
	coldSuccesses uint64
	frozenSuccesses uint64
	totalDiscarded uint64
	totalDecayed uint64
	totalFailures uint64
	totalSuccesses uint64
	totalProcessed uint64
	totalSwapped uint64
}

func(stat *Statistics) update(o *foodOrder, success bool, decayed bool){
	if success {
		atomic.AddUint64(&stat.totalSuccesses,1)
		switch o.Temp {
		case "hot":
			atomic.AddUint64(&stat.hotSuccesses,1)
		case "cold":
			atomic.AddUint64(&stat.coldSuccesses,1)
		case "frozen":
			atomic.AddUint64(&stat.frozenSuccesses,1)
		}
	} else {
		atomic.AddUint64(&stat.totalFailures,1)
		switch o.Temp {
		case "hot":
			if decayed {
				atomic.AddUint64(&stat.hotDecayed,1)
				atomic.AddUint64(&stat.totalDecayed,1)
			} else {
				atomic.AddUint64(&stat.hotDiscarded,1)
				atomic.AddUint64(&stat.totalDiscarded,1)
			}
		case "cold":
			if decayed {
				atomic.AddUint64(&stat.coldDecayed,1)
				atomic.AddUint64(&stat.totalDecayed,1)
			} else {
				atomic.AddUint64(&stat.coldDiscarded,1)
				atomic.AddUint64(&stat.totalDiscarded,1)
			}
		case "frozen":
			if decayed {
				atomic.AddUint64(&stat.frozenDecayed,1)
				atomic.AddUint64(&stat.totalDecayed,1)
			} else {
				atomic.AddUint64(&stat.frozenDiscarded,1)
				atomic.AddUint64(&stat.totalDiscarded,1)
			}
		}
	}
	atomic.AddUint64(&stat.totalProcessed,1)
}

func(stat *Statistics) updateSwapped() {atomic.AddUint64(&stat.totalSwapped,1)}

func(stat *Statistics) GetHotDiscarded() uint64    {return stat.hotDiscarded}
func(stat *Statistics) GetColdDiscarded() uint64   {return stat.coldDiscarded}
func(stat *Statistics) GetFrozenDiscarded() uint64 {return stat.frozenDiscarded}
func(stat *Statistics) GetHotDecayed() uint64      {return stat.hotDecayed}
func(stat *Statistics) GetColdDecayed() uint64     {return stat.coldDecayed}
func(stat *Statistics) GetFrozenDecayed() uint64   {return stat.frozenDecayed}
func(stat *Statistics) GetHotSuccesses() uint64    {return stat.hotSuccesses}
func(stat *Statistics) GetColdSuccesses() uint64   {return stat.coldSuccesses}
func(stat *Statistics) GetFrozenSuccesses() uint64 {return stat.frozenSuccesses}
func(stat *Statistics) GetTotalDiscarded() uint64  {return stat.totalDiscarded}
func(stat *Statistics) GetTotalDecayed() uint64    {return stat.totalDecayed}
func(stat *Statistics) GetTotalFailures() uint64   {return stat.totalFailures}
func(stat *Statistics) GetTotalSuccesses() uint64  {return stat.totalSuccesses}
func(stat *Statistics) GetTotalProcessed() uint64  {return stat.totalProcessed}
func(stat *Statistics) GetTotalSwapped() uint64    {return stat.totalSwapped}
