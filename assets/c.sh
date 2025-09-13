curl -s -u Electrum:Electrum -X POST http://127.0.0.1:7777      -H "Content-Type: application/json"          -d '{
           "id": "1",
           "jsonrpc": "2.0",
           "method": "paytomany",
           "params": {"outputs":[["2MzQCnSo839GFcyXNYeYGQD5wTzgN5exB96", 0.001], ["2Mydq5weSRT44Ej3ZLNykSFBzvnV8R8godU", 0.001]], "rbf": true}
         }'

