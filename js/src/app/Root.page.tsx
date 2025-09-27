import { useState } from "react";

import type { CodeDirectory, CodeFile } from "@/lib/api/types/code";

import CodeModal from "@/components/code/bottombuttons/CodeModal";
import CodeEditor from "@/components/code/CodeEditor";
import DiffEditor from "@/components/code/DiffEditor";
import MergeConflictButton from "@/components/code/MergeConflictButton";
import { useStream } from "@/lib/hooks/useStream";

import testFiles from "./test2.json";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const files: (CodeFile | CodeDirectory)[] = testFiles as unknown as any;

export default function RootPage() {
  const [codeString, setCodeString] = useState<string>("");
  const [selectedFile, setSelectedFile] = useState<CodeFile | undefined>(
    undefined
  );
  const [resolvedCode, setResolvedCode] = useState<string>("");

  const { startStream } = useStream({
    onChunk: (chunk) => {
      setCodeString((prev) => prev + chunk);
    },
    onComplete: (fullText) => {
      setCodeString(fullText);
      setResolvedCode(fullText); // Set resolved code for the modal
    },
  });

  const handleResolveConflict = async (context: string) => {
    console.log("Resolving merge conflict with context:", context);
    console.log("Selected file:", selectedFile);

    setCodeString("");

    await startStream({
      conflictContent: "",
      filePath: selectedFile?.fullPath || "",
      userQuery: context || "Please help resolve this merge conflict",
    });
  };

  const handleFileSelected = (file: CodeFile) => {
    console.log("File selected:", JSON.stringify(file));
    setSelectedFile(file);
  };

  const handleAcceptResolvedCode = async (
    filePath: string,
    resolvedCodeContent: string
  ) => {
    console.log("Accepting resolved code for:", filePath);
    console.log("Resolved code:", resolvedCodeContent);

    // TODO: Add API call to save the resolved code when endpoint is ready
    // const result = await saveResolvedCode(filePath, resolvedCodeContent);

    setResolvedCode(""); // Clear the modal
    setCodeString(""); // Clear the diff editor

    // TODO: Refetch the file list here when endpoint is ready
    console.log("Code accepted! (API call will be added later)");
  };

  const handleRejectResolvedCode = () => {
    setResolvedCode(""); // Clear the modal
    setCodeString(""); // Clear the diff editor
  };
  return (
    <>
      <div>
        <MergeConflictButton
          selectedFile={selectedFile}
          onResolveConflict={handleResolveConflict}
        />
      </div>
      <div className="relative flex w-[100vw] h-[100vh] justify-center items-start pt-16">
        <CodeEditor
          className="max-w-full"
          files={files}
          selectedFile={selectedFile}
          title={"My Project"}
          onFileSelected={handleFileSelected}
        />
        <DiffEditor code={codeString} />
      </div>
      <CodeModal
        resolvedCode={resolvedCode}
        originalFilePath={selectedFile?.fullPath || ""}
        isVisible={!!resolvedCode}
        onAccept={handleAcceptResolvedCode}
        onReject={handleRejectResolvedCode}
      />
    </>
  );
}
