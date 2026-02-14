import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { useState } from "react";
import { Check, Copy, ArrowRight } from "lucide-react";

const installCmd =
  'curl -fsSL https://raw.githubusercontent.com/dopejs/opencc/main/install.sh | sh';

export function Hero() {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(installCmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <section className="relative overflow-hidden py-20 sm:py-32">
      {/* Background gradient */}
      <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
        <div className="h-[500px] w-[800px] rounded-full bg-teal/5 blur-[120px]" />
      </div>

      <div className="relative mx-auto max-w-4xl px-4 text-center sm:px-6 lg:px-8">
        <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-border bg-bg-surface px-4 py-1.5 text-sm text-text-secondary">
          <span className="inline-block h-2 w-2 rounded-full bg-teal" />
          Open Source CLI Tool
        </div>

        <h1 className="mb-6 text-4xl font-bold tracking-tight text-text-primary sm:text-5xl lg:text-6xl">
          <span className="text-teal">OpenCC</span>
          <span className="mt-3 block">{t("hero.title")}</span>
        </h1>

        <p className="mx-auto mb-10 max-w-2xl text-lg text-text-secondary">
          {t("hero.subtitle")}
        </p>

        {/* Install command */}
        <div className="mx-auto mb-8 max-w-3xl">
          <div
            onClick={handleCopy}
            className="group flex cursor-pointer items-center gap-3 overflow-x-auto rounded-xl border border-border bg-bg-surface px-5 py-4 transition-all hover:border-border-strong hover:shadow-lg"
          >
            <span className="flex-shrink-0 text-text-muted">$</span>
            <code className="flex-1 whitespace-nowrap text-left text-sm text-text-primary">
              {installCmd}
            </code>
            <span className="flex-shrink-0 text-text-muted transition-colors group-hover:text-text-primary">
              {copied ? (
                <Check className="h-4 w-4 text-teal" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
            </span>
          </div>
        </div>

        {/* CTA buttons */}
        <div className="flex flex-wrap items-center justify-center gap-4">
          <Link
            to="/docs/getting-started"
            className="inline-flex items-center gap-2 rounded-xl bg-teal px-6 py-3 text-sm font-semibold text-bg-base shadow-[0_1px_3px_rgba(94,234,212,0.2)] transition-all no-underline hover:shadow-[0_2px_8px_rgba(94,234,212,0.3)]"
          >
            {t("hero.getDocs")}
            <ArrowRight className="h-4 w-4" />
          </Link>
          <a
            href="https://github.com/dopejs/opencc"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 rounded-xl border border-border bg-bg-surface px-6 py-3 text-sm font-semibold text-text-primary transition-all no-underline hover:border-border-strong hover:bg-bg-elevated"
          >
            GitHub
          </a>
        </div>
      </div>
    </section>
  );
}
