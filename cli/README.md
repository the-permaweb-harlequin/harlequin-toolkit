# CLI

The harlequin CLI uses Bubble Tea.

Plan for this cli

## Stage 1: flavored aos builds

Allow people to write lua processes with AOS as its wrapper

leverages the existing AO build container

## Stage 2: standard language builds

Allow people to leverage more wasm targeted languages - assemblyscript, c, and rust already have examples.

Provide the basic AOS framework and stdlib for each - ideally published as an installable lib

You should be able to import the AOS tooling and write a process that compiles with either a specific tool, or using the standard build tools with the right configuration.

## Stage 3: Contract templates - tokens, agents, etc with testing for each template.

Build an ecosystem of well tested contracts and add them as template projects.

Mimicking the Go install from url might be a good one here