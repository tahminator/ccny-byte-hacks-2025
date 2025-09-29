import type { ReactNode, SVGProps } from "react";

import { FolderOpen, GitBranch } from "lucide-react";

import { cn } from "@/lib/utils";

import type { FileTreeNodeProps } from "./types";

import { mapLanguageToIcon } from "../../FileIcon";

export default function FileTreeNode({
  node,
  selectedFile,
  onFileSelected = () => {},
  level = 0,
}: FileTreeNodeProps) {
  if (!node) return <></>;

  let CodeIcon: (props: SVGProps<SVGSVGElement>) => ReactNode = () => <></>;

  if (node.type === "FILE") {
    CodeIcon = mapLanguageToIcon(node.extension);
  }

  if (selectedFile === node) {
    console.log("FOUND IT");
  }

  return (
    <div>
      <div
        className={cn(
          "flex items-center gap-1 py-1 px-2 text-sm cursor-pointer hover:bg-gray-400 group",
          selectedFile === node && "bg-gray-300 text-black",
          node.type === "FILE" && node.isConflicted && "text-conflict-current",
        )}
        style={{ paddingLeft: `${level * 12 + 8}px` }}
        onClick={() => {
          if (node.type === "FILE") {
            onFileSelected(node);
          }
        }}
      >
        {node.type === "DIRECTORY" ?
          <FolderOpen className="h-4 w-4 text-blue-500" />
        : <CodeIcon
            className={cn(
              "h-4 w-4",
              node.isConflicted ?
                "text-conflict-current"
              : "text-muted-foreground",
            )}
          />
        }
        <span className="flex-1 truncate">{node.name}</span>
        {node.type === "FILE" && node.isConflicted && (
          <GitBranch className="h-4 w-4 text-red-500" />
        )}
      </div>
      {node.type === "DIRECTORY" && node.subDirectories && (
        <div>
          {node.subDirectories.map((child, key) => (
            <FileTreeNode
              key={key}
              node={child}
              level={level + 1}
              selectedFile={selectedFile}
              onFileSelected={onFileSelected}
            />
          ))}
        </div>
      )}
    </div>
  );
}
