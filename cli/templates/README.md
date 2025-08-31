# Project templates

Harlequin project templates are designed to provide a batteries included
build system with zero dependencies or alternatively with dependency resolution.

The project can be built with local tools (emcc, cmake, etc) or the harlequin cli
can build it in its docker container. Due to the amount of build tooling required for
some projects this is the more attractive solution. The CLI can also install deps
for the user by pulling the project into its docker container, installing, then
moving everything back to the working directory.

# Languages

A variety of languages are supported

- C
- C++
- Lua
- Rust
- Assemblyscript
- Go

# Contributing

When creating a language, keep in mind the following:

- The project must use a package manager, custom or third party, doesn't matter
- The project must provide the requisite tooling to provide a secure starting
  point for AO processes - this means JSON package, and AO module for checking and
  formatting messages and evaluation results
- The project must have a build command and the cli must have the requisite build
  tooling to handle the command. The project should be able to be built by both the
  CLI and its build command (if the necessary system deps are installed)
- The project must have a test - this means once init'd, it should be able to build
  and run a single test against the wasm output.
