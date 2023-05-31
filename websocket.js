export function websocketDemo() {
    socket = new WebSocket("ws://localhost:8000/guiapi/ws", "guiapi")
    console.log("websocket connecting:", socket)

    socket.onopen = (event) => {
        var data = { type: "hello", data: "Hello, world!" }
        socket.send(JSON.stringify(data));
    };

    socket.onmessage = (event) => {
        console.log("websocket message:", JSON.parse(event.data));
    };
}

export default {
    websocketDemo
}