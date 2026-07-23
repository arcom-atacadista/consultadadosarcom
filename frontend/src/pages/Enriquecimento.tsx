import { Layers } from "lucide-react";
import { EmConstrucao } from "@/components/EmConstrucao";

export default function Enriquecimento() {
  return (
    <EmConstrucao
      icon={Layers}
      titulo="Enriquecimento"
      descricao="Dossiê Trace360 por CNPJ, com status de processamento."
      fase="Chega na Fase 6"
    />
  );
}
