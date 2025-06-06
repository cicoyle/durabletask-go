# Contributing

## Cloning this repository

This repository contains submodules. Be sure to clone it with the option to include submodules. Otherwise you will not be able to generate the protobuf code.

```bash
git clone --recurse-submodules https://github.com/dapr/durabletask-go 
```

If you already cloned the repository without `--recurse-submodules`, you can initialize and update the submodules with:

```bash
git submodule update --init --recursive
```

This will initialize and update all submodules, including any nested submodules if they exist.

## Building the project

This project requires go v1.19.x or greater. You can build a standalone executable by simply running `go build` at the project root.

### Generating protobuf

Use the following command to regenerate the protobuf from the submodule. Use this whenever updating the submodule reference.

```bash
# NOTE: assumes the .proto file defines: option go_package = "/api/protos"
# NOTE: currently the .proto file actually defines: option go_package = "/internal/protos"; , we are manually changing that to be /api/protos
protoc --go_out=. --go-grpc_out=. -I submodules/durabletask-protobuf/protos orchestrator_service.proto
```

For local development with protobuf changes:

1. If you have local changes to the proto files in a neighboring durabletask-protobuf directory:
   ```bash
   # Point go.mod to your local durabletask-protobuf repo
   replace github.com/dapr/durabletask-protobuf => ../durabletask-protobuf
   
   # Regenerate protobuf files using your local proto definitions
   protoc --go_out=. --go-grpc_out=. -I ../durabletask-protobuf/protos orchestrator_service.proto
   ```

   This will use your local proto files instead of the ones in the submodule, which is useful when testing protobuf changes before submitting them upstream.

### Generating mocks for testing

Test mocks were generated using [mockery](https://github.com/vektra/mockery). Use the following command at the project root to regenerate the mocks.

```bash
mockery --dir ./backend --name="^Backend|^Executor|^TaskWorker" --output ./tests/mocks --with-expecter
```

## Running tests

All automated tests are under `./tests`. A separate test package hierarchy was chosen intentionally to prioritize [black box testing](https://en.wikipedia.org/wiki/Black-box_testing). This strategy also makes it easier to catch accidental breaking API changes.

Run tests with the following command.

```bash
go test ./tests/... -coverpkg ./api,./task,./client,./backend/...,./api/helpers
```

## Running integration tests

You can run pre-built container images to run full integration tests against the durable task host over gRPC.

### .NET Durable Task client SDK tests

Use the following docker command to run tests against a running worker.

```bash
docker run -e GRPC_HOST="host.docker.internal" cgillum/durabletask-dotnet-tester:0.5.0-beta
```

Note that the test assumes the gRPC server can be reached over `localhost` on port `4001` on the host machine. These values can be overridden with the following environment variables:

* `GRPC_HOST`: Use this to change from the default `127.0.0.1` to some other value, for example `host.docker.internal`.
* `GRPC_PORT`: Set this environment variable to change the default port from `4001` to something else.

If successful, you should see output that looks like the following:

```
Test run for /root/out/bin/Debug/Microsoft.DurableTask.Tests/net6.0/Microsoft.DurableTask.Tests.dll (.NETCoreApp,Version=v6.0)
Microsoft (R) Test Execution Command Line Tool Version 17.3.1 (x64)
Copyright (c) Microsoft Corporation.  All rights reserved.

Starting test execution, please wait...
A total of 1 test files matched the specified pattern.
[xUnit.net 00:00:00.00] xUnit.net VSTest Adapter v2.4.3+1b45f5407b (64-bit .NET 6.0.10)
[xUnit.net 00:00:00.82]   Discovering: Microsoft.DurableTask.Tests
[xUnit.net 00:00:00.90]   Discovered:  Microsoft.DurableTask.Tests
[xUnit.net 00:00:00.90]   Starting:    Microsoft.DurableTask.Tests
  Passed Microsoft.DurableTask.Tests.OrchestrationPatterns.ExternalEvents(eventCount: 100) [6 s]
  Passed Microsoft.DurableTask.Tests.OrchestrationPatterns.ExternalEvents(eventCount: 1) [309 ms]
  Passed Microsoft.DurableTask.Tests.OrchestrationPatterns.LongTimer [8 s]
  Passed Microsoft.DurableTask.Tests.OrchestrationPatterns.SubOrchestration [1 s]
  ...
  Passed Microsoft.DurableTask.Tests.OrchestrationPatterns.ActivityFanOut [914 ms]
[xUnit.net 00:01:01.04]   Finished:    Microsoft.DurableTask.Tests
  Passed Microsoft.DurableTask.Tests.OrchestrationPatterns.SingleActivity_Async [365 ms]

Test Run Successful.
Total tests: 33
     Passed: 33
 Total time: 1.0290 Minutes
```

## Running locally

You can run the engine locally by pressing `F5` in [Visual Studio Code](https://code.visualstudio.com/) (the recommended editor). You can also simply run `go run main.go` to start a local Durable Task gRPC server that listens on port 4001.

```bash
go run main.go --port 4001 --db ./test.sqlite3
```

The following is the expected output:

```
2022/09/14 17:26:50 backend started: sqlite::./test.sqlite3
2022/09/14 17:26:50 server listening at 127.0.0.1:4001
2022/09/14 17:26:50 orchestration-processor: waiting for new work items...
2022/09/14 17:26:50 activity-processor: waiting for new work items...
```

At this point you can use one of the [language SDKs](#language-sdks) mentioned earlier in a separate process to implement and execute durable orchestrations. Those SDKs will connect to port `4001` by default to interact with the Durable Task engine.

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
