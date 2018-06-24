package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/utrack/clay/integration/go_package_rel_proto_in_root/strings"
)

func main() {
	r := chi.NewMux()
	desc := strings.NewStrings().GetDescription()
	desc.RegisterHTTP(r)

	r.Handle("/swagger.json", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(desc.SwaggerDef())
	}))

	http.ListenAndServe(":8080", r)
}