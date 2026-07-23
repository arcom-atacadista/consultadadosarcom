// Package httputil traz helpers de resposta compartilhados pelos pacotes de
// recurso (auth, usuarios, ...) para não duplicar json.Marshal/erro em cada um.
package httputil

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// WriteError responde no formato { "erro": "..." } usado em toda a API.
func WriteError(w http.ResponseWriter, status int, mensagem string) {
	WriteJSON(w, status, map[string]string{"erro": mensagem})
}
