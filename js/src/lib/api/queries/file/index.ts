import { useQuery } from "@tanstack/react-query";

import { getFile } from "../../fetchers/file";

export const useFileQuery = (
  githubUsername: string,
  githubRepo: string,
  filePath: string
) =>
  useQuery({
    queryKey: ["file", githubUsername, githubRepo, filePath],
    queryFn: () => getFile(githubUsername, githubRepo, filePath),
  });
