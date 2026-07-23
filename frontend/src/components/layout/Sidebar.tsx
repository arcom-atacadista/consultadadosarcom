import { NavLink } from "react-router-dom";
import { Search, MapPinned, Layers, RefreshCcw, ShieldCheck } from "lucide-react";
import { cn } from "@/lib/cn";
import { useAuth } from "@/hooks/useAuth";

const NAV_ITEMS = [
  { to: "/consulta", label: "Consulta", icon: Search, adminOnly: false },
  { to: "/prospeccao", label: "Prospecção", icon: MapPinned, adminOnly: false },
  { to: "/enriquecimento", label: "Enriquecimento", icon: Layers, adminOnly: false },
  { to: "/conversao", label: "Conversão", icon: RefreshCcw, adminOnly: true },
  { to: "/admin", label: "Admin", icon: ShieldCheck, adminOnly: true },
] as const;

export function Sidebar() {
  const { isAdmin } = useAuth();
  return (
    <aside className="hidden md:flex w-60 shrink-0 flex-col bg-verde-escuro text-white">
      <div className="px-6 py-6">
        <span className="font-black text-2xl tracking-tight">ARCOM</span>
        <p className="text-xs text-white/60 mt-0.5">Consulta Dados Arcom</p>
      </div>
      <nav className="flex-1 px-2">
        {NAV_ITEMS.filter((item) => !item.adminOnly || isAdmin).map(({ to, label, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            className={({ isActive }) =>
              cn(
                "flex items-center gap-3 rounded-md px-4 py-2.5 mb-1 text-sm font-medium border-l-[3px] border-transparent text-white/70 transition-colors duration-fast hover:bg-white/5 hover:text-white",
                isActive && "border-verde-lima text-verde-lima bg-white/5",
              )
            }
          >
            <Icon className="h-5 w-5" strokeWidth={2} />
            {label}
          </NavLink>
        ))}
      </nav>
    </aside>
  );
}
