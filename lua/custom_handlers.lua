register_handler("/ping", function(req)
    return '{"pong":true,"method":"'..req.method..'"}'
end)

register_handler("/echo", function(req)
    local test = req.params["test"] or "nil"
    local foo  = req.params["foo"] or "nil"
    return '{"method":"'..req.method..'","test":"'..test..'","foo":"'..foo..'"}'
end)

register_handler("/test3", function(req)
	return '{testpass: true}'
end)
