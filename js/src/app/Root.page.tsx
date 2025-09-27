import { useCallback, useState } from "react";
import { toast } from "sonner";

import type { CodeFile } from "@/lib/api/types/code";

import CodeModal from "@/components/code/bottombuttons/CodeModal";
import CodeEditor from "@/components/code/CodeEditor";
import DiffEditor from "@/components/code/DiffEditor";
import MergeConflictButton from "@/components/code/MergeConflictButton";
import { Button } from "@/components/ui/button";
import { useFileTreeQuery } from "@/lib/api/queries/auth";
import {
  useAcceptMergeMutation,
  useCommitRepositoryMutation,
} from "@/lib/api/queries/github";
import { useStream } from "@/lib/hooks/useStream";

export default function RootPage() {
  const { data, status } = useFileTreeQuery("NewsTrusty");
  const [codeString, setCodeString] = useState<string>("");
  const [selectedFile, setSelectedFile] = useState<CodeFile | undefined>(
    undefined,
  );
  const [resolvedCode, setResolvedCode] = useState<string>("");
  const [currentEditorContent, setCurrentEditorContent] = useState<string>("");

  const commitMutation = useCommitRepositoryMutation();
  const acceptMergeMutation = useAcceptMergeMutation();
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
      conflictContent: currentEditorContent,
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
    resolvedCodeContent: string,
  ) => {
    console.log("Accepting resolved code for:", filePath);
    console.log("Resolved code:", resolvedCodeContent);

    acceptMergeMutation.mutate({
      newFileData: resolvedCodeContent,
      fullPath: selectedFile?.fullPath ?? "",
      repoName: "NewsTrusty",
    });

    setResolvedCode(""); // Clear the modal
    setCodeString(""); // Clear the diff editor

    // TODO: Refetch the file list here when endpoint is ready
    console.log("Code accepted! (API call will be added later)");
  };

  const handleRejectResolvedCode = () => {
    setResolvedCode(""); // Clear the modal
    setCodeString(""); // Clear the diff editor
  };

  const handleEditorChange = useCallback((value: string | undefined) => {
    // console.log("Editor content changed:", { value, file: file?.name });
    setCurrentEditorContent(value || "");
  }, []);

  if (status === "pending") {
    return <>Please put a loader here...</>;
  }

  if (status === "error") {
    return <>Failed to fetch repository tree</>;
  }

  const files = data;

  return (
    <>
      <div className="fixed top-0 left-0 w-full h-16 bg-[#181818] border-b-2 border-[#3c3c3c] shadow-2xl drop-shadow-2xl z-50">
        <div className="absolute top-2 left-18">
          <img 
            src="/delta.png" 
            alt="Delta Logo" 
            className="h-16 w-17 object-contain bg-transparent"
          />
        </div>
        <div className="absolute top-4 right-42">
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
                        error.message || "Failed to commit repository",
                      );
                    },
                  },
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
      
      <div className="relative flex w-[100vw] h-[100vh] justify-center items-start pt-16 shadow-[0_25px_50px_-12px_rgba(0,0,0,0.8)] drop-shadow-2xl">
        <CodeEditor
          className="max-w-full shadow-[0_25px_50px_-12px_rgba(0,0,0,0.8)] drop-shadow-2xl border-t-2 border-[#3c3c3c]"
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
