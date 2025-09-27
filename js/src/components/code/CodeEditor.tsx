import { Editor } from "@monaco-editor/react";

import FileTree from "@/components/code/tree/FileTree";
import { cn } from "@/lib/utils";

import type { CodeEditorProps } from "./types";

export default function CodeEditor({
  files,
  selectedFile,
  className,
  title,
  onFileSelected,
}: CodeEditorProps) {
  const getFileContent = () => {
    if (!selectedFile) {
      return "// Select a file from the tree to view its content\n// Or click 'Resolve Conflicts' to start resolving merge conflicts";
    }

    // For now, return a placeholder content for the selected file
    // In a real implementation, you'd fetch the actual file content
    return `// File: ${selectedFile.name}\n// Path: ${
      selectedFile.fullPath
    }\n// Type: ${
      selectedFile.extension
    }\n\n// This is where the file content would be displayed\n// In a real implementation, you would:\n// 1. Fetch the actual file content from the backend\n// 2. Display it here\n// 3. Allow editing if needed\n\n${
      selectedFile.isConflicted
        ? "// ⚠️ This file has merge conflicts!"
        : "// ✅ No conflicts in this file"
    }`;
  };

  return (
    <div className={cn("flex w-full h-full", className)}>
      <FileTree
        title={title}
        files={files}
        selectedFile={selectedFile}
        onFileSelected={onFileSelected}
      />
      <div className="flex-1 h-full bg-white dark:bg-gray-900">
        <Editor
          className="h-full"
          value={getFileContent()}
          language={selectedFile?.extension?.toLowerCase() || "plaintext"}
          options={{
            minimap: { enabled: true },
            wordWrap: "on",
            lineNumbers: "on",
            readOnly: false,
            automaticLayout: true,
          }}
        />
      </div>
    </div>
  );
}
