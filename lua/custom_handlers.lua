register_handler("/ping", function(req)
    return '{"pong":true,"method":"'..req.method..'"}'
end)

register_handler("/echo", function(req)
    local test = req.params["test"] or "nil"
    local foo  = req.params["foo"] or "nil"
    return '{"method":"'..req.method..'","test":"'..test..'","foo":"'..foo..'"}'
end)

register_handler("/test3", function(req)
    local port = config.Port or "undefined"
    return string.format('{"testpass": true, "server_port": "%s"}', port)
end)

-- local mnemonic = "ice kite panda monkey apple cat fish ice monkey zebra zebra panda"
-- print("Mnemonic:", mnemonic)

-- local user = get_user("testuser")
-- if user ~= nil then
--    print("User ID:", user.id)
-- else
--    print("User not found")
-- end

-- local restored = restore_user("testuser", mnemonic)
-- if restored ~= nil then
--    print("Restored ID:", restored.id)
-- else
--    print("User not found 2")
-- end
-- if restored ~= nil then
--	change_password("testuser","12345678")
-- end

-- local token = generate_jwt(restored.id, restored.username)
-- print("JWT:", token)

