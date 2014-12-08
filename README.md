go_sample
=========

sample of Go programming per request from TM


Accept requests with parameter "password=xxxx" and process "xxxx" into SHA512
hashed value returned in Base64.  Input value of "shutdown=true" terminates
server.

usage: sample.go delay port

   where "delay" is response delay time in seconds
         "port" is http/server listener port


