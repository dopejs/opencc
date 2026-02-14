import { useTranslation } from "react-i18next";
import {
  Terminal,
  Settings,
  Shield,
  GitBranch,
  FolderSymlink,
  Variable,
  Monitor,
  Globe,
  Sparkles,
} from "lucide-react";

const featureKeys = [
  { key: "multiCli", icon: Terminal },
  { key: "multiConfig", icon: Settings },
  { key: "failover", icon: Shield },
  { key: "routing", icon: GitBranch },
  { key: "binding", icon: FolderSymlink },
  { key: "envVars", icon: Variable },
  { key: "tui", icon: Monitor },
  { key: "webUi", icon: Globe },
];

export function Features() {
  const { t } = useTranslation();

  return (
    <section className="py-20">
      <div className="mx-auto max-w-6xl px-4 sm:px-6 lg:px-8">
        <h2 className="mb-12 text-center text-3xl font-bold tracking-tight text-text-primary">
          {t("features.title")}
        </h2>

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {featureKeys.map(({ key, icon: Icon }) => (
            <div
              key={key}
              className="group rounded-xl border border-border bg-bg-surface p-6 transition-all hover:border-border-strong hover:shadow-md"
            >
              <div className="mb-4 flex h-10 w-10 items-center justify-center rounded-lg bg-teal-dim text-teal">
                <Icon className="h-5 w-5" />
              </div>
              <h3 className="mb-2 text-base font-semibold text-text-primary">
                {t(`features.${key}.title`)}
              </h3>
              <p className="text-sm leading-relaxed text-text-secondary">
                {t(`features.${key}.desc`)}
              </p>
            </div>
          ))}

          {/* Coming soon card */}
          <div className="flex items-center justify-center rounded-xl border border-dashed border-border bg-bg-surface/50 p-6">
            <div className="text-center">
              <div className="mx-auto mb-3 flex h-10 w-10 items-center justify-center rounded-lg bg-lavender-dim text-lavender">
                <Sparkles className="h-5 w-5" />
              </div>
              <p className="text-sm font-medium text-text-muted">
                {t("features.comingSoon")}
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
