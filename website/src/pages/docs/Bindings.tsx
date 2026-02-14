import { useTranslation } from "react-i18next";
import { CodeBlock } from "@/components/docs/CodeBlock";

const bindingCode = `cd ~/work/company-project

# Bind profile
opencc bind work-profile

# Bind CLI
opencc bind --cli codex

# Bind both
opencc bind work-profile --cli codex

# Check status
opencc status

# Unbind
opencc unbind`;

export default function Bindings() {
  const { t } = useTranslation();

  return (
    <div>
      <h1 className="mb-4 text-3xl font-bold tracking-tight text-text-primary">
        {t("docs.bindings.title")}
      </h1>
      <p className="mb-8 text-text-secondary">{t("docs.bindings.intro")}</p>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.bindings.usageTitle")}
        </h2>
        <CodeBlock code={bindingCode} language="bash" />
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.bindings.priorityTitle")}
        </h2>
        <div className="rounded-xl border border-border bg-bg-surface p-5">
          <p className="text-sm font-medium text-text-primary">
            {t("docs.bindings.priorityDesc")}
          </p>
        </div>
      </section>
    </div>
  );
}
