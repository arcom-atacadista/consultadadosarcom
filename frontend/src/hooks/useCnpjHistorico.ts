import { useCallback, useEffect, useState } from "react";

export type Favorito = { cnpj: string; razao: string; tags: string[]; ts: number };
export type HistoricoItem = { cnpj: string; razao: string; ts: number };

const CHAVE_FAVORITOS = "favorites";
const CHAVE_HISTORICO = "cnpjHistory";
const HISTORICO_MAX = 50;

function lerLista<T>(chave: string): T[] {
  try {
    const raw = localStorage.getItem(chave);
    return raw ? (JSON.parse(raw) as T[]) : [];
  } catch {
    return [];
  }
}

// Favoritos e histórico ficam só no navegador (preferência local, não é
// segredo nem dado compartilhado — ver docs/migracao/05 §6).
export function useCnpjHistorico() {
  const [favoritos, setFavoritos] = useState<Favorito[]>(() => lerLista(CHAVE_FAVORITOS));
  const [historico, setHistorico] = useState<HistoricoItem[]>(() => lerLista(CHAVE_HISTORICO));

  useEffect(() => {
    localStorage.setItem(CHAVE_FAVORITOS, JSON.stringify(favoritos));
  }, [favoritos]);

  useEffect(() => {
    localStorage.setItem(CHAVE_HISTORICO, JSON.stringify(historico));
  }, [historico]);

  const isFavorito = useCallback(
    (cnpj: string) => favoritos.some((f) => f.cnpj === cnpj),
    [favoritos],
  );

  const alternarFavorito = useCallback((cnpj: string, razao: string) => {
    setFavoritos((atual) => {
      if (atual.some((f) => f.cnpj === cnpj)) {
        return atual.filter((f) => f.cnpj !== cnpj);
      }
      return [{ cnpj, razao, tags: [], ts: Date.now() }, ...atual];
    });
  }, []);

  const registrarHistorico = useCallback((itens: { cnpj: string; razao: string }[]) => {
    setHistorico((atual) => {
      const semDuplicados = atual.filter((h) => !itens.some((i) => i.cnpj === h.cnpj));
      const novo = itens.map((i) => ({ ...i, ts: Date.now() }));
      return [...novo, ...semDuplicados].slice(0, HISTORICO_MAX);
    });
  }, []);

  const limparHistorico = useCallback(() => setHistorico([]), []);

  return { favoritos, historico, isFavorito, alternarFavorito, registrarHistorico, limparHistorico };
}
