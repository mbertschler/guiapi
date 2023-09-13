import { handleResponse } from "./guiapi.js"

class Stream {
    constructor(url) {
        this.url = url
        this.socket = null
        this.open = false
        this.closed = false
        this.current = null
        this.tries = 0
    }

    subscribe = (data) => {
        this.tries = 0
        if (this.open) {
            this.current = data
            this.socket.send(JSON.stringify(data))
            return
        }

        this.current = data
        if (!this.socket) {
            this.connect()
        }
    }

    connect = () => {
        console.log("websocket connecting:", this.url, "try:", this.tries)
        this.tries++
        const socket = new WebSocket(this.url, "guiapi")
        socket.onopen = this.onopen
        socket.onmessage = this.onmessage
        socket.onclose = this.onclose
        socket.onerror = this.onerror
        this.socket = socket
    }

    onopen = (event) => {
        console.log("websocket opened:", event);
        this.tries = 0
        this.open = true
        this.socket.send(JSON.stringify(this.current))
    }

    onmessage = (event) => {
        const res = JSON.parse(event.data)
        console.log("stream message:", res)
        handleResponse(res, (err) => {
            if (err) {
                console.error("websocket handleResponse error:", err)
            }
        })
    }

    onclose = (event) => {
        this.open = false
        this.socket = null
        console.log("websocket closed:", event);
        if (this.closed) {
            return
        }
        if (this.tries < 3) {
            this.connect()
        } else {
            console.log("stopped reconnecting after 3 tries")
        }
    }

    onerror = (event) => {
        // this.open = false ???
        console.log("websocket error:", event);
    }
}

const streamHandler = new Stream("ws://localhost:8000/guiapi/ws");

export function handleStream(stream) {
    console.log("guiapi handleStream:", stream)
    streamHandler.subscribe(stream)
}

export default {
    handleStream
}