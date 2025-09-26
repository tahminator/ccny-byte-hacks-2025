import ReactQueryProvider from "@/lib/query";

import "./main.css";

import { router } from "@/lib/router";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider } from "react-router";

// eslint-disable-next-line @typescript-eslint/no-non-null-assertion
createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <ReactQueryProvider>
      <RouterProvider router={router} />
      <ReactQueryDevtools />
    </ReactQueryProvider>
  </StrictMode>,
);
