import { Editor } from "@monaco-editor/react";

import FileTree from "@/components/code/tree/FileTree";

import type { CodeEditorProps } from "./types";

export default function CodeEditor({
  files,
  selectedFile,
  code,
  title,
  onFileSelected,
}: CodeEditorProps) {
  return (
    <>
      <FileTree
        title={title}
        files={files}
        selectedFile={selectedFile}
        onFileSelected={onFileSelected}
      />
      <Editor className="max-w-5/6" defaultValue={code} />
    </>
  );
}
