/*Package ecsTime provides a Facility for managing time in an ecs.System.

Time in an ecs.System is counted as the number of Process()ing rounds that have
happened (counting the current one!); so Time(0) will never be seen in
practice.

NOTE while you can, and probably should, embed a single Facility in your
toplevel ecs.System, there's nothing to stop you from having multiple levels of
time. Such a setup might make sense if there's a non-trivial relationship
between toplevel Process()ing and an inner system's Process()ing.

*/
package ecsTime
