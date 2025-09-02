import binaryen from 'assemblyscript/lib/binaryen.js'
const { createType, i32, none } = binaryen

import { Transform } from 'assemblyscript/transform'

export default class EmscriptenImportResolver extends Transform {
  afterCompile(module: any): void {
    // Remove problematic imports and replace with internal implementations
    this.resolveImports(module)

    // Add minimal emscripten-compatible exports
    this.addEmscriptenExports(module)
  }

  resolveImports(module: any): void {
    // Try to remove the env.abort import that's causing issues
    try {
      module.removeImport('env', 'abort')
    } catch (e) {
      // Import might not exist, that's ok
    }

    // Add a simple abort function
    module.addFunction(
      'abort',
      none,
      none,
      [],
      module.unreachable()
    )

    // Add minimal malloc/free stubs
    module.addFunction(
      'malloc',
      createType([i32]),
      i32,
      [i32], // local variables
      // Simple malloc: just increment a global pointer
      module.block(null, [
        module.local.set(0,
          module.i32.add(
            module.global.get('__malloc_ptr', i32),
            module.local.get(0, i32)
          )
        ),
        module.global.set('__malloc_ptr', module.local.get(0, i32)),
        module.local.get(0, i32)
      ], i32)
    )

    module.addFunction(
      'free',
      createType([i32]),
      none,
      [],
      module.nop() // Simple free: do nothing
    )

    // Add a global for malloc pointer
    module.addGlobal('__malloc_ptr', i32, true, module.i32.const(65536))
  }

  addEmscriptenExports(module: any): void {
    // Export memory
    module.addMemoryExport('0', 'memory')

    // Export main function
    module.addFunction(
      'main',
      none,
      i32,
      [],
      module.i32.const(0)
    )
    module.addFunctionExport('main', 'main')

    // Export table
    module.addTableExport('0', '__indirect_function_table')

    // Export runtime functions
    module.addFunctionExport('abort', 'abort')
    module.addFunctionExport('malloc', 'malloc')
    module.addFunctionExport('free', 'free')

    // Stack management functions
    module.addFunction(
      'stackSave',
      none,
      i32,
      [],
      module.global.get('~lib/memory/__stack_pointer', i32)
    )
    module.addFunctionExport('stackSave', 'stackSave')

    module.addFunction(
      'stackRestore',
      createType([i32]),
      none,
      [],
      module.global.set('~lib/memory/__stack_pointer', module.local.get(0, i32))
    )
    module.addFunctionExport('stackRestore', 'stackRestore')

    module.addFunction(
      'stackAlloc',
      createType([i32]),
      i32,
      [i32], // local variables
      module.block(null, [
        module.local.set(1,
          module.i32.and(
            module.i32.sub(
              module.global.get('~lib/memory/__stack_pointer', i32),
              module.local.get(0, i32)
            ),
            module.i32.const(-16) // 16-byte align
          )
        ),
        module.global.set('~lib/memory/__stack_pointer', module.local.get(1, i32)),
        module.local.get(1, i32)
      ], i32)
    )
    module.addFunctionExport('stackAlloc', 'stackAlloc')

    // Other emscripten exports
    module.addFunction(
      '__wasm_call_ctors',
      none,
      none,
      [],
      module.nop()
    )
    module.addFunctionExport('__wasm_call_ctors', '__wasm_call_ctors')

    module.addFunction(
      'emscripten_stack_init',
      none,
      none,
      [],
      module.nop()
    )
    module.addFunctionExport('emscripten_stack_init', 'emscripten_stack_init')

    module.addFunction(
      'emscripten_stack_get_end',
      none,
      i32,
      [],
      module.global.get('~lib/memory/__heap_base', i32)
    )
    module.addFunctionExport('emscripten_stack_get_end', 'emscripten_stack_get_end')
  }
}
