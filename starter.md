# concurrent service starter

## states

start -> starting -> running -> halting -> completed

## high level usage

- instantiate the container
- `add` one or more services
  - error if duplicate
- declare a dependency relationship with `waitFor`, registers if not found
  - error on cyclic dep
- call _start_
  - the call will return once there are no running services
- call `shutdown`
  - the call will return once there are no running services

## implementation details

### start

- iterate services and call _start_ on each
- every service can `wait` (blocking) for its dependencies to start
- if a dependency fails the service will return error
- once all started, listen on an error channel
- if an error received, call `shutdown` on all of its dependants

## expectations

- `container` _start_ blocks while there are `starting` or `running` services
- `container` _start_ returns an empty result set if there are no `running` services
- `container` _start_ returns a result set containing errors from `services` that errored, timed out or paniced
- `container` will first start all the services an only after will start listening for the shutdown request
- any `running`/`starting` `service` can trigger a shutdown of the container by erroring
- `container` `shutdown` blocks while there are `starting` or `running` services
- `container` `shutdown` returns an empty result set if there are no `running` services
- `container` `shutdown` returns an empty result set if all `running` services have `completed` successfully
- `container` `shutdown` is called externally to trigger shutting down of all `running` services
- `container` `shutdown` is a recursive function that will traverse the tree and start shutting down from the leafs
- `container` `shutdown` will wait for a certain amount of time for the service to shutdown
- `service` can wait for its dependencies to start before it starts itself
- if a `service` dependency fails to start (errors) the error can be propagated to the `container`
- if a `service` _start_ returns a non-nil error all its dependants are `shut down` by the `container`
- `service` that returns nil is considered `completed`
- `service` that is `shuttingdown` can NOT call `container`.shutdown
- a running service will error immediately if `start` is called repetedly
- a running service will error immediately if `shutdown` is called repetedly
- a service will `panic` if `shutdown` before running

## dependency graph

- `dg` is an internal tree like data structure that holds the services and their dependencies
- `add` service to the graph
  - no op if the service is present
- `addTo` will add one `service` as a dependency to another
  - if the service is found in the `dg` root it will be moved to the new location
- the `walker` is a function that will use DF or BF strategies to walk the graph
- while walking the tree it will call the provided function on each node
- the provided function can filter, collect, start or stop the service on that node
