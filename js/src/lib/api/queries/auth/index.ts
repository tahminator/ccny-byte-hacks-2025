import { useQuery } from "@tanstack/react-query";

import { getAuth } from "@/lib/api/fetchers/auth";

import { getFileTree } from "../../fetchers/file";

export const useAuthQuery = () =>
  useQuery({
    queryKey: ["auth"],
    queryFn: getAuth,
  });

export const useFileTreeQuery = (repoName: string) =>
  useQuery({
    queryKey: ["tree"],
    queryFn: () => getFileTree(repoName),
      refetchOnWindowFocus: false,
  });
