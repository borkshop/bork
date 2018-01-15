package main

/*

# Abstract

The Wørkrüm proof sets out primarily to explore two aspects:
- level generation via an agent-based simulation of room building ( TODO link
  inspiration reference )
- a goal/task-oriented hierarchical work system; as a lofty aspiration, get all
  the way to specialized workers, architects, managers, etc.

# Narrative

A narrative sketch to inspire:
- the world starts out with a single small room within a field of impassable
  cells
- a single AI agent spawns, and sets itself a goal to go dig a
  bedroom
  - it takes a look at the map, chooses a place for the bedroom connected to
  the existing room (maybe by a hallway)
  - it then decomposes its larger goal "build a room" into "dig this cell, that
  cell, another cell, the rest of the cells..."
  - it tries to execute its first "dig a cell task"; oops, first it needs to
  move next to that cell, so another sub-ordinate task is created first
- it moves, digs the cell, repeats the process, eventually having dug the room
  out
- having dug itself a bedroom, it claims that room, and places itself a
  sleeping spot
- all of the above actions should spend energy points that must be regenerated
  by sleeping

Further notes:
- sleep should be able to be disturbed (movement in the same room while
  resting); this could provide a basis for motivating an agent to move its
  bedroom (or build it in the first place, rather than it being a static initial
  goal)
- an undisturbed night of sleep could have a chance to inspire a new goal
  (bigger bedroom, build some other sort of room, etc).
- agents could share or steal work (probably a conversational stealing model
  would be a good place to start)

# Technicals and Status

Current state is mechanic building:
- agents not even started yet; manual player only
- currently experimenting with recovering structure from agnostic world data:
  - very much want to avoid adding explicit high-level structure like "this is
  a room, here is its box/polygon"
  - rather high level structure should only exist as a notion during planning
  - any effects of structure should instead by a physical mechanic (from which
  the semantics of structure further derive)

## The World

The world is an infinite grid of (lazily instantiated!) impassable wall cells.

The world starts out with a single room, each tile has a floor and a wall.
Walls are marked solid for collision purposes. When a wall is destroyed, the
floor becomes visible, and the cell passable.

Walls are destroyed when a movable entity collides with them 4 times; TODO more
explicit damage/health system.

## Rooms and Halls

A room is at least a 2x2 area of passable cells; conversely a hallway cell is a
passable cell that is NOT part of at least a 2x2 passable region.

Currently rooms are notional only to prove and debug the room/hallway detection
logic: when the world is created, and whenever a wall is dug away, rooms
analysis is done, and every room filled with an identifying glyph. These glyph
labels are attached to the floors, and are currently the only artifact of room
detection.

## Action

Each character has a pool of action points (AP).

- moving costs 1 AP (TODO should diagonal moves cost 2?)
- digging costs 2 AP/round
  - since it takes 4 digs to destroy a wall, that means 8AP/wall
- building costs 2 AP/round
- carrying (e.g. stone) increases move cost by ceil (stone/2 AP/round) (so
  carrying 4 stone costs +2 AP/round for a total 3 AP/round)

TODO: all of the above costs are baseline; can be decreased by skill

Let's start from a budget of digging 8 walls / day. Let's grant 2 moves / wall.
So that gives us:

	(8 walls + 16 moves) / day
	= 96 AP / day = (64 + 32) AP / day
	= 64 rounds / day

Let's add another 32 AP of buffer (to account for transit to/from sleeping
spot), and say 128 AP / day budget.

If we expect to work for 64 rounds, then let's expect to sleep for half that
time, or 32 rounds:
- so we need to regenerate 4 = (128 / 32) AP / round
- if a round of sleep is disturbed, it looses a point of regen for each
  disturbance
- TODO rather than this linear regen model, maybe tweak it so theres some curve
  that rewards deep sleep and such

When a character hits 0 AP:
- digging is impossible
- movement is challenging (maybe a chance to fall asleep on the spot)

Digging a wall drops stone:
- into the cell occupied by the digging character
- 1 unit of stone / round (so 4/wall)

*/
