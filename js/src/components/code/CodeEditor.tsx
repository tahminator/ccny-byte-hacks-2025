import { Editor } from "@monaco-editor/react";

import FileTree from "@/components/code/tree/FileTree";
import { useFileQuery } from "@/lib/api/queries/file";
import { cn } from "@/lib/utils";

import type { CodeEditorProps } from "./types";

export default function CodeEditor({
  files,
  selectedFile,
  className,
  title,
  githubUsername,
  githubRepo,
  onFileSelected,
}: CodeEditorProps) {
  // Use the useFileQuery hook to fetch file content
  const {
    data: fileContent,
    isLoading,
    error,
  } = useFileQuery(
    // TODO: Replace with actual username and repo
    githubUsername || "manofshad",
    githubRepo || "NewsTrusty",
    selectedFile?.fullPath,
  );

  const getFileContent = () => {
    if (!selectedFile) {
      return "// Select a file from the tree to view its content\n// Or click 'Resolve Conflicts' to start resolving merge conflicts";
    }

    if (isLoading) {
      return "// Loading file content...";
    }

    if (error) {
      return `// Error loading file: ${error.message}\n// File: ${selectedFile.name}\n// Path: ${selectedFile.fullPath}`;
    }

    if (!fileContent) {
      return `// No content available for this file\n// File: ${selectedFile.name}\n// Path: ${selectedFile.fullPath}`;
    }

    return fileContent;
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
