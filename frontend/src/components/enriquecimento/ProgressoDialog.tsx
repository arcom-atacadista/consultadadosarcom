import { useEffect, useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { api } from "@/lib/api";
import { formatarCNPJ } from "@/lib/cnpj";

const ETAPAS_LABEL: Record<string, string> = {
  busca: "Busca no Maps",
  desambiguacao: "Desambiguação",
  foto_card: "Foto do card",
  street_view: "Street View",
  visao_geral: "Visão geral",
  sobre: "Sobre",
  avaliacoes: "Avaliações",
  fotos: "Fotos",
  fontes_externas: "Fontes externas",
  analise_deterministica: "Análise (regras)",
  analise_ia: "Análise (IA)",
  pdf: "PDF do dossiê",
};

type Evento = { etapa: string; status: string };
type Progresso = { status: string; etapaAtual: string; cnpj: string; eventos: Evento[] };

function badgeEtapa(status: string) {
  const s = status.toLowerCase();
  if (["ok", "concluido", "sucesso"].includes(s)) return <Badge variant="accent">Concluído</Badge>;
  if (["processando", "inicio", "start"].includes(s)) return <Badge variant="outline">Processando…</Badge>;
  if (["erro", "falha", "error"].includes(s)) return <Badge variant="danger">Erro</Badge>;
  if (["pulado", "skip", "skipped", "ignorado"].includes(s)) return <Badge variant="default">Pulado</Badge>;
  return <Badge variant="default">Aguardando</Badge>;
}

export function ProgressoDialog({ clienteId, cnpj, onClose }: { clienteId: string | null; cnpj: string; onClose: () => void }) {
  const [progresso, setProgresso] = useState<Progresso | null>(null);

  useEffect(() => {
    if (!clienteId) {
      setProgresso(null);
      return;
    }
    let timer: ReturnType<typeof setTimeout>;
    const terminal = new Set(["concluido", "nao_enriquecivel", "erro", "cancelado"]);

    async function carregar() {
      try {
        const { data } = await api.get(`/enriquecimento/${clienteId}/progresso`);
        setProgresso({ status: data.status, etapaAtual: data.etapa_atual, cnpj: data.cnpj, eventos: data.eventos ?? [] });
        if (!terminal.has((data.status ?? "").toLowerCase())) {
          timer = setTimeout(carregar, 6000);
        }
      } catch {
        // segue mostrando o último estado conhecido
      }
    }
    carregar();
    return () => clearTimeout(timer);
  }, [clienteId]);

  return (
    <Dialog open={!!clienteId} onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Fluxo de enriquecimento</DialogTitle>
          <DialogDescription>{formatarCNPJ(cnpj)}</DialogDescription>
        </DialogHeader>
        {!progresso ? (
          <p className="text-sm text-arcom-gray">Carregando o fluxo…</p>
        ) : (
          <div className="flex flex-col gap-2">
            {progresso.eventos.map((ev) => (
              <div
                key={ev.etapa}
                className="flex items-center justify-between border border-surface-border rounded-md px-3 py-2"
              >
                <span className="text-sm text-verde-escuro">{ETAPAS_LABEL[ev.etapa] ?? ev.etapa}</span>
                {badgeEtapa(ev.status)}
              </div>
            ))}
            <p className="text-xs text-arcom-gray mt-2">Status geral: {progresso.status}</p>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
