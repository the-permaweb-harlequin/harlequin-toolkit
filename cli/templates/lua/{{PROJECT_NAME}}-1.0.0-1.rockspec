package = "{{PROJECT_NAME}}"
version = "1.0.0-1"
source = {
   url = "git+https://github.com/{{GITHUB_USER}}/{{PROJECT_NAME}}.git",
   tag = "v1.0.0"
}
description = {
   summary = "AO Process built with Harlequin",
   detailed = [[
      An AO (Arweave Operating System) process template that provides
      a foundation for building decentralized applications on Arweave.

      Features:
      - Message handling with multiple actions
      - State management
      - JSON response formatting
      - Error handling
      - Comprehensive test suite
   ]],
   homepage = "https://github.com/{{GITHUB_USER}}/{{PROJECT_NAME}}",
   license = "MIT"
}
dependencies = {
   "lua >= 5.1, < 5.5",
   "luajson >= 1.3.4",
   "penlight >= 1.13.1",
   "luafilesystem >= 1.8.0"
}
build = {
   type = "builtin",
   modules = {
      ["{{PROJECT_NAME}}"] = "main.lua",
      ["{{PROJECT_NAME}}.handlers"] = "handlers/init.lua",
      ["{{PROJECT_NAME}}.utils"] = "utils/init.lua"
   },
   copy_directories = {
      "test"
   }
}
test = {
   type = "busted",
   flags = {
      "--verbose"
   }
}

