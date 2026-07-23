import { cn } from "@/lib/cn";
import type { Prospect } from "@/lib/prospeccao";

const CORES: Record<Prospect["temperatura"], string> = {
  quente: "bg-verde-arcom",
  morno: "bg-verde-lima",
  frio: "bg-danger",
};

export function ScoreBadge({ prospect }: { prospect: Prospect }) {
  const dica = prospect.sinais.length ? prospect.sinais.join(" · ") : "poucos sinais de loja física";
  return (
    <div className="flex items-center gap-2" title={`Provável loja física: ${prospect.score}/100 — ${dica}`}>
      <span className={cn("h-2.5 w-2.5 rounded-full", CORES[prospect.temperatura])} />
      <span className="font-semibold text-verde-escuro">{prospect.score}</span>
    </div>
  );
}
