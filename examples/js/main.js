import guiapi from "guiapi"

import TodoList from "./todolist.js"
import "../css/main.css"

guiapi.registerFunctions(TodoList)
guiapi.setupGuiapi({ debug: true })
