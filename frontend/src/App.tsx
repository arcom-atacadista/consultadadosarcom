import { useEffect, useState } from "react";
import { api } from "@/lib/api";

type Health = "verificando" | "ok" | "erro";

export default function App() {
  const [health, setHealth] = useState<Health>("verificando");

  useEffect(() => {
    api
      .get("/health")
      .then(() => setHealth("ok"))
      .catch(() => setHealth("erro"));
  }, []);

  return (
    <div className="min-h-screen bg-surface flex items-center justify-center">
      <div className="bg-white border border-surface-border rounded-lg shadow-sm p-8 max-w-md text-center">
        <h1 className="font-black text-3xl text-verde-escuro">CDA</h1>
        <p className="text-arcom-gray mt-1">Consulta Dados Arcom</p>
        <p className="mt-6 text-sm">
          Backend:{" "}
          <span
            className={
              health === "ok"
                ? "text-verde-arcom font-semibold"
                : health === "erro"
                  ? "text-danger font-semibold"
                  : "text-arcom-gray"
            }
          >
            {health === "ok" ? "conectado" : health === "erro" ? "indisponível" : "verificando…"}
          </span>
        </p>
      </div>
    </div>
  );
}
