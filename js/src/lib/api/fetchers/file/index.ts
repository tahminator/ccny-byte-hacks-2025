import type { CodeDirectory, CodeFile } from "@/lib/api/types/code";

export async function getFileTree(userId: string, repoName: string) {
  const res = await fetch(
    `/api/file/tree/generate?userId=${userId}&repoName=${repoName}`,
  );

  // TODO - Re-write endpoint to use ApiResponder.
  if (!res.ok) {
    throw new Error("failed to get file tree");
  }

  return (await res.json()) as (CodeFile | CodeDirectory)[];
}
