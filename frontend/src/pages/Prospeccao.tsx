import { MapPinned } from "lucide-react";
import { EmConstrucao } from "@/components/EmConstrucao";

export default function Prospeccao() {
  return (
    <EmConstrucao
      icon={MapPinned}
      titulo="Prospecção"
      descricao="Busca por cidade/UF/CNAE, score de loja física e pré-cadastro."
      fase="Chega na Fase 5"
    />
  );
}
