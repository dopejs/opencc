import { useTranslation } from "react-i18next";
import { useVersion } from "@/hooks/use-version";

export function Footer() {
  const { t } = useTranslation();
  const version = useVersion();

  return (
    <footer className="border-t border-border bg-bg-surface">
      <div className="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="flex flex-col items-center justify-between gap-4 sm:flex-row">
          <div className="flex items-center gap-4 text-sm text-text-muted">
            <span>{t("footer.license")}</span>
            <span>·</span>
            <span>{t("footer.builtWith")}</span>
            {version && (
              <>
                <span>·</span>
                <span>{version}</span>
              </>
            )}
          </div>
          <div className="flex items-center gap-4">
            <a
              href="https://github.com/dopejs/opencc"
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm text-text-muted transition-colors no-underline hover:text-text-primary"
            >
              GitHub
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
}
