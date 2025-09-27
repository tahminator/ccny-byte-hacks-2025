import { useState } from "react";

import type { CodeDirectory, CodeFile } from "@/lib/api/types/code";

import CodeEditor from "@/components/code/CodeEditor";
import MergeConflictButton from "@/components/code/MergeConflictButton";

import testFiles from "./test.json";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const files: (CodeFile | CodeDirectory)[] = testFiles as unknown as any;

export default function RootPage() {
  const [selectedFile, setSelectedFile] = useState<CodeFile | null>(null);

  const handleResolveConflict = (context: string) => {
    console.log("Resolving merge conflict with context:", context);
    console.log("Selected file:", selectedFile);
  };

  const handleFileSelected = (file: CodeFile) => {
    console.log("File selected:", JSON.stringify(file));
    setSelectedFile(file);
  };

  return (
    <div className="relative flex w-[100vw] h-[100vh] justify-center items-start pt-16">
      <MergeConflictButton
        selectedFile={selectedFile}
        onResolveConflict={handleResolveConflict}
      />
      <CodeEditor
        files={files}
        title={"My Project"}
        onFileSelected={handleFileSelected}
      />
      <div className="flex min-w-1/6"></div>
    </div>
  );
}
