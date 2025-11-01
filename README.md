# LiveGo

LiveGo is a small cli program that allows you to set up a local live server with hot-reload for `.html` files.

The program loads the different files as HTML and injects a connection to them with an [EventSource](https://developer.mozilla.org/en-US/docs/Web/API/EventSource).

## Getting started

Compile the program

    make build

Run it in Linux

    bin/livego -addr 8080 -path examples/

Run it in Windows

    bin\livego.exe -addr 8080 -path examples/

Alternatively, just use

    make dev

