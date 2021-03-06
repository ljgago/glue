# Glue - Robust Go and Javascript Socket Library

[![Join the chat at https://gitter.im/desertbit/glue](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/desertbit/glue?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Glue is a real-time bidirectional socket library. It is a **clean**, **robust** and **efficient** alternative to [socket.io](http://socket.io/). This library is designed to connect webbrowsers with a go-backend in a simple way. It automatically detects supported socket layers and chooses the most suitable one. This library handles automatic reconnections on disconnections and handles caching to bridge those disconnections.
The server implementation is **thread-safe**.


## Socket layers

Currently two socket layers are supported:

- **WebSockets** - This is the primary option. They are used if the webbrowser supports WebSockets defined by [RFC 6455](https://tools.ietf.org/html/rfc6455).
- **AjaxSockets** - This socket layer is used as a fallback mode.


## Support

Feel free to contribute to this project. Please check the [TODO](TODO.md) file for more information.


## Documentation 

Check the Documentation at [GoDoc.org](https://godoc.org/github.com/desertbit/glue).

Optional Javascript options which can be passed to Glue.

```js
var host = "https://foo.bar";

var opts = {
    // Force a socket type.
    // Values: false, "WebSocket", "AjaxSocket"
    forceSocketType: false,

    // Kill the connect attempt after the timeout.
    connectTimeout:  10000,

    // If the connection is idle, ping the server to check if the connection is stil alive.
    pingInterval:           35000,
    // Reconnect if the server did not response with a pong within the timeout.
    pingReconnectTimeout:   5000,

    // Whenever to automatically reconnect if the connection was lost.
    reconnected:        true,
    reconnectDelay:     1000,
    reconnectDelayMax:  5000,
    // To disable set to 0 (endless).
    reconnectAttempts:  10,

    // Reset the send buffer after the timeout.
    resetSendBufferTimeout: 7000 
};

// Create and connect to the server.
// Optional pass a host string and options.
var socket = glue(host, opts);
```


## Install

### Client

The client javascript Glue library is located in **[client/dist/glue.js](client/dist/glue.js)**.
It requires jQuery.

You can use bower to install the client library:

`bower install --save glue-socket`

### Server

Get the source and start hacking.

`go get github.com/desertbit/glue`

Import it with:

```go
import "github.com/desertbit/glue"
```

## Example

This socket library is very straightforward to use.
Check the sample directory for more examples.


### Client

```html
<script>
	// Create and connect to the server.
	// Optional pass a host string and options.
	var socket = glue();

    socket.onMessage(function(data) {
        console.log("onMessage: " + data);

        // Echo the message back to the server.
        socket.send("echo: " + data);
    });


    socket.on("connected", function() {
        console.log("connected");
    });

    socket.on("connecting", function() {
        console.log("connecting");
    });

    socket.on("disconnected", function() {
        console.log("disconnected");
    });

    socket.on("reconnecting", function() {
        console.log("reconnecting");
    });

    socket.on("error", function(e, msg) {
        console.log("error: " + msg);
    });

    socket.on("connect_timeout", function() {
        console.log("connect_timeout");
    });

    socket.on("timeout", function() {
        console.log("timeout");
    });

    socket.on("discard_send_buffer", function(e, buf) {
        console.log("discard_send_buffer: ");
        for (var i = 0; i < buf.length; i++) {
        	console.log("  i: " + buf[i]);
        }
    });
</script>
```


### Server

#### Read Event

Read data from the socket with a read event function.


```go
package main

import (
    "log"
    "net/http"
    "runtime"

    "github.com/desertbit/glue"
)

const (
    ListenAddress = ":8888"
)

func main() {
    // Set the glue event function.
    glue.OnNewSocket(onNewSocket)

    // Start the http server.
    err := http.ListenAndServe(ListenAddress, nil)
    if err != nil {
        log.Fatalf("ListenAndServe: %v", err)
    }
}

func onNewSocket(s *glue.Socket) {
    // Set a function which is triggered as soon as the socket is closed.
    s.OnClose(func() {
        log.Printf("socket closed with remote address: %s", s.RemoteAddr())
    })

    // Set a function which is triggered during each received message.
    s.OnRead(func(data string) {
        // Echo the received data back to the client.
        s.Write(data)
    })

    // Send a welcome string to the client.
    s.Write("Hello Client")
}

```

#### Read Write

Read data from the socket within a loop.


```go
package main

import (
    "log"
    "net/http"
    "runtime"

    "github.com/desertbit/glue"
)

const (
    ListenAddress = ":8888"
)

func main() {
    // Set the glue event function.
    glue.OnNewSocket(onNewSocket)

    // Start the http server.
    err := http.ListenAndServe(ListenAddress, nil)
    if err != nil {
        log.Fatalf("ListenAndServe: %v", err)
    }
}

func onNewSocket(s *glue.Socket) {
    // Set a function which is triggered as soon as the socket is closed.
    s.OnClose(func() {
        log.Printf("socket closed with remote address: %s", s.RemoteAddr())
    })

    // Run the read loop in a new goroutine.
    go readLoop(s)

    // Send a welcome string to the client.
    s.Write("Hello Client")
}

func readLoop(s *glue.Socket) {
    for {
        // Wait for available data.
        // Optional: pass a timeout duration to read.
        data, err := s.Read()
        if err != nil {
            // Just return and release this goroutine if the socket was closed.
            if err == glue.ErrSocketClosed {
                return
            }

            log.Printf("read error: %v", err)
            continue
        }

        // Echo the received data back to the client.
        s.Write(data)
    }
}

```

#### Only Write - Discard Read

Ignore data received from the client.

**Note:** If received data is not discarded, then the read buffer will block as soon as it is full, which will also block the keep-alive mechanism of the socket. The result would be a closed socket...

```go
package main

import (
    "log"
    "net/http"
    "runtime"

    "github.com/desertbit/glue"
)

const (
    ListenAddress = ":8888"
)

func main() {
    // Set the glue event function.
    glue.OnNewSocket(onNewSocket)

    // Start the http server.
    err := http.ListenAndServe(ListenAddress, nil)
    if err != nil {
        log.Fatalf("ListenAndServe: %v", err)
    }
}

func onNewSocket(s *glue.Socket) {
    // Set a function which is triggered as soon as the socket is closed.
    s.OnClose(func() {
        log.Printf("socket closed with remote address: %s", s.RemoteAddr())
    })

    // Discard all reads.
    // If received data is not discarded, then the read buffer will block as soon
    // as it is full, which will also block the keep-alive mechanism of the socket.
    // The result would be a closed socket...
    s.DiscardRead()

    // Send a welcome string to the client.
    s.Write("Hello Client")
}

```


## Similar Go Projects

- [go-socket.io](https://github.com/googollee/go-socket.io) - socket.io library for golang, a realtime application framework.