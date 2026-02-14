import { useTranslation } from "react-i18next";
import { CodeBlock } from "@/components/docs/CodeBlock";

const providerConfig = `{
  "providers": {
    "my-provider": {
      "base_url": "https://api.example.com",
      "auth_token": "sk-xxx",
      "model": "claude-sonnet-4-5",
      "claude_env_vars": {
        "CLAUDE_CODE_MAX_OUTPUT_TOKENS": "64000",
        "MAX_THINKING_TOKENS": "50000"
      },
      "codex_env_vars": {
        "CODEX_SOME_VAR": "value"
      },
      "opencode_env_vars": {
        "OPENCODE_EXPERIMENTAL_OUTPUT_TOKEN_MAX": "64000"
      }
    }
  }
}`;

const envVars = [
  { name: "CLAUDE_CODE_MAX_OUTPUT_TOKENS", key: "envMaxOutput" },
  { name: "MAX_THINKING_TOKENS", key: "envMaxThinking" },
  { name: "ANTHROPIC_MAX_CONTEXT_WINDOW", key: "envMaxContext" },
  { name: "BASH_DEFAULT_TIMEOUT_MS", key: "envBashTimeout" },
];

export default function Providers() {
  const { t } = useTranslation();

  return (
    <div>
      <h1 className="mb-4 text-3xl font-bold tracking-tight text-text-primary">
        {t("docs.providers.title")}
      </h1>
      <p className="mb-8 text-text-secondary">{t("docs.providers.intro")}</p>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.providers.configTitle")}
        </h2>
        <CodeBlock code={providerConfig} language="json" />
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.providers.envVarsTitle")}
        </h2>
        <p className="mb-4 text-text-secondary">
          {t("docs.providers.envVarsDesc")}
        </p>

        <h3 className="mb-3 text-lg font-semibold text-text-primary">
          {t("docs.providers.claudeEnvTitle")}
        </h3>
        <div className="overflow-hidden rounded-xl border border-border bg-bg-surface">
          <div className="divide-y divide-border">
            {envVars.map((v) => (
              <div
                key={v.name}
                className="grid grid-cols-[1fr_1fr] px-5 py-3"
              >
                <code className="text-sm text-teal">{v.name}</code>
                <span className="text-sm text-text-secondary">
                  {t(`docs.providers.${v.key}`)}
                </span>
              </div>
            ))}
          </div>
        </div>
      </section>
    </div>
  );
}
