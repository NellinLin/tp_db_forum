package serve

import (
	"encoding/json"
	"net/http"
)

type clientError struct {
	Message string `json:"message"`
}

// ServeJSONError - serve client error json
/*func ServeJSONError(ctx *fasthttp.RequestCtx, code int,  cliError string) {
	cliErr := clientError{
		Message: cliError,
	}
	errorJSON, _ := json.Marshal(&cliErr)

	ctx.SetStatusCode(code)
	ctx.SetContentType("application/json")
	ctx.SetBody(errorJSON)
}*/

// ServeJSON - serve data in json format
func ServeJSON(w http.ResponseWriter, code int, data interface{}) {
	dataJSON, _ := json.Marshal(&data)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dataJSON)
}
