import { useTranslation } from "react-i18next";
import { Link, useLocation } from "react-router";
import {
  Rocket,
  Server,
  Shield,
  GitBranch,
  FolderSymlink,
  Terminal,
  Globe,
  Monitor,
  FileJson,
} from "lucide-react";

const navItems = [
  { path: "/docs/getting-started", key: "gettingStarted", icon: Rocket },
  { path: "/docs/providers", key: "providers", icon: Server },
  { path: "/docs/profiles", key: "profiles", icon: Shield },
  { path: "/docs/routing", key: "routing", icon: GitBranch },
  { path: "/docs/bindings", key: "bindings", icon: FolderSymlink },
  { path: "/docs/multi-cli", key: "multiCli", icon: Terminal },
  { path: "/docs/web-ui", key: "webUi", icon: Globe },
  { path: "/docs/tui", key: "tui", icon: Monitor },
  { path: "/docs/config", key: "config", icon: FileJson },
];

export function DocsSidebar() {
  const { t } = useTranslation();
  const location = useLocation();

  return (
    <nav className="flex flex-col gap-1">
      {navItems.map((item) => {
        const active = location.pathname === item.path;
        const Icon = item.icon;
        return (
          <Link
            key={item.path}
            to={item.path}
            className={`flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm font-medium transition-colors no-underline ${
              active
                ? "bg-teal-dim text-teal"
                : "text-text-secondary hover:bg-bg-overlay hover:text-text-primary"
            }`}
          >
            <Icon className="h-4 w-4 flex-shrink-0" />
            {t(`docs.sidebar.${item.key}`)}
          </Link>
        );
      })}
    </nav>
  );
}
