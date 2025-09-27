import { useQuery } from "@tanstack/react-query";

import { getFile } from "../../fetchers/file";

export const useFileQuery = (
  githubUsername: string,
  githubRepo: string,
  filePath?: string,
) =>
  useQuery({
    queryKey: ["file", githubUsername, githubRepo, filePath],
    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    queryFn: () => getFile(githubUsername, githubRepo, filePath!),
    enabled: !!filePath,
  });
