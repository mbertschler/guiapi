import { guiapi } from "guiapi"
import "../css/reports.css"

let isRefreshing = false

function onRefresh(event) {
    if (isRefreshing) {
        return
    }

    const spinner = document.getElementById("refresh-spinner")
    refreshStart(event, spinner)
    guiapi("Reports.Refresh", null, () => {
        refreshDone(event, spinner)
    })
}

function refreshStart(event, spinner) {
    isRefreshing = true
    spinner.style.display = ""
    event.target.setAttribute("disabled", "")
}

function refreshDone(event, spinner) {
    isRefreshing = false
    spinner.style.display = "none"
    event.target.removeAttribute("disabled")
}

export default {
    Reports: {
        onRefresh
    }
}
