import { useState } from "react";

import type { CodeDirectory, CodeFile } from "@/lib/api/types/code";

import CodeEditor from "@/components/code/CodeEditor";
import DiffEditor from "@/components/code/DiffEditor";
import MergeConflictButton from "@/components/code/MergeConflictButton";
import { useStream } from "@/lib/hooks/useStream";

import testFiles from "./test.json";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const files: (CodeFile | CodeDirectory)[] = testFiles as unknown as any;

export default function RootPage() {
  const [selectedFile, setSelectedFile] = useState<CodeFile | null>(null);
  const [codeString, setCodeString] = useState<string>("");

  const { startStream } = useStream({
    onChunk: (chunk) => {
      setCodeString((prev) => prev + chunk);
    },
    onComplete: (fullText) => {
      setCodeString(fullText);
    },
  });

  const handleResolveConflict = async (context: string) => {
    console.log("Resolving merge conflict with context:", context);
    console.log("Selected file:", selectedFile);

    // Reset the code string when starting a new resolution
    setCodeString("");

    // Start streaming from the Gemini API
    const message = context || "Please help resolve this merge conflict";
    await startStream("http://localhost:8080/api/gemini/test", { message });
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
      <DiffEditor code={codeString} />
    </div>
  );
}
