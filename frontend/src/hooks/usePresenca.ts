import { useEffect } from "react";
import { api } from "@/lib/api";

const INTERVALO_MS = 60000; // mesmo intervalo do app antigo — grava a cada 60s

// usePresenca manda o heartbeat de presença enquanto a aba está visível (dono
// da regra "online = visto nos últimos 3 min" fica no backend, via TTL do
// Redis — ver internal/admin/presenca.go).
export function usePresenca() {
  useEffect(() => {
    function heartbeat() {
      if (document.hidden) return;
      api.post("/presenca").catch(() => {
        // presença é best-effort — uma falha aqui não afeta o resto do app
      });
    }

    heartbeat();
    const timer = setInterval(heartbeat, INTERVALO_MS);
    document.addEventListener("visibilitychange", heartbeat);

    return () => {
      clearInterval(timer);
      document.removeEventListener("visibilitychange", heartbeat);
    };
  }, []);
}
