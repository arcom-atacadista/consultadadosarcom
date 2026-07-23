import { useEffect, useState } from "react";
import { Loader2, Trophy } from "lucide-react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Alert } from "@/components/ui/alert";
import { api } from "@/lib/api";
import { formatarCNPJ } from "@/lib/cnpj";
import { empresaParaResumo, type ItemRanking } from "@/lib/ia";
import type { Empresa } from "@/lib/cnpj";

export function RankingDialog({ empresas, onClose }: { empresas: Empresa[] | null; onClose: () => void }) {
  const [carregando, setCarregando] = useState(false);
  const [erro, setErro] = useState<string | null>(null);
  const [ranking, setRanking] = useState<ItemRanking[]>([]);

  useEffect(() => {
    if (!empresas) return;
    setCarregando(true);
    setErro(null);
    api
      .post("/ia/ranking", { empresas: empresas.map(empresaParaResumo) })
      .then(({ data }) => setRanking(data.ranking ?? []))
      .catch((e) => setErro(e?.response?.data?.erro ?? "Falha ao gerar ranking."))
      .finally(() => setCarregando(false));
  }, [empresas]);

  return (
    <Dialog open={!!empresas} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-lg max-h-[80vh] overflow-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Trophy className="h-5 w-5 text-verde-lima" /> Ranking de leads (IA)
          </DialogTitle>
          <DialogDescription>Priorização por IA com base nos dados já consultados.</DialogDescription>
        </DialogHeader>
        {carregando && (
          <div className="flex items-center gap-2 text-arcom-gray text-sm py-6">
            <Loader2 className="h-4 w-4 animate-spin" /> Analisando empresas…
          </div>
        )}
        {erro && <Alert variant="danger">{erro}</Alert>}
        {!carregando && !erro && ranking.length === 0 && (
          <p className="text-sm text-arcom-gray">Não foi possível montar um ranking.</p>
        )}
        <div className="flex flex-col gap-2">
          {ranking.map((item) => (
            <div key={item.cnpj} className="border border-surface-border rounded-md p-3">
              <div className="font-semibold text-sm text-verde-escuro">
                #{item.posicao} — {item.razao}{" "}
                <span className="font-mono text-xs text-arcom-gray">{formatarCNPJ(item.cnpj)}</span>
              </div>
              <p className="text-sm text-arcom-gray mt-1">{item.motivo}</p>
            </div>
          ))}
        </div>
        {ranking.length > 0 && (
          <p className="text-xs text-arcom-gray">
            Ranking gerado por IA com base nos dados públicos já consultados — use como apoio à priorização.
          </p>
        )}
      </DialogContent>
    </Dialog>
  );
}
