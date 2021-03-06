package simulator

import (
	"testing"
	"time"
)

/*
	Scenario:
		Order has a shelf life of twelve seconds.
		Courier will arrive in seven seconds.
		one second has elapsed, with a decay rate of 1
		and a modifier of four.
		This means the current elapsed score is
			(12 - (1 second)*4*1)/12 => 0.67
		but the final score would be:
			(12 - (7 seconds)*1*4)/12 => -1.33 (failing)
		in two seconds, the order will decay out.
		if it is swapped to a shelf with modifier of 1,
		the final decay score would be:
			(12 - (1 second)*4*1)/12 => 0.67 (elapsed) +
			(12 - (6 seconds)*1)/12 => 0.5 (on new shelf)
		for a total of 1.17, which will let the order survive.
*/
func TestSwapWillPreserve(t *testing.T) {
	t.Run("Swapping will prevent at-risk order from decaying out.", func(t *testing.T) {

		overflowShelf := buildOrderShelf(1, "overflow", 4)
		hotShelf := buildOrderShelf(1, "hot", 1)
		mockNow := mockTimeNow()
		oneSecondAgo := mockNow.Add(time.Second * time.Duration(-1))
		arrivalTime := oneSecondAgo.Add(time.Second * time.Duration(7))
		order := foodOrder{Id: "a", Name: "dummy", Temp: "hot", ShelfLife: 12, DecayRate: 1,
			IsCritical: true, placementTime: oneSecondAgo,
			arrivalTime: arrivalTime, shelf: overflowShelf}
		order.DecayScore = order.computeDecayScore(overflowShelf.modifier, 7*1000)
		assertFloat32(t, order.DecayScore, float32(-4)/float32(3))
		res := order.swapWillPreserve(hotShelf.modifier, mockTimeNow)
		assertBoolean(t, res, true)
		assertBoolean(t, order.IsCritical, false)
		expectedElapsedScore := order.computeDecayScore(overflowShelf.modifier, 1000)
		expectedNewShelfScore := order.computeDecayScore(hotShelf.modifier, 6000)
		expectedProspectiveScore := expectedElapsedScore + expectedNewShelfScore
		assertFloat32(t, order.DecayScore, expectedProspectiveScore)
	})
	t.Run("Swapping will NOT prevent at-risk order from decaying out.", func(t *testing.T) {
		/*
			Scenario:
				Order has a shelf life of twelve seocnds.
				Courier will arrive in EIGHT seconds.
				One second has elapsed, with a decay rate of 1
				and a modifier of four.
				This means the current elapsed score is
					(12 - (1 second)*4*1)/12 => 0.67
				but the final score would be:
					(12 - (8 seconds)*1*4)/12 => -1.67 (failing)
				The other available shelf has a modifier of 3.
				If it is swapped to the shelf with a 3 modifier,
				the final decay score would be:
					(12 - (1 second)*4*1)/12 => 0.67 (elapsed) +
					(12 - (7 seconds)*3*1)/12 => -0.75 (on new shelf)
				for a total of roughly -0.08, which would still fail
		*/
		overflowShelf := buildOrderShelf(1, "overflow", 4)
		hotShelf := buildOrderShelf(1, "hot", 3)
		mockNow := mockTimeNow()
		oneSecondAgo := mockNow.Add(time.Second * time.Duration(-1))
		arrivalTime := oneSecondAgo.Add(time.Second * time.Duration(8))
		order := foodOrder{Id: "a", Name: "dummy", Temp: "hot", ShelfLife: 12, DecayRate: 1,
			IsCritical: true, placementTime: oneSecondAgo,
			arrivalTime: arrivalTime, shelf: overflowShelf}
		order.DecayScore = order.computeDecayScore(overflowShelf.modifier, 8*1000)
		assertFloat32(t, order.DecayScore, float32(-5)/float32(3))
		res := order.swapWillPreserve(hotShelf.modifier, mockTimeNow)
		assertBoolean(t, res, false)
		assertFloat32(t, order.DecayScore, float32(-5)/float32(3))
		assertBoolean(t, order.IsCritical, true)
	})
}

/*
	Three cases:
		1. zero shelf life
		2. B greater than A
		3. A greater than B
*/
func TestComputeDecayScore(t *testing.T) {
	order := foodOrder{Id: "a", Name: "dummy", Temp: "hot",
		ShelfLife: 200, DecayRate: 0.25}

	t.Run("Returns zero when order shelf life is zero",
		func(t *testing.T) {
			order.ShelfLife = 0
			res := order.computeDecayScore(1, 1*1000)
			expected := float32(0)
			assertFloat32(t, res, expected)
		})

	msg := `
Returns a negative result when the decay rate,
modifier, and arrival time outweigh shelf life.
`
	t.Run(msg, func(t *testing.T) {
		order.ShelfLife = 10
		res := order.computeDecayScore(2, 1000*1000)
		expected := float32(-49)
		assertFloat32(t, res, expected)
	})

	msg = `
Returns a positive result when shelf life
outweighs decay factors.
`
	t.Run(msg, func(t *testing.T) {
		order.ShelfLife = 200
		res := order.computeDecayScore(1, 2*1000)
		expected := float32(0.9975)
		assertFloat32(t, res, expected)
	})
}

// Verify it selects correct shelf else full then dead.
func TestSelectShelf(t *testing.T) {
	order := foodOrder{Id: "a", Name: "dummy", Temp: "hot",
		ShelfLife: 200, DecayRate: 0.25}
	overflow := buildOrderShelf(1, "overflow",
		0)
	cold := buildOrderShelf(1, "cold", 0)
	hot := buildOrderShelf(1, "hot", 0)
	frozen := buildOrderShelf(1, "frozen", 0)
	dead := buildOrderShelf(0, "dead", 0)
	shelves := orderShelves{overflow: overflow, cold: cold,
		hot: hot, frozen: frozen, dead: dead}

	t.Run("returns dead if matchingScore and overflorScore are both less than zero",
		func(t *testing.T) {
			order.ShelfLife = 0
			res := order.selectShelf(&shelves, 100, mockTimeNow)
			expected := dead
			assertShelf(t, res, expected)
		})

	t.Run("returns dead if no space in matching and overflow shelves",
		func(t *testing.T) {
			order.ShelfLife = 200
			overflow.counter = 0
			hot.counter = 0
			res := order.selectShelf(&shelves, 100, mockTimeNow)
			expected := dead
			assertShelf(t, res, expected)

		})

	msg := `
Returns overflow if overflow space is available
and item will survive storage in overflow.
Ensures that order's shelf is set to overflow,
and that its decay score is set.
`
	t.Run(msg, func(t *testing.T) {
		overflow.counter = 1
		order.ShelfLife = 200
		overflow.modifier = 2
		res := order.selectShelf(&shelves, 2, mockTimeNow)
		expected := overflow
		expectedOverflowCounter := int32(0)
		expectedDecayScore := order.computeDecayScore(overflow.modifier, 2*1000)
		assertShelf(t, res, expected)
		assertInt32(t, res.counter, expectedOverflowCounter)
		assertFloat32(t, order.DecayScore, expectedDecayScore)
	})
	msg = `
Returns matching shelf if eligible for matching shelf
and no space is available in overflow shelf.
Ensures that order's shelf is set to matching shelf,
and that its decay score is set.
`
	t.Run(msg, func(t *testing.T) {
		overflow.counter = 0
		hot.counter = 1
		overflow.modifier = 2
		hot.modifier = 1
		res := order.selectShelf(&shelves, 2, mockTimeNow)
		expected := hot
		expectedHotCounter := int32(0)
		expectedDecayScore := order.computeDecayScore(hot.modifier, 2*1000)
		assertShelf(t, res, expected)
		assertInt32(t, res.counter, expectedHotCounter)
		assertFloat32(t, order.DecayScore, expectedDecayScore)
	})

	msg = `
Returns overflow if overflow space is available,
no matching space is available, even if item
will expire in overflow region. Ensures
that the order is set to critical, its decay score is set,
and that its shelf is set to overflow.
`
	t.Run(msg, func(t *testing.T) {
		overflow.counter = 1
		hot.counter = 0
		overflow.modifier = 1000
		hot.modifier = 0
		res := order.selectShelf(&shelves, 1000, mockTimeNow)
		expected := overflow
		expectedOverflowCounter := int32(0)
		expectedDecayScore := order.computeDecayScore(overflow.modifier, 1000*1000)
		assertShelf(t, res, expected)
		assertInt32(t, res.counter, expectedOverflowCounter)
		assertFloat32(t, order.DecayScore, expectedDecayScore)
		assertBoolean(t, order.IsCritical, true)
	})
}
