import { useTranslation } from "react-i18next";
import { CodeBlock } from "@/components/docs/CodeBlock";

const files = [
  { path: "~/.opencc/opencc.json", key: "mainConfig" },
  { path: "~/.opencc/proxy.log", key: "proxyLog" },
  { path: "~/.opencc/web.log", key: "webLog" },
];

const fullConfig = `{
  "version": 5,
  "default_profile": "default",
  "default_cli": "claude",
  "web_port": 19840,
  "providers": {
    "anthropic": {
      "base_url": "https://api.anthropic.com",
      "auth_token": "sk-ant-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000"
      }
    }
  },
  "profiles": {
    "default": {
      "providers": ["anthropic"]
    }
  },
  "project_bindings": {
    "/path/to/project": {
      "profile": "work",
      "cli": "codex"
    }
  }
}`;

const fields = [
  { field: "version", key: "version" },
  { field: "default_profile", key: "defaultProfile" },
  { field: "default_cli", key: "defaultCli" },
  { field: "web_port", key: "webPort" },
  { field: "providers", key: "providers" },
  { field: "profiles", key: "profiles" },
  { field: "project_bindings", key: "projectBindings" },
];

export default function ConfigRef() {
  const { t } = useTranslation();

  return (
    <div>
      <h1 className="mb-4 text-3xl font-bold tracking-tight text-text-primary">
        {t("docs.configRef.title")}
      </h1>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.configRef.fileLocationTitle")}
        </h2>
        <div className="overflow-hidden rounded-xl border border-border bg-bg-surface">
          <div className="grid grid-cols-[1fr_1fr] border-b border-border bg-bg-elevated px-5 py-3 text-xs font-semibold uppercase tracking-wider text-text-muted">
            <span>{t("docs.configRef.filesTable.file")}</span>
            <span>{t("docs.configRef.filesTable.description")}</span>
          </div>
          <div className="divide-y divide-border">
            {files.map((f) => (
              <div key={f.path} className="grid grid-cols-[1fr_1fr] px-5 py-3">
                <code className="text-sm text-teal">{f.path}</code>
                <span className="text-sm text-text-secondary">
                  {t(`docs.configRef.filesTable.${f.key}`)}
                </span>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.configRef.fullExampleTitle")}
        </h2>
        <CodeBlock code={fullConfig} language="json" />
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.configRef.fieldsTitle")}
        </h2>
        <div className="overflow-hidden rounded-xl border border-border bg-bg-surface">
          <div className="divide-y divide-border">
            {fields.map((f) => (
              <div
                key={f.field}
                className="grid grid-cols-[180px_1fr] px-5 py-3"
              >
                <code className="text-sm text-teal">{f.field}</code>
                <span className="text-sm text-text-secondary">
                  {t(`docs.configRef.fields.${f.key}`)}
                </span>
              </div>
            ))}
          </div>
        </div>
      </section>
    </div>
  );
}
