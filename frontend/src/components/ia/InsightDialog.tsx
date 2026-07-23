import { useEffect, useState } from "react";
import { Phone, Globe, Instagram, Linkedin, Mail, Loader2 } from "lucide-react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Alert } from "@/components/ui/alert";
import { api } from "@/lib/api";
import { telMovel } from "@/lib/prospeccao";
import type { Empresa } from "@/lib/cnpj";
import type { Insight } from "@/lib/ia";

function TelefoneLink({ tel }: { tel: string }) {
  const d = tel.replace(/\D/g, "");
  if (!d) return <span>{tel}</span>;
  return (
    <span className="inline-flex items-center gap-2">
      <a href={`tel:${d}`} className="text-verde-arcom font-semibold hover:underline">
        {tel}
      </a>
      {telMovel(d) && (
        <a href={`https://wa.me/55${d}`} target="_blank" rel="noopener noreferrer" title="WhatsApp">
          <Phone className="h-3.5 w-3.5 text-verde-lima" />
        </a>
      )}
    </span>
  );
}

export function InsightDialog({ empresa, onClose }: { empresa: Empresa | null; onClose: () => void }) {
  const [carregando, setCarregando] = useState(false);
  const [erro, setErro] = useState<string | null>(null);
  const [insight, setInsight] = useState<Insight | null>(null);

  useEffect(() => {
    if (!empresa) {
      setInsight(null);
      return;
    }
    setCarregando(true);
    setErro(null);
    api
      .post("/ia/insight", { empresa })
      .then(({ data }) => setInsight(data))
      .catch((e) => setErro(e?.response?.data?.erro ?? "Falha ao gerar insight."))
      .finally(() => setCarregando(false));
  }, [empresa]);

  if (!empresa) return null;

  const c = insight?.contatos;
  const telWeb = c?.telefoneWeb;
  const telOficial = c?.telefoneOficial;
  const webBateOficial = telWeb && telOficial && telWeb.replace(/\D/g, "") === telOficial.replace(/\D/g, "");

  return (
    <Dialog open={!!empresa} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-xl max-h-[85vh] overflow-auto">
        <DialogHeader>
          <DialogTitle>Insight comercial</DialogTitle>
          <DialogDescription>{empresa.razao}</DialogDescription>
        </DialogHeader>

        {carregando && (
          <div className="flex items-center gap-2 text-arcom-gray text-sm py-6">
            <Loader2 className="h-4 w-4 animate-spin" /> Buscando na web e analisando…
          </div>
        )}
        {erro && <Alert variant="danger">{erro}</Alert>}

        {insight && (
          <div className="flex flex-col gap-4">
            <div>
              <h4 className="font-bold text-sm text-verde-escuro mb-1">Panorama comercial</h4>
              <p className="text-sm text-arcom-gray">{insight.resumo}</p>
            </div>

            <div className="border border-verde-arcom rounded-md p-3">
              <h4 className="font-bold text-sm text-verde-arcom mb-1">Telefone para contato</h4>
              {telWeb && (
                <div className="text-base">
                  <TelefoneLink tel={telWeb} />{" "}
                  <span className="text-xs text-arcom-gray">
                    · achado na web {webBateOficial ? "(confere com a Receita)" : "— confirme"}
                  </span>
                </div>
              )}
              {telOficial && !webBateOficial && (
                <div className="text-sm mt-1">
                  <TelefoneLink tel={telOficial} /> <span className="text-xs text-arcom-gray">· oficial (Receita)</span>
                </div>
              )}
              {!telOficial && !telWeb && (
                <p className="text-sm text-arcom-gray">Não achei telefone na web nem na Receita.</p>
              )}
            </div>

            {(c?.site || c?.instagram || c?.linkedin || c?.email) && (
              <div>
                <h4 className="font-bold text-sm text-verde-escuro mb-1">Outros contatos (web)</h4>
                <ul className="flex flex-col gap-1 text-sm">
                  {c?.site && (
                    <li className="flex items-center gap-2">
                      <Globe className="h-3.5 w-3.5 text-arcom-gray" />
                      <a href={c.site} target="_blank" rel="noopener noreferrer" className="text-verde-arcom hover:underline">
                        {c.site}
                      </a>
                    </li>
                  )}
                  {c?.instagram && (
                    <li className="flex items-center gap-2">
                      <Instagram className="h-3.5 w-3.5 text-arcom-gray" /> {c.instagram}
                    </li>
                  )}
                  {c?.linkedin && (
                    <li className="flex items-center gap-2">
                      <Linkedin className="h-3.5 w-3.5 text-arcom-gray" /> {c.linkedin}
                    </li>
                  )}
                  {c?.email && (
                    <li className="flex items-center gap-2">
                      <Mail className="h-3.5 w-3.5 text-arcom-gray" /> {c.email}
                    </li>
                  )}
                </ul>
              </div>
            )}

            {insight.pontosFortes.length > 0 && (
              <div>
                <h4 className="font-bold text-sm text-verde-escuro mb-1">Pontos fortes</h4>
                <ul className="list-disc list-inside text-sm text-arcom-gray">
                  {insight.pontosFortes.map((p, i) => (
                    <li key={i}>{p}</li>
                  ))}
                </ul>
              </div>
            )}

            {insight.sinaisAtencao.length > 0 && (
              <div>
                <h4 className="font-bold text-sm text-verde-escuro mb-1">Sinais de atenção</h4>
                <ul className="list-disc list-inside text-sm text-arcom-gray">
                  {insight.sinaisAtencao.map((p, i) => (
                    <li key={i}>{p}</li>
                  ))}
                </ul>
              </div>
            )}

            <div>
              <h4 className="font-bold text-sm text-verde-escuro mb-1">Abordagem sugerida</h4>
              <p className="text-sm text-arcom-gray">{insight.abordagemSugerida}</p>
            </div>

            {insight.perguntasQualificacao.length > 0 && (
              <div>
                <h4 className="font-bold text-sm text-verde-escuro mb-1">Perguntas de qualificação</h4>
                <ul className="list-disc list-inside text-sm text-arcom-gray">
                  {insight.perguntasQualificacao.map((p, i) => (
                    <li key={i}>{p}</li>
                  ))}
                </ul>
              </div>
            )}

            <Badge variant="outline" className="w-fit">
              Confiança: {insight.nivelConfianca}
            </Badge>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
