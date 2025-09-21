local out = wasm_call_bytes("plugin_entry", "Hello", 123,"World", "\x2A")
print(out)

wasm_call_bytes("set_key", "\x01")
local encr = wasm_call_bytes("encrypt_with_key", "data")
local dec = wasm_call_bytes("encrypt_with_key", encr)
print(encr, ":", dec)
