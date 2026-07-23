import { useEffect, useRef, useState } from "react";
import { MessageCircle, X, Send, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import type { MensagemChat } from "@/lib/ia";

const CHAVE_HISTORICO = "chatHistorico";
const HISTORICO_MAX = 40;

function carregarHistorico(): MensagemChat[] {
  try {
    return JSON.parse(localStorage.getItem(CHAVE_HISTORICO) || "[]");
  } catch {
    return [];
  }
}

function linkify(texto: string) {
  const partes = texto.split(/(https?:\/\/[^\s]+)/g);
  return partes.map((parte, i) =>
    /^https?:\/\//.test(parte) ? (
      <a key={i} href={parte} target="_blank" rel="noopener noreferrer" className="underline break-all">
        {parte}
      </a>
    ) : (
      <span key={i}>{parte}</span>
    ),
  );
}

export function ChatWidget() {
  const [aberto, setAberto] = useState(false);
  const [historico, setHistorico] = useState<MensagemChat[]>(carregarHistorico);
  const [texto, setTexto] = useState("");
  const [carregando, setCarregando] = useState(false);
  const boxRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    localStorage.setItem(CHAVE_HISTORICO, JSON.stringify(historico.slice(-HISTORICO_MAX)));
    boxRef.current?.scrollTo({ top: boxRef.current.scrollHeight });
  }, [historico]);

  async function enviar() {
    const mensagem = texto.trim();
    if (!mensagem || carregando) return;
    setTexto("");
    const novoHistorico = [...historico, { role: "user" as const, texto: mensagem }];
    setHistorico(novoHistorico);
    setCarregando(true);
    try {
      const { data } = await api.post("/ia/chat", {
        mensagem,
        historico: novoHistorico.slice(-12),
      });
      setHistorico((h) => [...h, { role: "assistant", texto: data.resposta }]);
    } catch (e: any) {
      setHistorico((h) => [
        ...h,
        { role: "assistant", texto: `Erro ao falar com a IA: ${e?.response?.data?.erro ?? "falha desconhecida"}` },
      ]);
    } finally {
      setCarregando(false);
    }
  }

  function limpar() {
    setHistorico([]);
  }

  if (!aberto) {
    return (
      <button
        onClick={() => setAberto(true)}
        className="fixed bottom-6 right-6 z-40 flex h-14 w-14 items-center justify-center rounded-full bg-verde-arcom text-white shadow-lg hover:brightness-[.85] no-print"
        title="Conversar com a IA"
      >
        <MessageCircle className="h-6 w-6" />
      </button>
    );
  }

  return (
    <div className="no-print fixed bottom-6 right-6 z-40 flex h-[480px] w-96 flex-col rounded-lg border border-surface-border bg-white shadow-lg">
      <div className="flex items-center justify-between border-b border-surface-border px-4 py-3">
        <h3 className="font-bold text-sm text-verde-escuro">Assistente CDA</h3>
        <div className="flex items-center gap-2">
          <button onClick={limpar} title="Limpar conversa">
            <Trash2 className="h-4 w-4 text-arcom-gray hover:text-danger" />
          </button>
          <button onClick={() => setAberto(false)} title="Fechar">
            <X className="h-4 w-4 text-arcom-gray hover:text-verde-escuro" />
          </button>
        </div>
      </div>
      <div ref={boxRef} className="flex-1 overflow-auto px-4 py-3 flex flex-col gap-3">
        {historico.length === 0 && (
          <p className="text-sm text-arcom-gray">
            Pergunte qualquer coisa — ou peça um resumo das empresas que você já consultou nesta sessão.
          </p>
        )}
        {historico.map((m, i) => (
          <div key={i} className={m.role === "user" ? "self-end max-w-[85%]" : "self-start max-w-[85%]"}>
            <div
              className={
                m.role === "user"
                  ? "rounded-md bg-verde-arcom text-white px-3 py-2 text-sm"
                  : "rounded-md bg-surface text-verde-escuro px-3 py-2 text-sm"
              }
            >
              {linkify(m.texto)}
            </div>
          </div>
        ))}
        {carregando && <p className="text-xs text-arcom-gray">Digitando…</p>}
      </div>
      <div className="flex items-center gap-2 border-t border-surface-border p-3">
        <Input
          value={texto}
          onChange={(e) => setTexto(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && enviar()}
          placeholder="Digite sua pergunta…"
          disabled={carregando}
        />
        <Button size="icon" onClick={enviar} disabled={carregando}>
          <Send className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
