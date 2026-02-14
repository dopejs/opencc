import { useTranslation } from "react-i18next";
import { CodeBlock } from "@/components/docs/CodeBlock";

const shortcuts = [
  { key: "a", action: "a" },
  { key: "e", action: "e" },
  { key: "d", action: "d" },
  { key: "Tab", action: "tab" },
  { key: "q", action: "q" },
];

export default function TUI() {
  const { t } = useTranslation();

  return (
    <div>
      <h1 className="mb-4 text-3xl font-bold tracking-tight text-text-primary">
        {t("docs.tui.title")}
      </h1>
      <p className="mb-8 text-text-secondary">{t("docs.tui.intro")}</p>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.tui.dashboardTitle")}
        </h2>
        <p className="mb-4 text-text-secondary">
          {t("docs.tui.dashboardDesc")}
        </p>
        <ul className="mb-6 space-y-2">
          {(["left", "right"] as const).map((key) => (
            <li
              key={key}
              className="flex items-start gap-2 text-sm text-text-secondary"
            >
              <span className="mt-1.5 inline-block h-1.5 w-1.5 flex-shrink-0 rounded-full bg-teal" />
              {t(`docs.tui.dashboardFeatures.${key}`)}
            </li>
          ))}
        </ul>
        <CodeBlock code="opencc config" language="bash" />
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.tui.shortcutsTitle")}
        </h2>
        <div className="overflow-hidden rounded-xl border border-border bg-bg-surface">
          <div className="divide-y divide-border">
            {shortcuts.map((s) => (
              <div
                key={s.key}
                className="grid grid-cols-[80px_1fr] px-5 py-3"
              >
                <kbd className="inline-flex h-6 w-fit items-center rounded border border-border bg-bg-elevated px-2 text-xs font-mono text-text-primary">
                  {s.key}
                </kbd>
                <span className="text-sm text-text-secondary">
                  {t(`docs.tui.shortcuts.${s.action}`)}
                </span>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.tui.legacyTitle")}
        </h2>
        <CodeBlock code="opencc config --legacy" language="bash" />
      </section>
    </div>
  );
}
