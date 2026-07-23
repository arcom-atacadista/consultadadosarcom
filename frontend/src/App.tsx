import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AppShell } from "@/components/layout/AppShell";
import Cover from "@/pages/Cover";
import Login from "@/pages/Login";
import Pendente from "@/pages/Pendente";
import Consulta from "@/pages/Consulta";
import Prospeccao from "@/pages/Prospeccao";
import Enriquecimento from "@/pages/Enriquecimento";
import Conversao from "@/pages/Conversao";
import Admin from "@/pages/Admin";

// Guarda de rota (redirecionar não-logado/não-aprovado) entra na Fase 3 (auth).
export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Cover />} />
        <Route path="/login" element={<Login />} />
        <Route path="/pendente" element={<Pendente />} />

        <Route element={<AppShell />}>
          <Route path="/consulta" element={<Consulta />} />
          <Route path="/prospeccao" element={<Prospeccao />} />
          <Route path="/enriquecimento" element={<Enriquecimento />} />
          <Route path="/conversao" element={<Conversao />} />
          <Route path="/admin" element={<Admin />} />
        </Route>

        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
