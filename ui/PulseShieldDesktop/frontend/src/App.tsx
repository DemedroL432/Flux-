import React from "react";
import { AppProvider, useApp } from "./context/AppContext";
import { BusyLine, Sidebar, Toasts, TopBar } from "./ui";
import { ShieldScreen } from "./pages/ShieldScreen";
import { ServersScreen } from "./pages/ServersScreen";
import { RoutesScreen } from "./pages/RoutesScreen";
import { DpiScreen } from "./pages/DpiScreen";
import { NetworkLogsScreen } from "./pages/NetworkLogsScreen";

function AppShell() {
  const { active } = useApp();

  return (
    <div className="fx-win">
      <Sidebar />
      <main className="fx-main">
        <TopBar />
        <BusyLine />
        <div className="fx-canvas">
          {active === "shield" && <ShieldScreen />}
          {active === "servers" && <ServersScreen />}
          {active === "routes" && <RoutesScreen />}
          {active === "dpi" && <DpiScreen />}
          {active === "net" && <NetworkLogsScreen />}
        </div>
      </main>
      <Toasts />
    </div>
  );
}

export default function App() {
  return (
    <AppProvider>
      <AppShell />
    </AppProvider>
  );
}
