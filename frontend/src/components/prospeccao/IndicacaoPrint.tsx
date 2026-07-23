import { useEffect, useMemo, useState } from "react";
import { Printer, X, Navigation } from "lucide-react";
import { Button } from "@/components/ui/button";
import { api } from "@/lib/api";
import { formatarCNPJ } from "@/lib/cnpj";
import { ramoDoCnae, rotaProximidade, type Prospect } from "@/lib/prospeccao";
import type { Empresa } from "@/lib/cnpj";

const MAX_PARADAS = 9; // Google Maps aceita ~10 pontos por link (incl. origem)

export function IndicacaoPrint({ prospects, onClose }: { prospects: Prospect[]; onClose: () => void }) {
  const [origem, setOrigem] = useState<{ lat: number; lng: number } | null>(null);
  const [socios, setSocios] = useState<Record<string, Empresa>>({});

  useEffect(() => {
    api
      .post("/cnpj/consultar", { cnpjs: prospects.map((p) => p.cnpj), fonte: "arcom" })
      .then(({ data }) => {
        const mapa: Record<string, Empresa> = {};
        (data.resultados as Empresa[]).forEach((e) => {
          if (e.encontrado) mapa[e.cnpj] = e;
        });
        setSocios(mapa);
      })
      .catch(() => {
        // sem quadro societário, segue sem ele (não bloqueia o PDF)
      });
  }, [prospects]);

  function usarLocalizacao() {
    if (!navigator.geolocation) return;
    navigator.geolocation.getCurrentPosition((pos) => {
      setOrigem({ lat: pos.coords.latitude, lng: pos.coords.longitude });
    });
  }

  const { grupos, mapsUrl, titulo } = useMemo(() => {
    const cidades = [...new Set(prospects.map((p) => `${p.cidade} - ${p.uf}`))];
    const titulo = cidades.length === 1 ? cidades[0] : cidades.join(" · ");

    const comCoord = prospects.filter((p) => p.latitude != null && p.longitude != null);
    const semCoord = prospects.filter(
      (p) => (p.latitude == null || p.longitude == null) && p.endereco && p.endereco !== "—",
    );
    const rota = rotaProximidade(comCoord, origem).concat(semCoord).filter((p) => p.endereco && p.endereco !== "—");
    const usados = rota.slice(0, MAX_PARADAS);
    let mapsUrl = "";
    if (usados.length) {
      const partes: string[] = [];
      if (origem) partes.push(`${origem.lat},${origem.lng}`);
      usados.forEach((p) => partes.push(encodeURIComponent(p.endereco)));
      mapsUrl = "https://www.google.com/maps/dir/" + partes.join("/");
    }

    const grupos: Record<string, Prospect[]> = {};
    prospects.forEach((p) => {
      const ramo = ramoDoCnae(p.cnaeCodigo, p.atividade);
      (grupos[ramo] ??= []).push(p);
    });

    return { grupos, mapsUrl, titulo };
  }, [prospects, origem]);

  return (
    <div className="fixed inset-0 z-50 bg-white overflow-auto">
      <div className="no-print sticky top-0 flex items-center justify-between border-b border-surface-border bg-white px-6 py-3">
        <h2 className="font-bold text-verde-escuro">Indicação de clientes — {titulo}</h2>
        <div className="flex gap-2">
          <Button variant="secondary" size="sm" onClick={usarLocalizacao}>
            <Navigation className="h-4 w-4" /> Usar minha localização
          </Button>
          <Button size="sm" onClick={() => window.print()}>
            <Printer className="h-4 w-4" /> Imprimir / Salvar PDF
          </Button>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X className="h-4 w-4" /> Fechar
          </Button>
        </div>
      </div>

      <div className="print-area max-w-3xl mx-auto p-8 font-sans text-verde-escuro">
        <div className="flex items-center gap-3 border-b-[3px] border-verde-arcom pb-3 mb-4">
          <div>
            <div className="font-black text-xl text-verde-arcom">Indicação de clientes para {titulo}</div>
            <div className="text-xs text-arcom-gray">
              Gerado pela ferramenta CDA — {prospects.length} empresa(s) · {new Date().toLocaleDateString("pt-BR")}
              {origem ? " · ordenado por proximidade" : ""}
            </div>
          </div>
        </div>

        {mapsUrl && (
          <div className="border border-surface-border rounded-md p-3 mb-4 bg-surface">
            <div className="text-sm font-bold mb-1">Rota de visita</div>
            <a href={mapsUrl} target="_blank" rel="noopener noreferrer" className="text-verde-arcom font-semibold text-sm">
              Abrir rota no Google Maps
            </a>
          </div>
        )}

        {Object.entries(grupos).map(([ramo, itens]) => (
          <div key={ramo} className="mb-4 break-inside-avoid">
            <div className="font-bold text-verde-escuro mb-2">{ramo}:</div>
            {itens.map((p) => {
              const enriquecido = socios[p.cnpj];
              const listaSocios = enriquecido?.socios?.filter((s) => s.nome_socio) ?? [];
              return (
                <div key={p.cnpj} className="border border-surface-border rounded-md p-3 mb-2 break-inside-avoid">
                  <div className="font-mono text-xs text-verde-arcom">{formatarCNPJ(p.cnpj)}</div>
                  <div className="font-bold">
                    {p.razao}
                    {p.nomeFantasia && <span className="font-medium text-arcom-gray"> ({p.nomeFantasia})</span>}
                  </div>
                  <div className="text-sm mt-1">
                    <strong>Endereço:</strong> {p.endereco || "—"}
                  </div>
                  <div className="text-sm">
                    <strong>CEP:</strong> {p.cep || "—"}
                  </div>
                  <div className="text-sm">
                    <strong>Telefone:</strong> {p.telefone || "—"}
                  </div>
                  {listaSocios.length > 0 && (
                    <div className="text-sm mt-1">
                      <strong>Quadro societário:</strong>{" "}
                      {listaSocios.map((s) => s.nome_socio).join(" · ")}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        ))}
      </div>
    </div>
  );
}
