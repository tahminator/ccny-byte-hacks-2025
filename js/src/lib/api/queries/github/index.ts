import { useMutation, type UseMutationOptions } from "@tanstack/react-query";

import { commitRepository } from "../../fetchers/github";

export const useCommitRepositoryMutation = (
  options?: UseMutationOptions<
    unknown,
    Error,
    { repoName: string; newFileData: string; path: string }
  >
) => {
  return useMutation({
    mutationFn: ({
      repoName,
      newFileData,
      path,
    }: {
      repoName: string;
      newFileData: string;
      path: string;
    }) => commitRepository(repoName, newFileData, path),
    ...options,
  });
};
