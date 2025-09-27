import { useState } from "react";
import { toast } from "sonner";

import type { CodeFile } from "@/lib/api/types/code";

import CodeModal from "@/components/code/bottombuttons/CodeModal";
import CodeEditor from "@/components/code/CodeEditor";
import DiffEditor from "@/components/code/DiffEditor";
import MergeConflictButton from "@/components/code/MergeConflictButton";
import { Button } from "@/components/ui/button";
import { useFileTreeQuery } from "@/lib/api/queries/auth";
import { useCommitRepositoryMutation } from "@/lib/api/queries/github";
import { useStream } from "@/lib/hooks/useStream";

export default function RootPage() {
  const { data, status } = useFileTreeQuery("NewsTrusty");
  const [codeString, setCodeString] = useState<string>("");
  const [selectedFile, setSelectedFile] = useState<CodeFile | undefined>(
    undefined
  );
  const [resolvedCode, setResolvedCode] = useState<string>("");
  const [currentEditorContent, setCurrentEditorContent] = useState<string>("");

  const commitMutation = useCommitRepositoryMutation();
  const { startStream } = useStream({
    onChunk: (chunk) => {
      setCodeString((prev) => prev + chunk);
    },
    onComplete: (fullText) => {
      setCodeString(fullText);
      setResolvedCode(fullText);
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

  const handleEditorChange = (
    value: string | undefined,
    file: CodeFile | undefined
  ) => {
    console.log("Editor content changed:", { value, file: file?.name });
    setCurrentEditorContent(value || "");
  };

  if (status === "pending") {
    return <>Please put a loader here...</>;
  }

  if (status === "error") {
    return <>Failed to fetch repository tree</>;
  }

  const files = data;

  return (
    <>
      <div>
        <div className="absolute top-4 right-42 z-50">
          <div className="relative">
            <Button
              onClick={() => {
                if (!selectedFile) {
                  toast.error("Please select a file first");
                  return;
                }
                if (!currentEditorContent.trim()) {
                  toast.error("File content is empty");
                  return;
                }
                commitMutation.mutate(
                  {
                    repoName: "NewsTrusty",
                    newFileData: currentEditorContent,
                    path: selectedFile.fullPath,
                  },
                  {
                    onSuccess: () => {
                      toast.success("Repository committed successfully!");
                    },
                    onError: (error) => {
                      toast.error(
                        error.message || "Failed to commit repository"
                      );
                    },
                  }
                );
              }}
            >
              Commit Changes
            </Button>
          </div>
        </div>
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
          onChange={handleEditorChange}
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
