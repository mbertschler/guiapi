# guiapi
Guiapi

## TODOs

- [x] is the Layout returning as beautiful as it can be?
  - [x] nope, we should return an interface 
- [x] something for long running requests, loading indicator?
- [x] debounce as a feature *ga-block*?
- [x] websocket for server side updates
- [x] server side instant redirects
- [x] global error handler
- [x] check bundling API again
- [x] page only init stuff needs nicer API (what is put in <script> globals)
- [x] consider removing html coupling from the API
- [ ] clean up library and examples
- [ ] documentation
- [ ] tests
  - [ ] maybe use https://github.com/chromedp/chromedp


### Refactoring ideas

- [x] try reflection for component config. Nope - reflection is never clear
- [x] split Context into PageCtx and ActionCtx 
- [ ] make the Response part of ActionCtx,  functions method on the context
- [ ] turn `StreamRouter` into `map[string]StreamFunc{}`, follow name/args convention
- [ ] move as much as possible into subpackages, asset building, the JSON api objects
  - [ ] what about Response and its methods???