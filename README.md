# LiveGo

LiveGo is a small cli program that allows you to set up a local live server with hot-reload of some files like `.html`, `.txt`, `.conf`, `.md`, `.toml` and `.yml`.

The program takes care of loading the different files as HTML and injecting a connection to them through an [EventSource](https://developer.mozilla.org/en-US/docs/Web/API/EventSource).