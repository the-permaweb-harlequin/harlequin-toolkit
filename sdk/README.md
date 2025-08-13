# Harlequin Toolkit SDK

## Features

There are both local and remote options.

The local client interacts with docker directly.

The remote client connects to a Harlequin gRPC server, which is an ar-io-node sidecar, to execute APIs remotely.
The server uses the _local_ version of the sdk after receiving the call 😊

This is useful for applications who wish to leverage low level APIs in the browser. (eg BetterIDEa, sh_ao, etc)


### HARLocal

Local instance, calls tooling using docker.

```typescript
import { createAoSigner, ArweaveSigner } from "@ar.io/sdk"


const harlequin = new HARLocal({
    signer: createAoSigner(new ArweaveSigner(jwk))
})
```

### HARemote

```typescript
import { createAoSigner, ArweaveSigner } from "@ar.io/sdk"
import { HARLocal } from "harlequin-permaweb"


const harlequin = new HARemote({
    url: "https://daemongate.io/harlequin"
    signer: createAoSigner(new ArweaveSigner(window.arweaveWallet))
})
```


### AOS Flavor compilers

AOS flavor compilers leverage the ao CLI docker container to build a WASM binary with your code in it by default.

```typescript
import { createAoSigner, ArweaveSigner } from "@ar.io/sdk"
import { HARemote, CompilerOptions } from "harlequin-permaweb"


const harlequin = new HARemote({
    url: "https://daemongate.io/harlequin"
    signer: createAoSigner(new ArweaveSigner(window.arweaveWallet))
}, {
    signerConfig: { tags: [{name: "App-Name", value: "Harlequin-Tools"}] } // optional, globally configurable tags to be added to all transactions
})

const wasm64AosCompiler = harlequin.compilers.createWasm64AosCompiler({
    memory: "large", // required. Shape looks like {stack_size, initial_memory, maximum_memory}
    aosGitHash: "latest", // optional, uses a pinned hash of 15dd81ee596518e2f44521e973b8ad1ce3ee9945 if not provided
    metering: true, // optional, defaults to false. This will enable gas metering on your module.
    sqlite: true, // optional, defaults to false. This will include the aos-sqlite module.
    llama: true, // optional, defaults to false, this will include the llama LLM module
    extensions: ["WeaveDrive"] // optional, enables provided extensions.
})

// Custom memory options for a controlled memory size
const wasm32AosCompiler = harlequin.compilers.createWasm32AosCompiler({
    memory: {
        stack_size: 3145728,
        initial_memory: 4194304,
        maximum_memory: 1073741824 
    }
})

// We can then take our Lua files and call the compiler to get our wasm binary.

const files = {
    "main.lua": "...lua file",
    "deps/my-module.lua": "...lua file"
}

const wasmBinary = await wasm32AosCompiler.compile({
    files,
    entryFile: "main.lua", // important to mark which file is the entry point
    onLog: (line) => console.log(line) // optional, client will write logs from the build back.
})

// Now we can publish it.

const { moduleId } = await wasm32AosCompiler.publish({
    wasmBinary,
    tags: [ {name: "App-Name", value: "Harlequin-Tools"} ], // extra tags, optional,
    signer, // optional, will use the default signer that was configured with the Harlequin instance, but this can override if desired.
})

// Alternatively we can compile and publish in a single call.

const { moduleId, wasmBinary } = await wasm32AosCompiler.build({
    files,
    entryFile: "main.lua", // important to mark which file is the entry point
    onLog: (line) => console.log(line) // optional, client will write logs from the build back.
    tags: [ {name: "App-Name", value: "Harlequin-Tools"} ], // optional, will also use the global tags.
    signer, // optional, will use the default signer that was configured with the Harlequin instance, but this can override if desired.
})

// TODO: C, RUST, ASSEMBLY SCRIPT COMPILERS
```


### Codegen

Codegen is a Harlequin APM package designed as a Handlers wrapper, and an external client. If your process implements Codegen then
you can leverage this for generating an SDK to interact with your process.

Plus, it makes handling messages and writing processes easy and fun 😊

To install, simply run the following. This can be run in the boot of the process or after the process is spawned.

```lua
apm.install("@harlequin/tools")
```

#### AOS Usage

```lua
local harlequin = require("@harlequin/tools")

-- This mounts the harlequin read handlers. It allows the external codegen tooling to read the JSON schemas for interacting
-- with your Handlers.
harlequin.mount()

-- creates a handler matching the tag name Action and the value Transfer
-- First argument is the pattern to match
-- Second argument is a lightweight JSON schema for defining your parameters to call the method.
--- These arguments are parsed and verified in the handler wrapper function.
-- Third argument is the handler function to be called.
harlequin.createHandler({
    Name = "transfer",
    Tags = { Action = "Transfer" },
    ContinueOnFinish = false, -- do you want to continue to another handler after? Defaults to true.
    EnforceHandlerPosition = 1 -- enforce this handler to be in position 1 in the handlers list
    -- By default, harlequin catches errors and returns information on the error as a message
    ---- This means that memory will be affected on the message. If you do not wish this behavior
    ---- you must disable it. Codegen will see this and handle it appropriately.
    ThrowOnError = true,
    -- MessageOutputs does a few things.
    --- 1. Messages configured in this table are sent upon completion of the handler.
    ----- If the handler failed, the error messages are sent. If it succeeds, the success messages are sent.
    --- 2. The SDK will be configured to validate that these message were sent, along with the expected parameters
    --- 3. The external Codegen client reads the configuration, and validates that the messages were sent.
    ----- The SDK will check that these messages were sent, and verify the are cranked by the MU (if sent to a process)
    MessageOutputs = {
        success = {
            -- the args and msg flag functions create keys that tell the client how to validate the variables.
            { Tags = { Action = "Debit-Notice", Quantity = harlequin.flags.args("Quantity") }, Target = harlequin.flags.msg("From") },
            { Tags = { Action = "Credit-Notice" }, Target = harlequin.flags.args("Recipient") },
        },
        error = {
            { Tags = { Action = "Transfer-Error" }, Target = harlequin.flags.msg("From") },
        }
    }
    }, {
    "From" = { type = "string", pattern = "...regex", required = true },
    "Quantity" = { type = "string", pattern = "...regex", required = true},
    "Recipient" = { type = "string", pattern = "...regex", required = true}
}, function(msg, args)
    local from = args.From,
    local qty = args.Quantity,
    local recipient = args.Recipient
    -- Your transfer logic
end)

```

#### Generating your SDK

Codegen will run your wasm binary to read your configurations from the Harlequin handlers.
The JSON schemas are pulled, parsed, and used to create APIs to connect to your process.

```typescript
const gzipped = await harlequin.codegen.generate({
    wasmBinary,
    clientName: "AwesomeProcess",
    typescript: true, // defaults to false
    generatePackageJson: false // defaults to false
})

// can write to the file system here if desired.
```

#### Using the generated SDK

##### Writes

```typescript
import { AwesomeProcess, createSigner, ArweaveSigner } from "@my-package/sdk"

const client = new AwesomeProcess({
    processId: 'the-process-id',
    signer: createSigner(new ArweaveSigner(window.arweaveWallet)),
    signerConfig: { tags: [{name: "App-Name", value: "Harlequin-Tools"}] }, // optional, globally configurable tags to be added to all transactions
    network: {
        CU_URL: "https://cu.ao-testnet.xyz",
        MU_URL: "https://mu.ao-testnet.xyz",
        SCHEDULER: "some-scheduler-id",
        HB_URL: "https://forward.computer",
        GATEWAY_URL: "https://arweave.net",
        GRAPHQL_URL: "https://arweave-search.goldsky.com/graphql"
    }
})

const { id, result } = await client.send.transfer({
    Recipient: "some-recipient",
    Quantity: "123"
}, { // this is all optional
    tags: [{ name: "App-Name", value: "My-Awesome-App" }], // must not conflict with tags generated match args
    data: "1234",
    anchor: "123456778654245",
    onSigningProgress: ({args, action, jsonSchema, processId, dataItem, step})=> {
        console.log(step)
    },
    onError: (e) => {
        console.error(e)
    },
    onMessageCrank: ({crankingChain, message}) => { console.log(crankingChain) },
    // the depth to validate message cranking to. 1 indicates output messages were scheduled to targets
    // 2 evaluates the results of those messages and ensures that they too were cranked.
    // useful for ping-pong calls
    // defaults to zero.
    validateMessageCrankingDepth: 1, 
})
```


##### Reads

Very similar to the write (send) api, the read api implements all the same apis, but instead of writing,
it does a dry run.

```typescript
import { AwesomeProcess, createSigner, ArweaveSigner } from "@my-package/sdk"

const client = new AwesomeProcess({
    processId: 'the-process-id',
})

const { id, result } = await client.read.transfer({
    Recipient: "some-recipient",
    Quantity: "123"
}, {
    tags: [{ name: "From", value: "1234" }],
    owner: '1234',
    data: "1234",
    anchor: "123456778654245",
    // might have aoloader options in the future here.
})
```

##### Spawns

Spawn a new process with the client

```typescript
import { AwesomeProcess, createSigner, ArweaveSigner } from "@my-package/sdk"

const processId = await AwesomeProcess.spawn({
        signer: createSigner(ArweaveSigner),
        tags: { tags: [{name: "App-Name", value: "Harlequin-Tools"}] },
        data: "could have your boot lua here",
        network: {
            CU_URL: "https://cu.ao-testnet.xyz",
            MU_URL: "https://mu.ao-testnet.xyz",
            SCHEDULER: "some-scheduler-id",
            HB_URL: "https://forward.computer",
            GATEWAY_URL: "https://arweave.net",
            GRAPHQL_URL: "https://arweave-search.goldsky.com/graphql"
        }
    })

```




