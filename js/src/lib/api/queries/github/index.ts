import { useMutation, useQueryClient } from "@tanstack/react-query";

import { commitRepository, declineMerge } from "../../fetchers/github";

export const useCommitRepositoryMutation = () => {
  const queryClient = useQueryClient();

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
    onSettled: () => {
      queryClient.invalidateQueries({});
    },
  });
};

export const useDeclineMergeMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      fullPath,
      repoName,
    }: {
      fullPath: string;
      repoName: string;
    }) => declineMerge(fullPath, repoName),
    onSettled: () => {
      queryClient.invalidateQueries({});
    },
  });
};
