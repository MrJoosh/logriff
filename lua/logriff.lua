#!/usr/bin/env lua
-- LogRiff: Lua 5.4 / LuaJIT implementation.
-- See README.md for the full CLI specification and output formats.

local function main(args)
    -- TODO: parse flags, stream input, apply filters, aggregate metrics, and emit output.
    io.stderr:write("logriff: Lua implementation is not yet written\n")
    return 1
end

os.exit(main(arg))
