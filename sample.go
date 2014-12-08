//
// Accept requests with parameter "password=xxxx" and process "xxxx" into SHA512
// hashed value returned in Base64.  Input value of "shutdown=true" terminates
// server.
//
// usage: sample.go delay port
//
//    where "delay" is response delay time in seconds
//          "port" is http/server listener port
//

package main

import "crypto/sha512"
import "encoding/base64"
import "fmt"
import "io"
import "net/http"
import "os"
import "strconv"
import "time"


var shuttingDown       = false  // latched, indicates when "shutdown" is seen
var pendingRequests    = 0      // number of current in-flight operations
var responseDelay      = 0      // response slew time in seconds
var listenPort string           // http port to use
var errorValue error            // generic error


// http handler...accepts requests and processes into output reply
func do_request(reply http.ResponseWriter, request *http.Request) {
    // starting a new request
    pendingRequests = pendingRequests + 1

    // grab input values
    request.ParseForm()
    inputValues := request.Form

    // was this a shutdown request?
    hasShutdown := inputValues.Get("shutdown") == "true"
    shuttingDown = shuttingDown || hasShutdown

    // honor the shutdown request
    if (hasShutdown) {
        // request completed
        pendingRequests = pendingRequests - 1

        // spin wait for all others to finish
        for pendingRequests > 0 {
            time.Sleep(1)
        }

        // terminate the server
        http.Error(reply, "Shutdown Requested", 404)
        os.Exit(0)
    } else {
        // process the request
        reply.Header().Set("Content-Type", "text/html",)
        io.WriteString(reply, "<doctype html> <html> <head> ", )
        io.WriteString(reply, "<title>Hashed Password</title> </head> <body>", )
        io.WriteString(reply, do_password_encrypt(inputValues.Get("password")), )
        io.WriteString(reply, "</body> </html>", )

        // request completed
        pendingRequests = pendingRequests - 1
    }
}


// receive plain text password, hash using SHA512, and return result in Base64
func do_password_encrypt(plainText string) string {
    var result string

    // launch delay timer so we do not respond too quickly
    timer := time.NewTimer(time.Second * time.Duration(responseDelay))

    // sanity check on input value
    result = ""

    if (plainText == "") {
        // rendezvous on response delay timer before returning results
        <-timer.C
        return result
    }

    // are we still honoring requests?
    if (shuttingDown) {
        result = "*** No longer accepting inputs...try again later ***"
    } else {
        // hash input received and encode into Base64 (URL-friendly)
        hashEngine := sha512.New()
        hashEngine.Write([]byte(plainText))
        hashedByteString := hashEngine.Sum(nil)

        result = base64.URLEncoding.EncodeToString([]byte(hashedByteString))
    }

    // rendezvous on response delay timer before returning results
    <-timer.C

    return result
}


func main() {
    // get args supplied
    cmdArgs := os.Args[1:]
    numArgs := len(cmdArgs)

    // sanity check the call
    if (numArgs > 2) {
        fmt.Println("usage: sample.go responseDelay httpPort")
        os.Exit(0)
    }

    if (numArgs >= 1 && (cmdArgs[0] == "-?" || cmdArgs[0] == "--help")) {
        fmt.Println("usage: sample.go responseDelay httpPort")
        os.Exit(0)
    }

    // arg0 = response timer delay in seconds
    responseDelay = 5

    if (numArgs >= 1 && cmdArgs[0] != "") {
        responseDelay, errorValue = strconv.Atoi(cmdArgs[0])
        if (errorValue != nil) {
            fmt.Println("usage: sample.go responseDelay httpPort")
            os.Exit(0)
        }
    }

    // arg1 = http listener port
    listenPort = "8080"

    if (numArgs == 2 && cmdArgs[1] != "") {
        listenPort = cmdArgs[1]
    }

    checkPort := 0
    checkPort, errorValue = strconv.Atoi(listenPort)

    if (errorValue != nil || checkPort < 1 || checkPort > 65535) {
        fmt.Println("usage: sample.go responseDelay httpPort")
        os.Exit(0)
    }

    listenPort = ":" + listenPort  // reformat into ListenAndServe parameter

    // invoke http listener and handle requests until told to shutdown
    http.HandleFunc("/", do_request)
    http.ListenAndServe(listenPort, nil)
}

