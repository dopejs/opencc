import { useTranslation } from "react-i18next";
import { CodeBlock } from "@/components/docs/CodeBlock";

export default function WebUI() {
  const { t } = useTranslation();

  return (
    <div>
      <h1 className="mb-4 text-3xl font-bold tracking-tight text-text-primary">
        {t("docs.webUi.title")}
      </h1>
      <p className="mb-8 text-text-secondary">{t("docs.webUi.intro")}</p>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.webUi.usageTitle")}
        </h2>
        <CodeBlock
          code={`# Start (runs in background, port 19840)
opencc web start

# Open browser
opencc web open

# Stop
opencc web stop`}
          language="bash"
        />
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.webUi.featuresTitle")}
        </h2>
        <ul className="space-y-2">
          {(
            [
              "providerManage",
              "bindingManage",
              "settings",
              "logs",
              "autocomplete",
            ] as const
          ).map((key) => (
            <li
              key={key}
              className="flex items-start gap-2 text-sm text-text-secondary"
            >
              <span className="mt-1.5 inline-block h-1.5 w-1.5 flex-shrink-0 rounded-full bg-teal" />
              {t(`docs.webUi.features.${key}`)}
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}
