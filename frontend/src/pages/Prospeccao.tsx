import { useMemo, useState } from "react";
import { Search, Phone, Mail, Eye, UserPlus, Star, Copy, FileText, Loader2 } from "lucide-react";
import { api } from "@/lib/api";
import { formatarCNPJ } from "@/lib/cnpj";
import { RAMOS_BUSCA, parseCidades, telMovel, type Prospect } from "@/lib/prospeccao";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Alert } from "@/components/ui/alert";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ScoreBadge } from "@/components/prospeccao/ScoreBadge";
import { PreCadastroDialog } from "@/components/prospeccao/PreCadastroDialog";
import { FachadaMapaModal } from "@/components/prospeccao/FachadaMapaModal";
import { IndicacaoPrint } from "@/components/prospeccao/IndicacaoPrint";

export default function Prospeccao() {
  const [cidadesTexto, setCidadesTexto] = useState("");
  const [cnaes, setCnaes] = useState<string[]>([]);
  const [carregando, setCarregando] = useState(false);
  const [erro, setErro] = useState<string | null>(null);
  const [prospects, setProspects] = useState<Prospect[]>([]);

  const [cortarRedes, setCortarRedes] = useState(true);
  const [somenteNovas, setSomenteNovas] = useState(false);
  const [buscaTexto, setBuscaTexto] = useState("");
  const [selecionados, setSelecionados] = useState<Set<string>>(new Set());

  const [prospectPreCadastro, setProspectPreCadastro] = useState<Prospect | null>(null);
  const [prospectFachada, setProspectFachada] = useState<Prospect | null>(null);
  const [mostrarIndicacao, setMostrarIndicacao] = useState(false);

  function alternarCnae(valor: string) {
    setCnaes((atual) => (atual.includes(valor) ? atual.filter((c) => c !== valor) : [...atual, valor]));
  }

  async function buscar() {
    const cidades = parseCidades(cidadesTexto);
    if (cidades.length === 0) {
      setErro("Informe ao menos uma cidade (uma por linha, no formato Cidade,UF).");
      return;
    }
    const semUF = cidades.filter((c) => !c.uf);
    if (semUF.length) {
      setErro(`Falta a UF para: ${semUF.map((c) => c.cidade).join(", ")} — use "Cidade,UF".`);
      return;
    }
    if (cnaes.length === 0) {
      setErro("Selecione ao menos um ramo.");
      return;
    }

    setErro(null);
    setCarregando(true);
    setSelecionados(new Set());
    try {
      const { data } = await api.post("/prospeccao/buscar", { cidades, cnaes });
      setProspects(data.prospects ?? []);
    } catch (e: any) {
      setErro(e?.response?.data?.erro ?? "Falha ao buscar prospecção.");
    } finally {
      setCarregando(false);
    }
  }

  const visiveis = useMemo(() => {
    const termos = buscaTexto.trim().toLowerCase().split(/\s+/).filter(Boolean);
    return prospects
      .filter((p) => (cortarRedes ? !p.redeGrande : true))
      .filter((p) => (somenteNovas ? p.nova : true))
      .filter((p) => {
        if (!termos.length) return true;
        const alvo = `${p.razao} ${p.nomeFantasia} ${p.cidade} ${p.atividade} ${p.cnpj}`.toLowerCase();
        return termos.every((t) => alvo.includes(t));
      })
      .sort((a, b) => b.score - a.score);
  }, [prospects, cortarRedes, somenteNovas, buscaTexto]);

  const paraAcao = useMemo(
    () => (selecionados.size ? visiveis.filter((p) => selecionados.has(p.cnpj)) : visiveis),
    [visiveis, selecionados],
  );

  function alternarSelecionado(cnpj: string) {
    setSelecionados((atual) => {
      const novo = new Set(atual);
      if (novo.has(cnpj)) novo.delete(cnpj);
      else novo.add(cnpj);
      return novo;
    });
  }

  function copiarCNPJs() {
    navigator.clipboard.writeText(paraAcao.map((p) => p.cnpj).join("\n"));
  }

  async function salvarLista() {
    const nome = window.prompt("Nome da lista:");
    if (!nome) return;
    await api.post("/prospeccao/listas", {
      nome,
      filtros: { cidades: parseCidades(cidadesTexto), cnaes },
      itens: paraAcao.map((p) => p.cnpj),
    });
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-[340px_1fr] gap-6">
      <div className="flex flex-col gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Prospecção</CardTitle>
            <CardDescription>Cidades (uma por linha, formato Cidade,UF) e ramos.</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            <Textarea
              value={cidadesTexto}
              onChange={(e) => setCidadesTexto(e.target.value)}
              placeholder={"Uberlândia,MG\nSão Simão,GO"}
              rows={3}
            />
            <div className="flex flex-col gap-1.5 max-h-48 overflow-auto border border-surface-border rounded-md p-2">
              {RAMOS_BUSCA.map((r) => (
                <label key={r.value} className="flex items-center gap-2 text-sm text-verde-escuro cursor-pointer">
                  <input
                    type="checkbox"
                    checked={cnaes.includes(r.value)}
                    onChange={() => alternarCnae(r.value)}
                  />
                  {r.label}
                </label>
              ))}
            </div>
            <Button onClick={buscar} disabled={carregando}>
              {carregando ? <Loader2 className="h-4 w-4 animate-spin" /> : <Search className="h-4 w-4" />}
              {carregando ? "Buscando…" : "Buscar"}
            </Button>
            {erro && <Alert variant="danger">{erro}</Alert>}
          </CardContent>
        </Card>

        {prospects.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Filtros de exibição</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-2">
              <label className="flex items-center gap-2 text-sm cursor-pointer">
                <input type="checkbox" checked={cortarRedes} onChange={(e) => setCortarRedes(e.target.checked)} />
                Cortar grandes redes
              </label>
              <label className="flex items-center gap-2 text-sm cursor-pointer">
                <input type="checkbox" checked={somenteNovas} onChange={(e) => setSomenteNovas(e.target.checked)} />
                Só empresas novas (últimos 12 meses)
              </label>
              <Input
                placeholder="Buscar nos resultados…"
                value={buscaTexto}
                onChange={(e) => setBuscaTexto(e.target.value)}
              />
            </CardContent>
          </Card>
        )}
      </div>

      <Card>
        <CardHeader className="flex-row items-center justify-between">
          <div>
            <CardTitle>Resultados</CardTitle>
            <CardDescription>
              {visiveis.length} prospect(s) — {selecionados.size ? `${selecionados.size} selecionado(s)` : "nenhum marcado, ações usam todos os visíveis"}
            </CardDescription>
          </div>
          {visiveis.length > 0 && (
            <div className="flex gap-2">
              <Button variant="secondary" size="sm" onClick={copiarCNPJs}>
                <Copy className="h-4 w-4" /> Copiar CNPJs
              </Button>
              <Button variant="secondary" size="sm" onClick={salvarLista}>
                <Star className="h-4 w-4" /> Salvar lista
              </Button>
              <Button size="sm" onClick={() => setMostrarIndicacao(true)}>
                <FileText className="h-4 w-4" /> Indicação (PDF)
              </Button>
            </div>
          )}
        </CardHeader>
        <CardContent>
          {visiveis.length === 0 ? (
            <p className="text-sm text-arcom-gray">Nenhum resultado ainda.</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead />
                  <TableHead>Potencial</TableHead>
                  <TableHead>CNPJ</TableHead>
                  <TableHead>Razão social</TableHead>
                  <TableHead>Bairro</TableHead>
                  <TableHead>Atividade</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {visiveis.map((p) => {
                  const tel = p.telefone?.replace(/\D/g, "") ?? "";
                  return (
                    <TableRow key={p.cnpj}>
                      <TableCell>
                        <input
                          type="checkbox"
                          checked={selecionados.has(p.cnpj)}
                          onChange={() => alternarSelecionado(p.cnpj)}
                        />
                      </TableCell>
                      <TableCell>
                        <ScoreBadge prospect={p} />
                      </TableCell>
                      <TableCell className="font-mono text-xs">{formatarCNPJ(p.cnpj)}</TableCell>
                      <TableCell>
                        {p.razao}
                        {p.nova && (
                          <Badge variant="accent" className="ml-2">
                            Nova
                          </Badge>
                        )}
                        {p.redeGrande && (
                          <Badge variant="outline" className="ml-2">
                            Rede grande
                          </Badge>
                        )}
                      </TableCell>
                      <TableCell>{p.bairro}</TableCell>
                      <TableCell>{p.atividade}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <button title="Ver fachada/mapa" onClick={() => setProspectFachada(p)}>
                            <Eye className="h-4 w-4 text-arcom-gray hover:text-verde-arcom" />
                          </button>
                          {tel && (
                            <a title={`Ligar para ${tel}`} href={`tel:${tel}`}>
                              <Phone className="h-4 w-4 text-arcom-gray hover:text-verde-arcom" />
                            </a>
                          )}
                          {tel && telMovel(tel) && (
                            <a title="WhatsApp" href={`https://wa.me/55${tel}`} target="_blank" rel="noopener noreferrer">
                              <Phone className="h-4 w-4 text-verde-arcom" />
                            </a>
                          )}
                          {p.email && (
                            <a title={`E-mail: ${p.email}`} href={`mailto:${p.email}`}>
                              <Mail className="h-4 w-4 text-arcom-gray hover:text-verde-arcom" />
                            </a>
                          )}
                          <button title="Pré-cadastrar" onClick={() => setProspectPreCadastro(p)}>
                            <UserPlus className="h-4 w-4 text-arcom-gray hover:text-verde-arcom" />
                          </button>
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <PreCadastroDialog prospect={prospectPreCadastro} onClose={() => setProspectPreCadastro(null)} />
      <FachadaMapaModal prospect={prospectFachada} onClose={() => setProspectFachada(null)} />
      {mostrarIndicacao && <IndicacaoPrint prospects={paraAcao} onClose={() => setMostrarIndicacao(false)} />}
    </div>
  );
}
