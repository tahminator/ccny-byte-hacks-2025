import type { CodeDirectory, CodeFile } from "@/lib/api/types/code";

export async function getFileTree(repoName: string) {
  const res = await fetch(`/api/file/tree/generate?repoName=${repoName}`);

  // TODO - Re-write endpoint to use ApiResponder.
  if (!res.ok) {
    throw new Error("failed to get file tree");
  }

  return (await res.json()) as (CodeFile | CodeDirectory)[];
}

export async function getFile(
  githubUsername: string,
  githubRepo: string,
  filePath: string,
) {
  const res = await fetch(
    `/api/file/data/${githubUsername}/${githubRepo}/${filePath}`,
  );

  if (!res.ok) {
    throw new Error("failed to get file");
  }

  return await res.text();
}
