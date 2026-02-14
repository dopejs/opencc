import { useTranslation } from "react-i18next";
import { Link } from "react-router";
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
  ArrowRight,
} from "lucide-react";

const sections = [
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

export default function DocsIndex() {
  const { t } = useTranslation();

  return (
    <div>
      <h1 className="mb-3 text-3xl font-bold tracking-tight text-text-primary">
        {t("docs.index.title")}
      </h1>
      <p className="mb-8 text-text-secondary">{t("docs.index.subtitle")}</p>

      <div className="grid gap-3 sm:grid-cols-2">
        {sections.map((s) => {
          const Icon = s.icon;
          return (
            <Link
              key={s.path}
              to={s.path}
              className="group flex items-center gap-4 rounded-xl border border-border bg-bg-surface p-4 transition-all no-underline hover:border-border-strong hover:shadow-sm"
            >
              <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-teal-dim text-teal">
                <Icon className="h-5 w-5" />
              </div>
              <span className="flex-1 text-sm font-medium text-text-primary">
                {t(`docs.sidebar.${s.key}`)}
              </span>
              <ArrowRight className="h-4 w-4 text-text-muted transition-transform group-hover:translate-x-0.5 group-hover:text-teal" />
            </Link>
          );
        })}
      </div>
    </div>
  );
}
