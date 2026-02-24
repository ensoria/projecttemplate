## Structure

### internal/cli

In Go projects, it is common to use a Makefile or Taskfile to create a project-specific task runner.

The `internal/cli` directory is intended for writing project-specific task runners directly in Go, without introducing an external task runner.

Additionally, if you want to execute batch processes via the command line instead of using the scheduler, you can implement those batch command handlers within this directory.

#### TODO

This part has not yet been implemented, but it is planned to be built on top of Cobra and integrated with `encli`.

The intended design is that when `encli [command name]` is executed and `[command name]` does not exist as a built-in command, the command implemented within this directory will be executed instead.

### internal/module

This structure is designed with Domain-Driven Design (DDD) in mind.
All logic related to a specific domain should be implemented within that domain’s directory inside this folder.

If interaction with another domain’s data is required, access should be handled by injecting the corresponding service via dependency injection (DI).
This approach is intended to make it easier to transition from a monolithic architecture to a microservices architecture.

When separating individual modules from a monolith, communication handled through service can be migrated to gRPC-based communication. This design consideration helps ensure that access from other services can be transitioned smoothly.

In addition, the `config` library is designed to allow each module directory to load its own configuration data independently. This makes it possible for each module to connect to different databases, caches, or other resources, which further facilitates the transition from a monolithic architecture to microservices.

### internal/query

This directory has a structure similar to `internal/module`, but it is specialized for retrieving data that spans multiple domains.

In the case of module, attempting to retrieve data across multiple domains requires communication between multiple modules, which can sometimes be inefficient.

If retrieving the data using a simple SQL `JOIN` is significantly more efficient, then instead of accessing data through a module endpoint, the data should be retrieved via a query endpoint.

However, delegating too much data retrieval responsibility to query can create overly tight coupling between domains, which becomes a burden when transitioning to a microservices architecture.

Therefore, processing within the query directory should be limited to cases where retrieving multiple domain datasets is clearly much simpler and more maintainable when implemented using SQL JOIN.

### plamo

The `plamo` directory contains packages that are not substantial enough to be separated into independent libraries in their own repositories, as well as tools that are more effective when freely customized within client code rather than being provided as standalone libraries.

The name "plamo" comes from the Japanese word plastic model. It conveys the idea: "We provide the base tools—please customize them freely as needed."
