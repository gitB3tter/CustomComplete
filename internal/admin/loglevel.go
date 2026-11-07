package admin

import (
	"customcomplete/internal/logging"
	"fmt"
	"net/http"
	"strconv"
)

func changeLogLevel(w http.ResponseWriter, r *http.Request) {
	loglevel, err := strconv.ParseInt(r.FormValue("level"), 10, 64)
	if err != nil {
		http.Error(w, "level missing or invalid", http.StatusBadRequest)
		return
	}
	logging.SetLevel(loglevel)
	_, _ = fmt.Fprintf(w, "loglevel set to %d", loglevel)
}

func resetLogLevel(w http.ResponseWriter, r *http.Request) {
	logging.ResetLevel()
	fmt.Fprint(w, "loglevel reset")
}
