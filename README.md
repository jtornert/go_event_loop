# `go_event_loop`

*test repo for review :3*

In the spirit of hyperlith and the like, I want to get some feedback on this very very minimal Go code which features:

- A simple HTTP/1.1 server
- One resource with real-time updates (GET /)
- A tick-based event loop
- Handling for slow connections, open two tabs and see a simulated slow vs good connection (index/handler.go:75)
- Max connections to resource in naïve handle map implementation
- Total 3 files, 218 SLOC, datastar-go only dependency


There are a LOT of things that I'm intentionally leaving out which are not related to the event loop, such as HTTP/2, compression, and caching of pages. I just want some kind soul to look at it and give feedback on my reasoning around the event loop. Much appreciated 😊