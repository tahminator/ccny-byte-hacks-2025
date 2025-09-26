import RootPage from "@/app/Root.page";
import { createBrowserRouter } from "react-router";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <RootPage />,
  },
]);
