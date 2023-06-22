import { handleStream } from "./websocket.js"

export var callableFunctions = {}

export function registerFunctions(object) {
    for (const [key, value] of Object.entries(object)) {
        if (typeof value === 'function') {
            callableFunctions[key] = value
            continue
        }
        if (typeof value === 'object') {
            for (const [subkey, subvalue] of Object.entries(value)) {
                callableFunctions[key + "." + subkey] = subvalue
            }
            continue
        }
        console.warn("registerFunctions: unknown type", key, value)
    }
}

var state = null

let debugGuiapi = false

export function debugPrinting(enable) {
    debugGuiapi = enable
}

export function guiapi(name, args, callback) {
    if (debugGuiapi) {
        console.log("guiapi action:", name, "args:", args, "state:", state)
    }
    var req = {
        Name: name,
        Args: args,
        State: state,
    }
    guiapiRequest(req, callback)
}

function guiapiPage(url, callback) {
    if (debugGuiapi) {
        console.log("guiapi page:", url, "state:", state)
    }
    var req = {
        URL: url,
        State: state,
    }
    guiapiRequest(req, callback)
}

function guiapiRequest(req, callback) {
    if (!callback) {
        callback = () => { }
    }
    fetch("/guiapi", {
        method: 'POST',
        mode: 'cors',
        credentials: 'same-origin',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(req)
    }).then((response) => {
        response.json().then(r => {
            if (debugGuiapi) {
                console.log("guiapi response:", r)
            }
            handleResponse(r, callback)
        }).catch(r => console.error('response.json() error:', r))
    }).catch((reason) => {
        console.error('fetch() error:', reason)
        callback(reason)
    })
}

export function handleResponse(r, callback) {
    if (r.State) {
        state = r.State
    }
    if (r.Error) {
        console.error("[" + r.Error.Code + "]", r.Error.Message, r.Error)
        errorHandler(r.Error)
        callback(r.Error)
        return
    }
    if (r.HTML) {
        for (var j = 0; j < r.HTML.length; j++) {
            var update = r.HTML[j]
            const el = document.querySelector(update.Selector)
            if (!el) {
                console.warn("update selector not found :(", update.Selector, update)
                continue
            }

            switch (update.Operation) {
                case 1:
                    el.innerHTML = update.Content
                    break
                case 2:
                    el.outerHTML = update.Content
                    break
                case 3:
                    el.insertAdjacentHTML('beforebegin', update.Content)
                    break
                case 4:
                    el.insertAdjacentHTML('afterend', update.Content)
                    break
                default:
                    console.warn("update type not implemented :(", update)
            }
        }
    }
    if (r.JS) {
        for (var j = 0; j < r.JS.length; j++) {
            var call = r.JS[j]
            var func = callableFunctions[call.Name]
            if (func) {
                func(call.Args)
            } else {
                console.warn("function call not implemented :(", call)
            }
        }
    }
    if (r.Stream) {
        handleStream(r.Stream)
    }
    if (r.URL) {
        addPageToHistory(r.URL)
    }
    hydrate()
    callback(null)
}

function hydrate() {
    var elements = document.querySelectorAll(".ga")
    for (var el of elements) {
        if (el.dataset.action) {
            hydrateAction(el)
        }
        if (el.attributes.getNamedItem("ga-on")) {
            hydrateOn(el)
        }
        if (el.attributes.getNamedItem("ga-init")) {
            hydrateInit(el)
        }
        if (el.attributes.getNamedItem("ga-link")) {
            hydrateLink(el)
        }
    }
}

function hydrateAction(el) {
    var action = el.dataset.action
    var args = null
    // still needed?
    var variable = el.dataset.var
    if (variable) {
        args = window[variable]
    }
    var fn = callableFunctions[action]
    if (fn) {
        fn(el, args)
    } else {
        console.warn("function call not implemented :(", action, el)
        console.log("during hydrating", elements.length, "elements", elements)
    }
    el.classList.remove("ga")
}

function hydrateOn(el) {
    var eventType = el.attributes.getNamedItem("ga-on").value
    if (el.attributes.getNamedItem("ga-func")) {
        const func = el.attributes.getNamedItem("ga-func").value
        const callable = callableFunctions[func]
        if (!callable) {
            console.warn("function call not implemented :(", func, el)
        }
        el.addEventListener(eventType, callable)
    } else {
        var action = el.attributes.getNamedItem("ga-action").value
        var args = null
        if (el.attributes.getNamedItem("ga-args")) {
            args = el.attributes.getNamedItem("ga-args").value
            try {
                args = JSON.parse(args)
            } catch (e) { }
        }
        var selector = null
        if (el.attributes.getNamedItem("ga-values")) {
            selector = el.attributes.getNamedItem("ga-values").value
        }
        el.addEventListener(eventType, function (e) {
            if (selector) {
                if (args == null) {
                    args = {}
                }
                var elements = document.querySelectorAll(selector)
                for (ele of elements) {
                    args[ele.name] = ele.value
                }
            }
            guiapi(action, args)
            e.preventDefault()
            e.stopPropagation()
            return false
        })
    }
    el.classList.remove("ga")
}

function hydrateInit(el) {
    var initFunc = el.attributes.getNamedItem("ga-init").value
    var args = null
    var argsAttr = el.attributes.getNamedItem("ga-args")
    if (argsAttr) {
        args = argsAttr.value
        try {
            args = JSON.parse(args)
        } catch (e) { }
    }
    var fn = callableFunctions[initFunc]
    if (fn) {
        fn(el, args)
    } else {
        console.warn("function call not implemented :(", action, el)
        console.log("during hydrating", elements.length, "elements", elements)
    }
    el.classList.remove("ga")
}

let originalState = null

function hydrateLink(el) {
    var url = el.attributes.getNamedItem("href").value
    el.addEventListener("click", function (e) {
        guiapiPage(url, err => {
            if (err) {
                console.error("error", err)
                return
            }
            addPageToHistory(url)
        })
        e.preventDefault()
        e.stopPropagation()
        return false
    })
    el.classList.remove("ga")
}

function addPageToHistory(url) {
    console.log("addPageToHistory", url, state)
    if (originalState === null) {
        originalState = {
            url: window.location.pathname + window.location.search,
            oldState: { ...state },
        }
    }
    const pushedState = {
        url,
        oldState: { ...state },
    }
    window.history.pushState(pushedState, "", url)
}

let errorHandler = (err) => { }

export function setupGuiapi(options) {
    if (options && options.debug) {
        debugGuiapi = true
    }
    if (options && options.errorHandler) {
        errorHandler = options.errorHandler
    }
    if (options.state) {
        state = options.state
    }
    if (options.stream) {
        handleStream(options.stream)
    }
    hydrate()
    setupHistory()
}

function setupHistory() {
    window.addEventListener("popstate", function (e) {
        let s = e.state
        if (!s) {
            s = originalState
        }
        guiapiPage(s.url)
    })
}

export default {
    guiapi,
    setupGuiapi,
    registerFunctions,
    debugPrinting
}