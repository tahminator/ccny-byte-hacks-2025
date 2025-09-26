import { useQuery } from "@tanstack/react-query";

import { getAuth } from "@/lib/api/fetchers/auth";

export const useAuthQuery = () =>
  useQuery({
    queryKey: ["auth"],
    queryFn: getAuth,
  });
