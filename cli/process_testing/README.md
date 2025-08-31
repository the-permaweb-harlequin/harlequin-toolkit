# AO Process testing

This is a suite of tools to test AO processes for security and memory limitations.

This will have a typescript project and a docker container to handle testing wasm binaries.

## Scope

binaryen and other wasm tooling is used to analyze the wasm binary itself to verify that
the appropriate functions are exported.

- Should export the proper EMSCRIPTEN functions
- Should have the handle function be callable, handle the msgJSON and envJSON arguments
- Should have the handle function return properly formatted
- Should not allow authorities not in the authorities list to assign messages
- Should correctly identify assignments
- Should correctly use the msg timestamp for the languages clock/time behaviour
- Should be compatible with WeaveDrive.
