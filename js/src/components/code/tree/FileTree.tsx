import { useMemo } from "react";

import type { CodeDirectory, CodeFile } from "@/lib/api/types/code";

import { ModeToggle } from "@/components/modeToggle";
import { Badge } from "@/components/ui/badge";

import type { FileTreeProps } from "./types";

import FileTreeNode from "./node/FileTreeNode";

export default function FileTree({
  files,
  selectedFile,
  title,
  onFileSelected,
}: FileTreeProps) {
  const conflictCount = useMemo(() => {
    if (!files) return 0;

    const countInNode = (node: CodeDirectory | CodeFile): number => {
      if (node.type === "DIRECTORY" && node.subDirectories) {
        return node.subDirectories.reduce(
          (sum, child) => sum + countInNode(child),
          0,
        );
      }
      if (node.type === "FILE") {
        return node.isConflicted ? 1 : 0;
      }
      return 0;
    };

    return files.reduce((sum, file) => sum + countInNode(file), 0);
  }, [files]);

  if (!files) return <></>;

  return (
    <div className="h-full bg-card border-r border-border flex flex-col min-w-1/6">
      <div className="p-3 border-b border-border">
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-sm font-semibold">{title}</h2>
          <ModeToggle />
        </div>
        {conflictCount > 0 && (
          <div className="flex justify-center">
            <Badge
              variant="outline"
              className="text-xs border-conflict-current text-conflict-current"
            >
              {conflictCount} conflicts
            </Badge>
          </div>
        )}
      </div>
      <div className="flex-1 overflow-y-auto">
        {files.map((file, k) => (
          <FileTreeNode
            level={0}
            key={k}
            node={file}
            selectedFile={selectedFile}
            onFileSelected={onFileSelected}
          />
        ))}
      </div>
    </div>
  );
}
