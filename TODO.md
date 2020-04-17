
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
9. add in atomic arrays for shelves for displaying progress
10. add in logging for all critical events per rubric

## Fri 4/17

1. update heuristic on overflow
2. unit tests on functions
3. add in pooling for coroutines
4. better logging 
5. narrative of behaviors + list of decisions made
6. add in CLI options for modifying behaviors
7. handoff to woody for polishing


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


