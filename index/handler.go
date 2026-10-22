package index

import (
	"fmt"
	"go_event_loop/runtime"
	"net/http"
	"time"

	"github.com/starfederation/datastar-go/datastar"
)

var EventLoopId runtime.EventLoopId
var connections runtime.ConnectionMap

const html = `<!doctype html><html>
<head>
	<meta name="charset" content="UTF-8">
	<meta name="viewport" content="width=device-width">
	<style>
		html {
			color-scheme: dark;
			font-family: sans-serif;
		}
	</style>
	<title>Hello</title>
	<script type="module" src="https://cdn.jsdelivr.net/gh/starfederation/datastar@v1.0.4/bundles/datastar.js"></script>
</head>
<body
	data-indicator:fetching
	data-init="@get(location.pathname, {retryMaxCount: 3})"
>
	<h1>Hello</h1>
	<p id="message"></p>
	<p
		data-signals="{online: true}"
		data-on:online__window="$online = true"
		data-on:offline__window="$online = false"
		data-text="!$online ? 'You are offline. Please reconnect to the Internet.'
				   : !$fetching ? 'We cannot reach the server at the moment. Try refreshing the page.'
				   : $alert"
	></p>
	<p id="alert"></p>
</body>
</html>`

func Get(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("datastar-request") == "true" {
		sse := datastar.NewSSE(w, r)
		handle, ok := connections.Add(runtime.Connection{
			Sse:     sse,
			Request: r,
		})
		if !ok {
			sse.PatchElements(`<p id="message">woot</p>`)
			return
		}
		<-r.Context().Done()
		connections.Remove(handle)
		return
	}
	w.Header().Set("content-type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func EventHandler(t uint64) {
	connections.Mutex.Lock()
	defer connections.Mutex.Unlock()
	for i := 1; i < int(connections.MaxLength); i += 1 {
		conn := &connections.Connections[i]
		if !connections.Unused[i] && !conn.Busy {
			conn.Busy = true
			go func() {
				h, m, s := time.Now().UTC().Clock()
				err := conn.Sse.PatchElementf(`<p id="message">%02d:%02d:%02d</p>`, h, m, s)
				if i == 2 {
					time.Sleep(2 * time.Second)
					if conn.LastUpdateTick > 0 && t-conn.LastUpdateTick >= runtime.TicksPerSecond {
						fmt.Println("slow connection", i, t-conn.LastUpdateTick)
						err = conn.Sse.PatchSignals([]byte(`{"alert":"Slow connection! Try reconnecting later."}`))
						conn.Request.Body.Close()
						connections.Remove(conn.Handle)
					}
				}
				connections.Mutex.Lock()
				defer connections.Mutex.Unlock()
				if err != nil {
					connections.Remove(conn.Handle)
					return
				}
				conn.LastUpdateTick = t
				conn.Busy = false
			}()
		}
	}
}

func Handler() http.Handler {
	if connections.MaxLength == 0 {
		connections = runtime.NewConnectionMap(32)
		runtime.AddEventHandler(&EventLoopId, EventHandler)
	}

	return http.HandlerFunc(Get)
}
