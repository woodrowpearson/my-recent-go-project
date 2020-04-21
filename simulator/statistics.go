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

func(stat *Statistics) update(o *Order, success bool, decayed bool){
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

func(s *Statistics) updateSwapped(){atomic.AddUint64(&s.totalSwapped,1)}

func(s *Statistics) GetHotDiscarded() uint64 {return s.hotDiscarded}
func(s *Statistics) GetColdDiscarded() uint64 {return s.coldDiscarded}
func(s *Statistics) GetFrozenDiscarded() uint64 {return s.frozenDiscarded}
func(s *Statistics) GetHotDecayed() uint64 {return s.hotDecayed}
func(s *Statistics) GetColdDecayed() uint64 {return s.coldDecayed}
func(s *Statistics) GetFrozenDecayed() uint64 {return s.frozenDecayed}
func(s *Statistics) GetHotSuccesses() uint64 {return s.hotSuccesses}
func(s *Statistics) GetColdSuccesses() uint64 {return s.coldSuccesses}
func(s *Statistics) GetFrozenSuccesses() uint64 {return s.frozenSuccesses}
func(s *Statistics) GetTotalDiscarded() uint64 {return s.totalDiscarded}
func(s *Statistics) GetTotalDecayed() uint64 {return s.totalDecayed}
func(s *Statistics) GetTotalFailures() uint64 {return s.totalFailures}
func(s *Statistics) GetTotalSuccesses() uint64 {return s.totalSuccesses}
func(s *Statistics) GetTotalProcessed() uint64 {return s.totalProcessed}
func(s *Statistics) GetTotalSwapped() uint64 {return s.totalSwapped}
