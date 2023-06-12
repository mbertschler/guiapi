import guiapi from "guiapi"

import "../css/main.css"

import TodoList from "./todolist.js"
import Reports from "./reports.js"

guiapi.registerFunctions(TodoList)
guiapi.registerFunctions(Reports)
guiapi.setupGuiapi({
    debug: true,
    errorHandler: (error) => {
        console.warn("guiapi error handler:", error)
        const el = document.getElementById("error-box")
        el.style.display = "block"
        el.innerText = error.Message
    },
})
