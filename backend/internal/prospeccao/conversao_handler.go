package prospeccao

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/httputil"
)

const diasPadraoConversao = 30

type ConversaoHandler struct {
	service *ConversaoService
}

func NewConversaoHandler(service *ConversaoService) *ConversaoHandler {
	return &ConversaoHandler{service: service}
}

func (h *ConversaoHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.relatorio)
	r.Post("/verificar", h.verificar)
	r.Get("/exportar", h.exportarCSV)
	return r
}

func diasDaQuery(r *http.Request) int {
	dias := diasPadraoConversao
	if v := r.URL.Query().Get("dias"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			dias = n
		}
	}
	return dias
}

func (h *ConversaoHandler) relatorio(w http.ResponseWriter, r *http.Request) {
	rel, err := h.service.Relatorio(r.Context(), diasDaQuery(r))
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao montar o relatório de conversão")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, rel)
}

func (h *ConversaoHandler) verificar(w http.ResponseWriter, r *http.Request) {
	rel, err := h.service.Verificar(r.Context(), diasDaQuery(r))
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "falha ao reconsultar a API Arcom: "+err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, rel)
}

// exportarCSV gera a tabela "Listas" em CSV — sem lib de planilha (fora do
// allowlist), com encoding/csv da stdlib (docs/migracao/08 §3).
func (h *ConversaoHandler) exportarCSV(w http.ResponseWriter, r *http.Request) {
	rel, err := h.service.Relatorio(r.Context(), diasDaQuery(r))
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "falha ao montar o relatório de conversão")
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="conversao_prospeccao.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"Lista", "Prospectou", "Assessor", "Cidade", "Quando", "Convertidos", "Total", "Conversão (%)"})
	for _, l := range rel.Listas {
		conv, pct := "—", "—"
		if l.Convertidos != nil {
			conv = strconv.Itoa(*l.Convertidos)
			if l.Total > 0 {
				pct = fmt.Sprintf("%.0f%%", float64(*l.Convertidos)/float64(l.Total)*100)
			}
		}
		_ = cw.Write([]string{l.Nome, l.NomeUsuario, l.Assessor, l.Cidade, l.CriadoEm, conv, strconv.Itoa(l.Total), pct})
	}
	cw.Flush()
}
