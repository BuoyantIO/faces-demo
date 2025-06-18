//go:generate go run go.bytecodealliance.org/cmd/wit-bindgen-go generate --world hello --out gen ./wit
package main

import (
	"net/http"

	"go.wasmcloud.dev/component/log/wasilog"
	"go.wasmcloud.dev/component/net/wasihttp"

	"github.com/BuoyantIO/faces-demo/v2/pkg/faces"
	"github.com/BuoyantIO/faces-demo/v2/pkg/utils"
)

struct WASIHTTPProvider struct {
	provider *BaseProvider
}

func NewWASIHTTPServer(provider *BaseProvider) *WASIHTTPProvider {
	wsrv := &WASIHTTPProvider{
		provider: provider,
	}

	return wsrv
}

func (wsrv *WASIHTTPProvider) HandleRequest(w http.ResponseWriter, r *http.Request) {
	prv := wsrv.provider
	start := time.Now()

	// Our request URL should start with /center/ or /edge/, and we want to
	// propagate that to our smiley and color services.
	subrequest := strings.Split(r.URL.Path, "/")[1]

	userAgent := r.Header.Get("user-agent")

	if userAgent == "" {
		userAgent = "unknown"
	}

	// Parse the query
	query := r.URL.Query()
	query_row := query.Get("row")
	query_col := query.Get("col")

	row := -1
	col := -1

	if query_row != "" {
		r, err := strconv.Atoi(query_row)

		if err == nil {
			row = r
		} else {
			prv.Warnf("couldn't parse row '%s', using -1: %s\n", query_row, err)
		}
	}

	if query_col != "" {
		c, err := strconv.Atoi(query_col)

		if err == nil {
			col = c
		} else {
			prv.Warnf("couldn't parse column '%s', using -1: %s\n", query_col, err)
		}
	}

	user := r.Header.Get(prv.userHeaderName)

	if user == "" {
		user = "unknown"
	}

	prvReq := &ProviderRequest{
		subrequest: subrequest,
		user:       user,
		userAgent:  userAgent,
		row:        row,
		col:        col,
	}

	resp := prv.HandleRequest(start, prvReq)

	wsrv.StandardResponse(w, r, resp)
}

func init() {
	sprv := &SmileyProvider{
		BaseProvider: BaseProvider{
			Name: "Smiley",
		},
	}

	logger := wasilog.ContextLogger("SmileyProvider")

	sprv.SetLogger(logger)

	sprv.SetGetHandler(sprv.Get)

	// sprv.BaseProvider.SetupFromEnvironment()
	sprv.SetupFromConstants(20, 0, 0.0, []int{0})
	sprv.SetSmiley(StringFromEnv("SMILEY", "Grinning"))

	// Get a new WASIHTTPServer to handle requests.
	wsrv := NewWASIHTTPServer(&sprv.BaseProvider)
	wasihttp.HandleFunc(wsrv.HandleRequest)
}

// Since we don't run this program like a CLI, the `main` function is empty. Instead,
// we call the `handleRequest` function when an HTTP request is received.
func main() {}
