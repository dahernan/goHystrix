package httpexp

import (
	"github.com/dahernan/goHystrix"
	"net/http"
)

func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, goHystrix.Circuits().ToJSON())
}

func init() {
	http.HandleFunc("/debug/circuits", expvarHandler)
}
