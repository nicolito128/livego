# LiveGo

LiveGo is a small cli program that allows you to set up a local live server with hot-reload of some files like `.html`, `.txt`, `.conf`, `.md`, `.toml` and `.yml`.

The program loads the different files as HTML and injects a connection to them with an [EventSource](https://developer.mozilla.org/en-US/docs/Web/API/EventSource) (this makes hot-reloading works).
