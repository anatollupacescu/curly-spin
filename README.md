# concurrent service starter

Given an application that comprises of more than one service:

- service dependencies must start in parallel
- service should start only after all their dependencies have successfully started (within timeout)
- on stop it must wait for all its dependencies to report shutdown (within timeout)

## TODO

- user provided timeouts
