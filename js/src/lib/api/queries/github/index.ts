import { useMutation, type UseMutationOptions } from "@tanstack/react-query";

import { commitRepository } from "../../fetchers/github";

export const useCommitRepositoryMutation = (
  options?: UseMutationOptions<unknown, Error, { repoName: string }>
) => {
  return useMutation({
    mutationFn: ({ repoName }: { repoName: string }) =>
      commitRepository(repoName),
    ...options,
  });
};
