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

    // If the file has code content, display it
    if (selectedFile.code) {
      return selectedFile.code;
    }

    // Fallback to placeholder content if no code is available
    return `// File: ${selectedFile.name}\n// Path: ${
      selectedFile.fullPath
    }\n// Type: ${
      selectedFile.extension
    }\n\n// No code content available for this file\n// In a real implementation, you would:\n// 1. Fetch the actual file content from the backend\n// 2. Display it here\n// 3. Allow editing if needed\n\n${
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
