import { useTranslation } from "react-i18next";
import { CodeBlock } from "@/components/docs/CodeBlock";

const steps = [
  {
    key: "step1",
    code: "curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh",
  },
  {
    key: "step2",
    code: "opencc config",
  },
  {
    key: "step3",
    code: "opencc",
  },
];

export function Installation() {
  const { t } = useTranslation();

  return (
    <section className="py-20">
      <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">
        <h2 className="mb-12 text-center text-3xl font-bold tracking-tight text-text-primary">
          {t("install.title")}
        </h2>

        <div className="flex flex-col gap-8">
          {steps.map((step, i) => (
            <div key={step.key} className="flex gap-6">
              {/* Step number */}
              <div className="flex flex-col items-center">
                <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-teal text-base font-bold text-bg-base">
                  {i + 1}
                </div>
                {i < steps.length - 1 && (
                  <div className="my-2 w-px flex-1 bg-border" />
                )}
              </div>

              {/* Step content */}
              <div className="flex-1 pb-4">
                <h3 className="mb-1 text-lg font-semibold text-text-primary">
                  {t(`install.${step.key}.title`)}
                </h3>
                <p className="mb-3 text-sm text-text-secondary">
                  {t(`install.${step.key}.desc`)}
                </p>
                <CodeBlock code={step.code} language="bash" />
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
