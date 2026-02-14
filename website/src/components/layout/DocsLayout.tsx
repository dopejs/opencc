import { Outlet, useLocation } from "react-router";
import { DocsSidebar } from "@/components/docs/DocsSidebar";
import { Menu, X } from "lucide-react";
import { useState, useEffect } from "react";

export function DocsLayout() {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const location = useLocation();

  useEffect(() => {
    setSidebarOpen(false);
  }, [location.pathname]);

  return (
    <div className="mx-auto max-w-6xl px-4 sm:px-6 lg:px-8">
      <div className="flex gap-8 py-8">
        {/* Mobile sidebar toggle */}
        <button
          onClick={() => setSidebarOpen(!sidebarOpen)}
          className="fixed bottom-4 right-4 z-40 flex h-12 w-12 items-center justify-center rounded-full bg-teal text-bg-base shadow-lg lg:hidden"
        >
          {sidebarOpen ? (
            <X className="h-5 w-5" />
          ) : (
            <Menu className="h-5 w-5" />
          )}
        </button>

        {/* Sidebar */}
        <aside
          className={`${
            sidebarOpen
              ? "fixed inset-0 z-30 block bg-bg-base/80 backdrop-blur-sm lg:static lg:bg-transparent"
              : "hidden lg:block"
          }`}
          onClick={(e) => {
            if (e.target === e.currentTarget) setSidebarOpen(false);
          }}
        >
          <div className="fixed left-0 top-16 h-[calc(100vh-4rem)] w-64 overflow-y-auto border-r border-border bg-bg-base p-4 lg:sticky lg:border-r-0 lg:bg-transparent lg:p-0">
            <DocsSidebar />
          </div>
        </aside>

        {/* Content */}
        <main className="min-w-0 flex-1">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
