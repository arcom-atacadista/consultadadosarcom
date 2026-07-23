import { useCallback, useEffect, useRef, useState } from "react";
import { Layers, Loader2, GitBranch, FileDown, Code, RotateCw } from "lucide-react";
import { api } from "@/lib/api";
import { formatarCNPJ, parseCNPJs, cnpjValido } from "@/lib/cnpj";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Alert } from "@/components/ui/alert";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ProgressoDialog } from "@/components/enriquecimento/ProgressoDialog";

type ItemEnriquecimento = {
  id: string;
  cnpj: string;
  clienteId: string;
  status: string;
  razaoSocial: string;
  criadoEm: string;
};

const EM_ANDAMENTO = new Set(["pendente", "enfileirado", "processando", "ja_ativo", "ambiguo_pausado"]);
const CONCLUIDO = new Set(["concluido", "em_cache"]);
const FALHOU = new Set(["erro", "nao_enriquecivel", "cancelado"]);

function badgeStatus(status: string) {
  const s = (status || "").toLowerCase();
  if (CONCLUIDO.has(s)) return <Badge variant="accent">Concluído</Badge>;
  if (EM_ANDAMENTO.has(s)) return <Badge variant="outline">Processando…</Badge>;
  if (s === "erro") return <Badge variant="danger">Erro</Badge>;
  if (FALHOU.has(s)) return <Badge variant="default">Sem dados</Badge>;
  return <Badge variant="default">{status || "—"}</Badge>;
}

export default function Enriquecimento() {
  const [texto, setTexto] = useState("");
  const [enviando, setEnviando] = useState(false);
  const [erro, setErro] = useState<string | null>(null);
  const [itens, setItens] = useState<ItemEnriquecimento[]>([]);
  const [progressoAlvo, setProgressoAlvo] = useState<ItemEnriquecimento | null>(null);
  const pollRef = useRef<ReturnType<typeof setTimeout>>();

  const carregar = useCallback(async () => {
    try {
      const { data } = await api.get("/enriquecimento");
      setItens(data ?? []);
      clearTimeout(pollRef.current);
      if ((data ?? []).some((x: ItemEnriquecimento) => EM_ANDAMENTO.has(x.status))) {
        pollRef.current = setTimeout(carregar, 25000);
      }
    } catch {
      // mantém a última lista carregada
    }
  }, []);

  useEffect(() => {
    carregar();
    return () => clearTimeout(pollRef.current);
  }, [carregar]);

  async function enriquecer() {
    const cnpjs = parseCNPJs(texto).filter(cnpjValido);
    if (cnpjs.length === 0) {
      setErro("Cole ao menos um CNPJ válido.");
      return;
    }
    if (cnpjs.length > 500) {
      setErro("Máximo de 500 CNPJs por vez.");
      return;
    }
    setErro(null);
    setEnviando(true);
    try {
      await api.post("/enriquecimento", { cnpjs });
      setTexto("");
      await carregar();
    } catch (e: any) {
      setErro(e?.response?.data?.erro ?? "Falha ao enviar para enriquecimento.");
    } finally {
      setEnviando(false);
    }
  }

  async function baixarDossie(item: ItemEnriquecimento) {
    const resp = await api.get(`/enriquecimento/${item.clienteId}/dossie`, { responseType: "blob" });
    const url = URL.createObjectURL(resp.data);
    const a = document.createElement("a");
    a.href = url;
    a.download = `dossie-${item.clienteId}.pdf`;
    a.click();
    setTimeout(() => URL.revokeObjectURL(url), 8000);
  }

  async function verResultado(item: ItemEnriquecimento) {
    const resp = await api.get(`/enriquecimento/${item.clienteId}/resultado`, { responseType: "blob" });
    const url = URL.createObjectURL(resp.data);
    window.open(url, "_blank");
    setTimeout(() => URL.revokeObjectURL(url), 30000);
  }

  async function reprocessar(item: ItemEnriquecimento) {
    await api.post(`/enriquecimento/${item.clienteId}/reprocessar`);
    await carregar();
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-[360px_1fr] gap-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Layers className="h-5 w-5 text-verde-arcom" /> Enriquecimento
          </CardTitle>
          <CardDescription>Dossiê completo por CNPJ (Trace360) — até 500 por vez.</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <Textarea
            value={texto}
            onChange={(e) => setTexto(e.target.value)}
            placeholder={"11.222.333/0001-81\n19.131.243/0001-97"}
            rows={6}
          />
          <Button onClick={enriquecer} disabled={enviando}>
            {enviando ? <Loader2 className="h-4 w-4 animate-spin" /> : <Layers className="h-4 w-4" />}
            {enviando ? "Enviando…" : "Enriquecer"}
          </Button>
          {erro && <Alert variant="danger">{erro}</Alert>}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Meus enriquecimentos</CardTitle>
          <CardDescription>{itens.length} CNPJ(s) enviado(s) por você.</CardDescription>
        </CardHeader>
        <CardContent>
          {itens.length === 0 ? (
            <p className="text-sm text-arcom-gray">
              Você ainda não enriqueceu nenhum CNPJ. Cole os CNPJs ao lado e clique em "Enriquecer".
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>CNPJ</TableHead>
                  <TableHead>Razão social</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {itens.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell className="font-mono text-xs">{formatarCNPJ(item.cnpj)}</TableCell>
                    <TableCell>{item.razaoSocial || "—"}</TableCell>
                    <TableCell>{badgeStatus(item.status)}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <button title="Ver o fluxo de etapas" onClick={() => setProgressoAlvo(item)}>
                          <GitBranch className="h-4 w-4 text-arcom-gray hover:text-verde-arcom" />
                        </button>
                        {CONCLUIDO.has(item.status) && (
                          <>
                            <button title="Baixar dossiê (PDF)" onClick={() => baixarDossie(item)}>
                              <FileDown className="h-4 w-4 text-arcom-gray hover:text-verde-arcom" />
                            </button>
                            <button title="Ver os dados (JSON)" onClick={() => verResultado(item)}>
                              <Code className="h-4 w-4 text-arcom-gray hover:text-verde-arcom" />
                            </button>
                          </>
                        )}
                        {FALHOU.has(item.status) && (
                          <button title="Tentar de novo" onClick={() => reprocessar(item)}>
                            <RotateCw className="h-4 w-4 text-arcom-gray hover:text-verde-arcom" />
                          </button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <ProgressoDialog
        clienteId={progressoAlvo?.clienteId ?? null}
        cnpj={progressoAlvo?.cnpj ?? ""}
        onClose={() => setProgressoAlvo(null)}
      />
    </div>
  );
}
