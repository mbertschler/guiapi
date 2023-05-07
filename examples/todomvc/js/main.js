import "../css/main.css"

import TodoList from "./todolist.js"
import guiapi from "guiapi"

guiapi.registerFunctions(TodoList)
guiapi.setupGuiapi({ debug: true })
