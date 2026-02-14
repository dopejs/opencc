import { useTranslation } from "react-i18next";
import { CodeBlock } from "@/components/docs/CodeBlock";

export default function GettingStarted() {
  const { t } = useTranslation();

  return (
    <div className="prose-custom">
      <h1 className="mb-8 text-3xl font-bold tracking-tight text-text-primary">
        {t("docs.gettingStarted.title")}
      </h1>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.gettingStarted.installTitle")}
        </h2>
        <p className="mb-3 text-text-secondary">
          {t("docs.gettingStarted.installDesc")}
        </p>
        <CodeBlock code="curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh" />

        <p className="mb-3 mt-6 text-text-secondary">
          {t("docs.gettingStarted.uninstallDesc")}
        </p>
        <CodeBlock code="curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh -s -- --uninstall" />
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.gettingStarted.firstRunTitle")}
        </h2>

        <p className="mb-3 text-text-secondary">
          {t("docs.gettingStarted.firstRunStep1")}
        </p>
        <CodeBlock code="opencc config" />

        <p className="mb-3 mt-6 text-text-secondary">
          {t("docs.gettingStarted.firstRunStep2")}
        </p>
        <CodeBlock code="opencc" />

        <p className="mb-3 mt-6 text-text-secondary">
          {t("docs.gettingStarted.firstRunStep3")}
        </p>
        <CodeBlock code="opencc -p work" />

        <p className="mb-3 mt-6 text-text-secondary">
          {t("docs.gettingStarted.firstRunStep4")}
        </p>
        <CodeBlock code="opencc --cli codex" />
      </section>
    </div>
  );
}
