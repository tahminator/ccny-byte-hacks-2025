import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { StrictMode } from "react";

import "./main.css";
import { createRoot } from "react-dom/client";
import { RouterProvider } from "react-router";

import ReactQueryProvider from "@/lib/query";
import { router } from "@/lib/router";

// eslint-disable-next-line @typescript-eslint/no-non-null-assertion
createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <ReactQueryProvider>
      <RouterProvider router={router} />
      <ReactQueryDevtools />
    </ReactQueryProvider>
  </StrictMode>,
);
