import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider } from "@/hooks/useAuth";
import { ExigeLogin, ExigeAprovado, ExigeAdmin } from "@/components/RotaProtegida";
import { AppShell } from "@/components/layout/AppShell";
import Cover from "@/pages/Cover";
import Login from "@/pages/Login";
import Registrar from "@/pages/Registrar";
import Pendente from "@/pages/Pendente";
import Consulta from "@/pages/Consulta";
import Prospeccao from "@/pages/Prospeccao";
import Enriquecimento from "@/pages/Enriquecimento";
import Conversao from "@/pages/Conversao";
import Admin from "@/pages/Admin";

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/" element={<Cover />} />
          <Route path="/login" element={<Login />} />
          <Route path="/registrar" element={<Registrar />} />

          <Route element={<ExigeLogin />}>
            <Route path="/pendente" element={<Pendente />} />

            <Route element={<ExigeAprovado />}>
              <Route element={<AppShell />}>
                <Route path="/consulta" element={<Consulta />} />
                <Route path="/prospeccao" element={<Prospeccao />} />
                <Route path="/enriquecimento" element={<Enriquecimento />} />
                <Route path="/conversao" element={<Conversao />} />

                <Route element={<ExigeAdmin />}>
                  <Route path="/admin" element={<Admin />} />
                </Route>
              </Route>
            </Route>
          </Route>

          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
