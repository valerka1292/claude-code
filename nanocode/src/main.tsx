/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App.tsx";
import { ProvidersProvider } from "./contexts/ProvidersContext.tsx";
import { ProjectProvider } from "./contexts/ProjectContext.tsx";
import { SessionProvider } from "./contexts/SessionContext.tsx";
import "./index.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <ProvidersProvider>
      <ProjectProvider>
        <SessionProvider>
          <App />
        </SessionProvider>
      </ProjectProvider>
    </ProvidersProvider>
  </StrictMode>
);