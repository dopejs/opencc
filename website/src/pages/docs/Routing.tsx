import { useTranslation } from "react-i18next";
import { CodeBlock } from "@/components/docs/CodeBlock";

const scenarios = [
  { name: "think", key: "think" },
  { name: "image", key: "image" },
  { name: "longContext", key: "longContext" },
  { name: "webSearch", key: "webSearch" },
  { name: "background", key: "background" },
];

const routingConfig = `{
  "profiles": {
    "smart": {
      "providers": ["main-api"],
      "long_context_threshold": 60000,
      "routing": {
        "think": {
          "providers": [{"name": "thinking-api", "model": "claude-opus-4-5"}]
        },
        "longContext": {
          "providers": [{"name": "long-context-api"}]
        }
      }
    }
  }
}`;

export default function Routing() {
  const { t } = useTranslation();

  return (
    <div>
      <h1 className="mb-4 text-3xl font-bold tracking-tight text-text-primary">
        {t("docs.routing.title")}
      </h1>
      <p className="mb-8 text-text-secondary">{t("docs.routing.intro")}</p>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.routing.scenariosTitle")}
        </h2>
        <div className="overflow-hidden rounded-xl border border-border bg-bg-surface">
          <div className="divide-y divide-border">
            {scenarios.map((s) => (
              <div key={s.name} className="grid grid-cols-[140px_1fr] px-5 py-3">
                <code className="text-sm text-teal">{s.name}</code>
                <span className="text-sm text-text-secondary">
                  {t(`docs.routing.scenarios.${s.key}`)}
                </span>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.routing.fallbackTitle")}
        </h2>
        <p className="mb-4 text-text-secondary">
          {t("docs.routing.fallbackDesc")}
        </p>
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.routing.configTitle")}
        </h2>
        <CodeBlock code={routingConfig} language="json" />
      </section>
    </div>
  );
}
