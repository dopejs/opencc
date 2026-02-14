import { useTranslation } from "react-i18next";

const commands = [
  { cmd: "opencc", key: "start" },
  { cmd: "opencc -p <profile>", key: "profile" },
  { cmd: "opencc -p", key: "profilePick" },
  { cmd: "opencc --cli <cli>", key: "cli" },
  { cmd: "opencc use <provider>", key: "use" },
  { cmd: "opencc pick", key: "pick" },
  { cmd: "opencc list", key: "list" },
  { cmd: "opencc config", key: "config" },
  { cmd: "opencc config --legacy", key: "configLegacy" },
  { cmd: "opencc bind <profile>", key: "bind" },
  { cmd: "opencc bind --cli <cli>", key: "bindCli" },
  { cmd: "opencc unbind", key: "unbind" },
  { cmd: "opencc status", key: "status" },
  { cmd: "opencc web start", key: "webStart" },
  { cmd: "opencc web open", key: "webOpen" },
  { cmd: "opencc web stop", key: "webStop" },
  { cmd: "opencc upgrade", key: "upgrade" },
  { cmd: "opencc version", key: "version" },
];

export function Commands() {
  const { t } = useTranslation();

  return (
    <section className="py-20">
      <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">
        <h2 className="mb-12 text-center text-3xl font-bold tracking-tight text-text-primary">
          {t("commands.title")}
        </h2>

        <div className="overflow-hidden rounded-xl border border-border bg-bg-surface">
          <div className="grid grid-cols-[minmax(200px,1fr)_2fr] border-b border-border bg-bg-elevated px-5 py-3 text-xs font-semibold uppercase tracking-wider text-text-muted">
            <span>{t("commands.command")}</span>
            <span>{t("commands.description")}</span>
          </div>
          <div className="divide-y divide-border">
            {commands.map((item) => (
              <div
                key={item.key}
                className="grid grid-cols-[minmax(200px,1fr)_2fr] px-5 py-3 transition-colors hover:bg-bg-elevated/50"
              >
                <code className="text-sm text-teal">{item.cmd}</code>
                <span className="text-sm text-text-secondary">
                  {t(`commands.items.${item.key}`)}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
