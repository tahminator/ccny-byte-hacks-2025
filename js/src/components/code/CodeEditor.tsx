import { Editor } from "@monaco-editor/react";
import { useEffect, useState } from "react";

import { REPO_NAME, USER_NAME } from "@/app/Root.page";
import FileTree from "@/components/code/tree/FileTree";
import { useFileQuery } from "@/lib/api/queries/file";
import { cn } from "@/lib/utils";
import { createHighlighter } from "shiki"

import type { CodeEditorProps } from "./types";

const hl = await createHighlighter({
  themes: [
      "github-dark",
  ],
  langs: [
      "python"
  ]
})

export default function CodeEditor({
  files,
  selectedFile,
  className,
  title,
  githubUsername,
  githubRepo,
  onFileSelected,
  onChange,
}: CodeEditorProps) {
  const [monacoTheme, setMonacoTheme] = useState("");

  useEffect(() => {
    setTimeout(() => {
      setMonacoTheme("vs-dark");
    }, 50);
  }, []);

  const {
    data: fileContent,
    isLoading,
    error,
    status,
  } = useFileQuery(
    // TODO: Replace with actual username and repo
    githubUsername || USER_NAME,
    githubRepo || REPO_NAME,
    selectedFile?.fullPath,
  );

  useEffect(() => {
    if (status === "success") {
      onChange?.(fileContent);
    }
  }, [fileContent, onChange, status]);

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

  const handleEditorChange = (value: string | undefined) => {
    if (onChange) {
      onChange(value);
    }
  };

  return (
    <div className={cn("flex w-full h-full", className)}>
      <FileTree
        title={title}
        files={files}
        selectedFile={selectedFile}
        onFileSelected={onFileSelected}
      />
      <div
        className={cn(
          "flex-1 h-full",
          monacoTheme === "vs-dark" ? "bg-[#1e1e1e]" : "bg-[#ffffff]",
        )}
      >
        <Editor
          className="h-full"
          value={getFileContent()}
          key={selectedFile?.fullPath || "default"}
          language={"python"}
          onChange={handleEditorChange}
          theme={monacoTheme}
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
