import { useTranslation } from "react-i18next";
import { CodeBlock } from "@/components/docs/CodeBlock";

const profileConfig = `{
  "profiles": {
    "default": {
      "providers": ["anthropic-main", "anthropic-backup"]
    },
    "work": {
      "providers": ["company-api"],
      "routing": {
        "think": {"providers": [{"name": "thinking-api"}]}
      }
    }
  }
}`;

const usageCode = `# Use default profile
opencc

# Use specified profile
opencc -p work

# Interactively select
opencc -p`;

export default function Profiles() {
  const { t } = useTranslation();

  return (
    <div>
      <h1 className="mb-4 text-3xl font-bold tracking-tight text-text-primary">
        {t("docs.profiles.title")}
      </h1>
      <p className="mb-8 text-text-secondary">{t("docs.profiles.intro")}</p>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.profiles.configTitle")}
        </h2>
        <CodeBlock code={profileConfig} language="json" />
      </section>

      <section className="mb-10">
        <h2 className="mb-4 text-xl font-semibold text-text-primary">
          {t("docs.profiles.usageTitle")}
        </h2>
        <CodeBlock code={usageCode} language="bash" />
      </section>
    </div>
  );
}
