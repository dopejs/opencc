import { useTranslation } from "react-i18next";
import { CodeBlock } from "@/components/docs/CodeBlock";

export default function MultiCli() {
  const { t } = useTranslation();

  return (
    <div>
      <h1 className="mb-4 text-3xl font-bold tracking-tight text-text-primary">
        {t("docs.multiCli.title")}
      </h1>
      <p className="mb-8 text-text-secondary">{t("docs.multiCli.intro")}</p>

      {/* CLI Table */}
      <section className="mb-10">
        <div className="overflow-hidden rounded-xl border border-border bg-bg-surface">
          <div className="grid grid-cols-3 border-b border-border bg-bg-elevated px-5 py-3 text-xs font-semibold uppercase tracking-wider text-text-muted">
            <span>{t("docs.multiCli.cliTable.cli")}</span>
            <span>{t("docs.multiCli.cliTable.description")}</span>
            <span>{t("docs.multiCli.cliTable.format")}</span>
          </div>
          <div className="divide-y divide-border">
            {[
              { cli: "claude", desc: "claude", format: "claudeFormat" },
              { cli: "codex", desc: "codex", format: "codexFormat" },
              { cli: "opencode", desc: "opencode", format: "opencodeFormat" },
            ].map((row) => (
              <div key={row.cli} className="grid grid-cols-3 px-5 py-3">
                <code className="text-sm text-teal">{row.cli}</code>
                <span className="text-sm text-text-secondary">
                  {t(`docs.multiCli.cliTable.${row.desc}`)}
                </span>
                <span className="text-sm text-text-muted">
                  {t(`docs.multiCli.cliTable.${row.format}`)}
                </span>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.multiCli.defaultTitle")}
        </h2>
        <CodeBlock
          code={`# Via TUI
opencc config  # Settings → Default CLI

# Via Web UI
opencc web open  # Settings page`}
          language="bash"
        />
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.multiCli.projectTitle")}
        </h2>
        <CodeBlock
          code={`cd ~/work/project
opencc bind --cli codex  # This directory uses Codex`}
          language="bash"
        />
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.multiCli.tempTitle")}
        </h2>
        <CodeBlock
          code="opencc --cli opencode  # Use OpenCode for this session"
          language="bash"
        />
      </section>
    </div>
  );
}
