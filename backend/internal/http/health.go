package http

import (
	"encoding/json"
	"net/http"
)

// Health responde GET /api/health — usado pelo healthcheck do Docker Compose.
func Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
