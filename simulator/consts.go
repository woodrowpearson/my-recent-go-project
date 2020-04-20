package simulator


const DispatchSuccessMsg = `
Dispatched order %s to courier.
Current shelf: %s.
Current shelf contents: %s.
`
const DispatchErrMsg = "Order %s discarded due to lack of capacity\n"
const PickupSuccessMsg = `
Courier fetched item %s with remaining value of %.2f.
Current shelf: %s.
Current shelf contents: %s.
`
const PickupErrMsg = `
Discarded item with id %s due to expiration value of %.2f.
Current shelf: %s.
Current shelf contents: %s.
`
const ShelfSizePrompt = "Specifies shelf capacity."
const ShelfModifierPrompt = "Specifies shelf decay modifier"
const CourierPrompt = `
Specify the timeframe bound for courier arrival.
courier_lower_bound must be less than or equal to courier_upper_bound.
courier_lower_bound and courier_upper_bound must be greater than or
equal to 1.
`
const OrderRatePrompt = `
Specify the number of orders ingested per second.
Must be greater than zero.
`
