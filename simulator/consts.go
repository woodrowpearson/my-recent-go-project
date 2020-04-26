package simulator

import "time"

// moved to this file for mocking purposes
//noinspection ALL
var getTimeNow = time.Now

const DispatchSuccessMsg = `
Dispatched order %s to courier.
Current shelf: %s.
Current shelf contents: %v.
`
const DispatchErrMsg = "Order %s discarded due to lack of capacity.\n"
const PickupSuccessMsg = `
Courier fetched item %s with remaining value of %.2f.
Current shelf: %s.
Current shelf contents: %v.
`
const PickupErrMsg = `
Discarded item with id %s due to expiration value of %.2f.
Current shelf: %s.
Current shelf contents: %v.
`

const OrderReceivedMsg = `
Received Order %s. Name: %s. Temp: %s. Shelf Life: %d. Decay Rate: %.2f.
`

const ShelfSwapMsg = `
Swapped Order %s from overflow shelf to %s shelf. Old Decay Score: %.2f. New Decay Score: %.2f.
`

const ShelfSizePrompt = "Specifies shelf capacity."

const ShelfModifierPrompt = "Specifies shelf decay modifier"

const CourierPrompt = `
Specify the timeframe bound for courier arrival.
courier_lower_bound must be less than or equal to courier_upper_bound.
courier_lower_bound and courier_upper_bound must be greater than or
equal to 0.
`
const OrderRatePrompt = `
Specify the number of orders ingested per second.
If specified as 0, Simulator will consume orders immediately upon ingestion from input source.
`

const VerbosePrompt = `
Passing this flag in will print orders to stdout upon ingestion from the input source.
`
