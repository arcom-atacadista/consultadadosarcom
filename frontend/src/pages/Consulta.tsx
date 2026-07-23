import { Search } from "lucide-react";
import { EmConstrucao } from "@/components/EmConstrucao";

export default function Consulta() {
  return (
    <EmConstrucao
      icon={Search}
      titulo="Consulta de CNPJ"
      descricao="Consulta individual e em lote, cache de 24h, favoritos e histórico."
      fase="Chega na Fase 4"
    />
  );
}
