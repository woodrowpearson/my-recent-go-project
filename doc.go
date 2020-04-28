// Package main is an application that simulates the receiving & processing of
// food orders from a kitchen.
//
// Usage - To run the program with the default arguments simply type `./main`
//
//  Usage Options: All of the following flags can have their defaults overwritten.
//
//  Shelf Sizes:
//	 -overflow_size
//	 -hot_size
//	 -cold_size
//	 -frozen_size
//
//  Decay modifier rates:
//	 -overflowModifier
//	 -cold_modifier
//   -hot_modifier
//	 -frozen_modifier
//
//  Modify the upper and lower bound the courier will arrive.
//	 -courier_lower_bound
//	 -courier_upper_bound
//
//  Modify the orders received per second
//	 -orders_per_second
//
//  Example:
//   `./main -orders_per_second 20 -frozen_size 5`
//
// Output: The program will first write to std.out the configuration for this run.
//    ex: Configuration: &{overflowSize:15 hotSize:10 coldSize:10 frozenSize:5 courierLowerBound:2 courierUpperBound:6 ordersPerSecond:20 overflowModifier:2 coldModifier:1 hotModifier:1 frozenModifier:1
//    Once completed it will output the overall total processed, swapped, and expired. It will then output this data by order temperature category.
//
// Logs:
//    All logs to stdout will also be retained in the logs directory. If any errors occurred they will be logged accordingly to their specific file with a verbose stacktrace.
//
//
// Testing
//  - cd simulator && go test
//
//  Race Conditions
//   - cd simulator && go test -race
//
//  It is best to run the race conditions test a few times for any given test run.
//   - `for i in {1..5}; do go test -race; done`
//
package main // godoc -http ":8123"
