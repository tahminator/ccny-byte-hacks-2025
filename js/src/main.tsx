import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { StrictMode } from "react";

import "./main.css";
import { createRoot } from "react-dom/client";
import { RouterProvider } from "react-router";
import { Toaster } from "sonner";

import ReactQueryProvider from "@/lib/query";
import { router } from "@/lib/router";
import { ThemeProvider } from "@/lib/themeProvider";

// eslint-disable-next-line @typescript-eslint/no-non-null-assertion
createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <ReactQueryProvider>
        <Toaster />
      <RouterProvider router={router} />
        <ReactQueryDevtools />
      </ReactQueryProvider>
    </ThemeProvider>
  </StrictMode>
);
