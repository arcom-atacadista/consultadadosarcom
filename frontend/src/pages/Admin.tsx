import { ShieldCheck } from "lucide-react";
import { EmConstrucao } from "@/components/EmConstrucao";

export default function Admin() {
  return (
    <EmConstrucao
      icon={ShieldCheck}
      titulo="Administração"
      descricao="Dashboard multiusuário: contas, atividades e presença online."
      fase="Chega na Fase 7"
    />
  );
}
