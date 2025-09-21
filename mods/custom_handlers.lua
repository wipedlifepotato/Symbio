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

-- local addrs, err = electrum_list_addresses()
-- if not addrs then
--    print("Failed to list addresses:", err)
-- else
--    print("Addresses in wallet:")
--    for i = 1, #addrs do
--        print(i, addrs[i])
--    end
-- end

-- local newAddr, err = electrum_create_address()
-- if not newAddr then
--    print("Failed to create new address:", err)
-- else
--    print("New address created:", newAddr)
-- end

-- local balance, err = electrum_get_balance(newAddr)
-- if not balance then
--    print("Failed to get balance:", err)
-- else
--    print("Balance of new address:", balance)
-- end
-- local balance, err = electrum_get_balance("tb1q55rqf7um2636a7evrmzkqet34ww85wdjl02lg0")
-- print(balance)
-- local txid, err = electrum_pay_to("tb1q4fd0atukx96557ql07av5enl2u73ltdp06hqys", "0.00005502")
-- if not txid then
--     print("Ошибка:", err)
-- else
--    print("Отправлено, txid:", txid)
-- end

--local total, unlocked = monero_get_balance()
--print("Total XMR:", total)
--print("Unlocked XMR:", unlocked)

--local addr = monero_create_address("mylabel")
--print("New address:", addr)

-- local tx = monero_transfer("44xyz...", 0.1)
-- print("Sent TX:", tx)

-- local total, unlocked, addr = monero_get_subaddress_info(0, 1)
-- print("Subaddress index 1:", addr)
--print("Balance:", total, "XMR")
--print("Unlocked:", unlocked, "XMR")
--local total, unlocked = monero_get_subaddress_balance(0, 1) 
--print("Subaddress balance:", total, "XMR")
--print("Unlocked:", unlocked, "XMR")

register_handler("/mywallet", function(req)
    local token = req.params["Authorization"]
    if not token then
        return '{"error":"missing token"}'
    end

    local user, err = get_user_from_jwt(token)
    if not user then
        return '{"error":"invalid token: '..err..'"}'
    end

    local user_id = user.user_id
    local username = user.username

    return '{"user_id":'..user_id..', "username":" '..username..'"}'
end)

register_handler("/hello_handler", function(req)
	return '{"msg":"hello"}'
end)

--local test_user_id = 5
--local currency = "BTC"
--local bal, err = get_balance(test_user_id, currency)
--if bal then
--    print("Current balance:", bal)
--else
--    print("Failed to get balance:", err)
--end
--local ok, err = add_balance(test_user_id, currency, "0.01")
--if ok then
--    print("Balance increased by 0.01")
--    print("New balance:", get_balance(test_user_id, currency))
--else
--    print("Failed to add balance:", err)
--end
--local ok, err = sub_balance(test_user_id, currency, "0.05")
---if ok then
--    print("Balance decreased by 0.05")
--    print("New balance:", get_balance(test_user_id, currency))
--else
--    print("Failed to subtract balance:", err)
--end
--local x = plugin_hello("52525")
--print("X:",x) -- => WASM[hello.wasm]: 10
