import { RefreshCcw } from "lucide-react";
import { EmConstrucao } from "@/components/EmConstrucao";

export default function Conversao() {
  return (
    <EmConstrucao
      icon={RefreshCcw}
      titulo="Conversão"
      descricao="Retorno da prospecção por assessor e cruzamento com a planilha de vendas."
      fase="Chega na Fase 7"
    />
  );
}
