import { useState, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Check, Copy } from "lucide-react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { oneDark } from "react-syntax-highlighter/dist/esm/styles/prism";

interface CodeBlockProps {
  code: string;
  language?: string;
  showCopy?: boolean;
}

export function CodeBlock({
  code,
  language = "bash",
  showCopy = true,
}: CodeBlockProps) {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);

  // Strip background colors from all tokens, keep only text colors
  const cleanStyle = useMemo(() => {
    const style: Record<string, React.CSSProperties> = {};
    for (const [key, value] of Object.entries(oneDark)) {
      if (value && typeof value === "object") {
        const cleaned = { ...value } as Record<string, unknown>;
        delete cleaned.background;
        delete cleaned.backgroundColor;
        style[key] = cleaned as React.CSSProperties;
      }
    }
    return style;
  }, []);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="group relative rounded-lg border border-border bg-bg-base">
      {showCopy && (
        <button
          onClick={handleCopy}
          className="absolute right-2 top-2 flex h-8 w-8 items-center justify-center rounded-md bg-bg-overlay text-text-muted opacity-0 transition-all hover:text-text-primary group-hover:opacity-100"
          title={copied ? t("docs.copied") : t("hero.copyTip")}
        >
          {copied ? (
            <Check className="h-4 w-4 text-teal" />
          ) : (
            <Copy className="h-4 w-4" />
          )}
        </button>
      )}
      <SyntaxHighlighter
        language={language}
        style={cleanStyle}
        customStyle={{
          margin: 0,
          padding: "1rem",
          background: "transparent",
          fontSize: "0.875rem",
          lineHeight: "1.6",
        }}
      >
        {code}
      </SyntaxHighlighter>
    </div>
  );
}
