import { HashRouter, Routes, Route } from "react-router";
import { Header } from "@/components/layout/Header";
import { Footer } from "@/components/layout/Footer";
import { DocsLayout } from "@/components/layout/DocsLayout";
import Home from "@/pages/Home";
import DocsIndex from "@/pages/docs/DocsIndex";
import GettingStarted from "@/pages/docs/GettingStarted";
import Providers from "@/pages/docs/Providers";
import Profiles from "@/pages/docs/Profiles";
import Routing from "@/pages/docs/Routing";
import Bindings from "@/pages/docs/Bindings";
import MultiCli from "@/pages/docs/MultiCli";
import WebUI from "@/pages/docs/WebUI";
import TUI from "@/pages/docs/TUI";
import ConfigRef from "@/pages/docs/ConfigRef";

export default function App() {
  return (
    <HashRouter>
      <div className="flex min-h-screen flex-col">
        <Header />
        <main className="flex-1">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/docs" element={<DocsLayout />}>
              <Route index element={<DocsIndex />} />
              <Route path="getting-started" element={<GettingStarted />} />
              <Route path="providers" element={<Providers />} />
              <Route path="profiles" element={<Profiles />} />
              <Route path="routing" element={<Routing />} />
              <Route path="bindings" element={<Bindings />} />
              <Route path="multi-cli" element={<MultiCli />} />
              <Route path="web-ui" element={<WebUI />} />
              <Route path="tui" element={<TUI />} />
              <Route path="config" element={<ConfigRef />} />
            </Route>
          </Routes>
        </main>
        <Footer />
      </div>
    </HashRouter>
  );
}
