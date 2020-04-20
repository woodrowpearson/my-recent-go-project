
## Wed 4/15

1. sketch out project - DONE
2. solidify shelf selection behavior - WIP
3. figure out the shelf scoring thing better.
4. add in logging output to conform with test rubric
5. add unit tests

## Thurs 4/16 

1. reread shelf scoring logic and implement - DONE
2. sniff out bug in the shelf assignation logic - DONE
3. add in logging output - DONE
4. add in logic for going with second-best shelf on selection if target shelf is full, and then additional logic for discard if both shelves are full. - DONE
5. discuss next steps with woody: move to golang? or try with coroutines? - DONE
6. If coroutines, rework dispatching behavior to use a pool of coroutines.? - DONE
7. rewrite in golang - WIP
8. add in waitgroup for cleanup - DONE
9. add in atomic arrays for shelves for displaying progress - DONE
10. add in logging for all critical events per rubric - DONE

## Fri 4/17

1. make shelves an object for easier logging and less boilerplate - DONE
2. add in CLI options for modifying behaviors - DONE
3. make shelf constructor function and move counters to structs  - DONE
4. account for order rates different than 2 - DONE
5. read chapters 1-8 on go with testing - WIP
6. better logging to match rubric - DONE
7. use built in struct constructors for args - DONE
8. make the decrementAndUpdate a method on the struct - DONE

## Sat 4/18

Testing Strategy:

- Pass an io.Writer handle to functions
that perform logging. find a way to do concurrent file writing - DONE
- write a unit test that reads from the io.Writer
- look at the DI and mocking chapters
- use the race condition detector in the integration tests: go test -race
- make a suite of benchmarks at different input sizes, along with a function to generate random orders for benchmarking
- we can use a channel for the updates on the array. This would avoid a scenario wherein somehow two order IDs are duplicated
- be sure to run go vet
1. read chapters 9-16 on go with testing - DONE
2. have logging go to an io.Writer - DONE
3. add unit tests to functions
4. update heuristic on overflow
5. move file to streaming ingestion
6. pool for coroutines
7. narrative of behaviors + list of decisions made
8. handoff to woody for polishing
9. move argparse to a separate file - DONE

Test cases to write

1. decay factor computation tests
2. courier output (good) - need mocks for the sleep call - DONE
3. courier output(bad) - need mocks for the sleep call - DONE
4. concurrent array access
5. selectShelf (5 cases) - need outputs logged for tests
6. buildShelf (1 case)
7. argument parsing(1 case)
8. main loop (2 cases)
9. need to add heuristic for dispatching, along with tests for heuristic function
10. mocks/configuration for the sleep calls - DONE

modifications to make:
	- coro pool
	- streaming file ingestion
	- thread-safe logging that can be tested - DONE
	- mocks for decay functions

## Sunday 4/19

1. make JSON stream from a file - DONE
2. make JSON stream push to a channel, on a timer - DONE
3. rework main loop to ingest in blocks based on a Wait - DONE

1. add in goroutine pool - DROPPED
2. add in streaming json ingestion - DONE
3. dispatching heuristic - DONE
5. unit tests for all remaining cases incl concurrency
6. move to a package
7. have main.go import from the package as a CLI client.
8. narrative for woody.

1. dispatching heuristic update - DONE
2. move computeDecayRate to a method on order - DONE
3. add decayCriticality and decayScore to the computeDecayRate - DONE
4. extend Snapshot to compute the new score on swap.
4. WRITE TESTS

## Monday 4/20

1. finish up swapWillSave() function - DONE
2. make it into a local package - DONE
3. write tests against everything.
4. be sure to check for race conditions on the map iteration - that was a nasty error.
5. make everything but SimulatorConfig, BuildConfig, and RunSimulator private

Following tests needed:

- logging output for the courier and dispatch functions
- selectShelf: all paths - DONE
- computeDecayScore  - DONE
- swapWillPreserve (for true and false) (need Mocks) - DONE
- incrementAndUpdate (isCritical vs not isCritical)
- decrementAndUpdate (isCritical vs not isCritical)
- selectCritical - DONE
- swapAssessment - DONE
- streamFromSource -DROPPED
- BuildConfig - DROPPED

- courier logging test - DONE
- dispatch logging test - DONE
- main IO loop test - DONE

14 non-concurrency tests


- testing for race conditions (how do we do that?)
- benchmarking on integration test

### Notes

1. current implementation with celery/redis WILL NOT SCALE. the logging rubric requires listing
of all contents of shelves whenever an event occurs. With redis, this implies an additional redis
access via a pipeline. We're already at 2 separate redis accesses per event, this would put it to 3,
and each event introduces socket+parsing overhead, not to mention the btree lookup (however minimal it is
due to small dataset size) on the redis keys themselves.
2. i've avoided unit tests on this initial run as a matter of sketching things out.
3. we need to add in discard logic.

## to run (for rough draft)

first window:
celery worker -A worker.celery_app --loglevel=info
second window:
python order_queue.py

## To Run

go build main.go
./main.go 

