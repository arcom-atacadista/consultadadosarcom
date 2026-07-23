import { Outlet } from "react-router-dom";
import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";

export function AppShell() {
  return (
    <div className="flex min-h-screen bg-surface">
      <div className="no-print contents">
        <Sidebar />
      </div>
      <div className="flex flex-1 flex-col">
        <div className="no-print contents">
          <Topbar />
        </div>
        <main className="flex-1 p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
