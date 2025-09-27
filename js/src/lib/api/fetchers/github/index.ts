export async function commitRepository(
  repoName: string,
  newFileData: string,
  path: string
) {
  const res = await fetch("/api/github/commit", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ repoName, newFileData, path }),
  });

  if (!res.ok) {
    const errorData = await res
      .json()
      .catch(() => ({ error: "Failed to commit repository" }));
    throw new Error(errorData.error || "Failed to commit repository");
  }

  return await res.json();
}

export async function declineMerge(fullPath: string, repoName: string) {
  const res = await fetch("/api/github/merge/decline", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ fullPath, repoName }),
  });

  if (!res.ok) {
    const errorData = await res
      .json()
      .catch(() => ({ error: "Failed to decline merge" }));
    throw new Error(errorData.error || "Failed to decline merge");
  }

  return await res.json();
}
