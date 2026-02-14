import { useTranslation } from "react-i18next";
import { Link, useLocation } from "react-router";
import { Menu, X, Sun, Moon, ChevronDown } from "lucide-react";
import { useState, useEffect, useRef } from "react";

const languages = [
  { code: "en", label: "English" },
  { code: "zh", label: "简体中文" },
  { code: "zh-TW", label: "繁體中文" },
  { code: "es", label: "Español" },
];

export function Header() {
  const { t, i18n } = useTranslation();
  const location = useLocation();
  const [dark, setDark] = useState(true);
  const [menuOpen, setMenuOpen] = useState(false);
  const [langOpen, setLangOpen] = useState(false);
  const langRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const saved = localStorage.getItem("opencc-theme");
    if (saved === "light") {
      setDark(false);
      document.documentElement.classList.remove("dark");
    } else {
      setDark(true);
      document.documentElement.classList.add("dark");
    }
  }, []);

  // Close lang dropdown on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (langRef.current && !langRef.current.contains(e.target as Node)) {
        setLangOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  const toggleTheme = () => {
    const next = !dark;
    setDark(next);
    if (next) {
      document.documentElement.classList.add("dark");
      localStorage.setItem("opencc-theme", "dark");
    } else {
      document.documentElement.classList.remove("dark");
      localStorage.setItem("opencc-theme", "light");
    }
  };

  const currentLang =
    languages.find((l) => l.code === i18n.language) ?? languages[0];

  const isActive = (path: string) =>
    location.pathname === path || location.pathname.startsWith(path + "/");

  return (
    <header className="sticky top-0 z-50 border-b border-border bg-bg-base/80 backdrop-blur-md">
      <div className="mx-auto max-w-6xl px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
          {/* Logo */}
          <Link to="/" className="flex items-center gap-2.5 no-underline">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-teal-dim">
              <span className="text-lg font-bold text-teal">⌘</span>
            </div>
            <span className="text-lg font-bold tracking-tight text-text-primary">
              OpenCC
            </span>
          </Link>

          {/* Desktop nav */}
          <nav className="hidden items-center gap-1 md:flex">
            <Link
              to="/"
              className={`rounded-lg px-3 py-2 text-sm font-medium transition-colors no-underline ${
                location.pathname === "/"
                  ? "text-teal"
                  : "text-text-secondary hover:text-text-primary"
              }`}
            >
              {t("nav.home")}
            </Link>
            <Link
              to="/docs"
              className={`rounded-lg px-3 py-2 text-sm font-medium transition-colors no-underline ${
                isActive("/docs")
                  ? "text-teal"
                  : "text-text-secondary hover:text-text-primary"
              }`}
            >
              {t("nav.docs")}
            </Link>
            <a
              href="https://github.com/dopejs/opencc"
              target="_blank"
              rel="noopener noreferrer"
              className="rounded-lg px-3 py-2 text-sm font-medium text-text-secondary transition-colors no-underline hover:text-text-primary"
            >
              {t("nav.github")}
            </a>
          </nav>

          {/* Actions */}
          <div className="flex items-center gap-1">
            {/* Language dropdown */}
            <div className="relative" ref={langRef}>
              <button
                onClick={() => setLangOpen(!langOpen)}
                className="flex h-9 items-center gap-1 rounded-lg px-2.5 text-sm font-medium text-text-muted transition-colors hover:bg-bg-overlay hover:text-text-primary"
              >
                <span className="hidden sm:inline">{currentLang.label}</span>
                <span className="sm:hidden">{currentLang.code.toUpperCase()}</span>
                <ChevronDown className={`h-3.5 w-3.5 transition-transform ${langOpen ? "rotate-180" : ""}`} />
              </button>
              {langOpen && (
                <div className="absolute right-0 top-full mt-1 w-40 overflow-hidden rounded-lg border border-border bg-bg-surface shadow-lg">
                  {languages.map((lang) => (
                    <button
                      key={lang.code}
                      onClick={() => {
                        i18n.changeLanguage(lang.code);
                        setLangOpen(false);
                      }}
                      className={`flex w-full items-center px-3 py-2 text-left text-sm transition-colors ${
                        i18n.language === lang.code
                          ? "bg-teal-dim text-teal"
                          : "text-text-secondary hover:bg-bg-overlay hover:text-text-primary"
                      }`}
                    >
                      {lang.label}
                    </button>
                  ))}
                </div>
              )}
            </div>

            <button
              onClick={toggleTheme}
              className="flex h-9 w-9 items-center justify-center rounded-lg text-text-muted transition-colors hover:bg-bg-overlay hover:text-text-primary"
            >
              {dark ? (
                <Sun className="h-4 w-4" />
              ) : (
                <Moon className="h-4 w-4" />
              )}
            </button>
            <button
              onClick={() => setMenuOpen(!menuOpen)}
              className="flex h-9 w-9 items-center justify-center rounded-lg text-text-muted transition-colors hover:bg-bg-overlay hover:text-text-primary md:hidden"
            >
              {menuOpen ? (
                <X className="h-5 w-5" />
              ) : (
                <Menu className="h-5 w-5" />
              )}
            </button>
          </div>
        </div>
      </div>

      {/* Mobile menu */}
      {menuOpen && (
        <div className="border-t border-border px-4 py-3 md:hidden">
          <nav className="flex flex-col gap-1">
            <Link
              to="/"
              onClick={() => setMenuOpen(false)}
              className={`rounded-lg px-3 py-2 text-sm font-medium no-underline ${
                location.pathname === "/"
                  ? "text-teal"
                  : "text-text-secondary"
              }`}
            >
              {t("nav.home")}
            </Link>
            <Link
              to="/docs"
              onClick={() => setMenuOpen(false)}
              className={`rounded-lg px-3 py-2 text-sm font-medium no-underline ${
                isActive("/docs")
                  ? "text-teal"
                  : "text-text-secondary"
              }`}
            >
              {t("nav.docs")}
            </Link>
            <a
              href="https://github.com/dopejs/opencc"
              target="_blank"
              rel="noopener noreferrer"
              className="rounded-lg px-3 py-2 text-sm font-medium text-text-secondary no-underline"
            >
              {t("nav.github")}
            </a>
          </nav>
        </div>
      )}
    </header>
  );
}
