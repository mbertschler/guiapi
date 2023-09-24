import guiapi from "guiapi"
import "../css/todolist.css"

const EnterKey = "Enter"
const EscapeKey = "Escape"

function newTodoKeydown(event) {
    if (event.key != EnterKey) {
        return
    }
    guiapi.action("TodoList.NewTodo", { text: event.target.value })
}

function initEdit(element, args) {
    let stoppedEditing = false
    element.addEventListener("blur", function (event) {
        if (stoppedEditing) {
            return
        }
        guiapi.action("TodoList.UpdateItem", { id: args.id, text: event.target.value });
        return false;
    })
    element.addEventListener("keydown", function (event) {
        if (event.key == EscapeKey) {
            stoppedEditing = true
            guiapi.action("TodoList.EditItem", { id: 0 })
            return false
        }
        if (event.key != EnterKey) {
            return false
        }
        guiapi.action("TodoList.UpdateItem", { id: args.id, text: event.target.value });
    })
    element.focus()
}

export default {
    newTodoKeydown,
    initEdit,
}
