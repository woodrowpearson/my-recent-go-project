package simulator

import "sync/atomic"

type Statistics struct {
	hotDiscarded    uint64
	coldDiscarded   uint64
	frozenDiscarded uint64
	hotDecayed      uint64
	coldDecayed     uint64
	frozenDecayed   uint64
	hotSuccesses    uint64
	coldSuccesses   uint64
	frozenSuccesses uint64
	totalDiscarded  uint64
	totalDecayed    uint64
	totalFailures   uint64
	totalSuccesses  uint64
	totalProcessed  uint64
	totalSwapped    uint64
}

func (s *Statistics) update(o *foodOrder, success bool, decayed bool) {
	if success {
		atomic.AddUint64(&s.totalSuccesses, 1)
		switch o.Temp {
		case "hot":
			atomic.AddUint64(&s.hotSuccesses, 1)
		case "cold":
			atomic.AddUint64(&s.coldSuccesses, 1)
		case "frozen":
			atomic.AddUint64(&s.frozenSuccesses, 1)
		}
	} else {
		atomic.AddUint64(&s.totalFailures, 1)
		switch o.Temp {
		case "hot":
			if decayed {
				atomic.AddUint64(&s.hotDecayed, 1)
				atomic.AddUint64(&s.totalDecayed, 1)
			} else {
				atomic.AddUint64(&s.hotDiscarded, 1)
				atomic.AddUint64(&s.totalDiscarded, 1)
			}
		case "cold":
			if decayed {
				atomic.AddUint64(&s.coldDecayed, 1)
				atomic.AddUint64(&s.totalDecayed, 1)
			} else {
				atomic.AddUint64(&s.coldDiscarded, 1)
				atomic.AddUint64(&s.totalDiscarded, 1)
			}
		case "frozen":
			if decayed {
				atomic.AddUint64(&s.frozenDecayed, 1)
				atomic.AddUint64(&s.totalDecayed, 1)
			} else {
				atomic.AddUint64(&s.frozenDiscarded, 1)
				atomic.AddUint64(&s.totalDiscarded, 1)
			}
		}
	}
	atomic.AddUint64(&s.totalProcessed, 1)
}

func (s *Statistics) updateSwapped() { atomic.AddUint64(&s.totalSwapped, 1) }

func (s *Statistics) GetHotDiscarded() uint64 {
	hotDiscarded := atomic.LoadUint64(&s.hotDiscarded)
	return hotDiscarded
}
func (s *Statistics) GetColdDiscarded() uint64 {
	coldDiscarded := atomic.LoadUint64(&s.coldDiscarded)
	return coldDiscarded
}
func (s *Statistics) GetFrozenDiscarded() uint64 {
	frozenDecayed := atomic.LoadUint64(&s.frozenDecayed)
	return frozenDecayed
}
func (s *Statistics) GetHotDecayed() uint64 {
	hotDecayed := atomic.LoadUint64(&s.hotDecayed)
	return hotDecayed
}
func (s *Statistics) GetColdDecayed() uint64 {
	coldDecayed := atomic.LoadUint64(&s.coldDecayed)
	return coldDecayed
}
func (s *Statistics) GetFrozenDecayed() uint64 {
	frozenDecayed := atomic.LoadUint64(&s.frozenDecayed)
	return frozenDecayed
}
func (s *Statistics) GetHotSuccesses() uint64 {
	hotSuccesses := atomic.LoadUint64(&s.hotSuccesses)
	return hotSuccesses
}
func (s *Statistics) GetColdSuccesses() uint64 {
	coldSuccesses := atomic.LoadUint64(&s.coldSuccesses)
	return coldSuccesses
}
func (s *Statistics) GetFrozenSuccesses() uint64 {
	frozenSuccesses := atomic.LoadUint64(&s.frozenSuccesses)
	return frozenSuccesses
}
func (s *Statistics) GetTotalDiscarded() uint64 {
	totalDiscarded := atomic.LoadUint64(&s.totalDiscarded)
	return totalDiscarded
}
func (s *Statistics) GetTotalDecayed() uint64 {
	totalDecayed := atomic.LoadUint64(&s.totalDecayed)
	return totalDecayed
}
func (s *Statistics) GetTotalFailures() uint64 {
	totalFailures := atomic.LoadUint64(&s.totalFailures)
	return totalFailures
}
func (s *Statistics) GetTotalSuccesses() uint64 {
	totalSuccesses := atomic.LoadUint64(&s.totalSuccesses)
	return totalSuccesses
}
func (s *Statistics) GetTotalProcessed() uint64 {
	totalProcessed := atomic.LoadUint64(&s.totalProcessed)
	return totalProcessed
}
func (s *Statistics) GetTotalSwapped() uint64 {
	totalSwapped := atomic.LoadUint64(&s.totalSwapped)
	return totalSwapped
}
