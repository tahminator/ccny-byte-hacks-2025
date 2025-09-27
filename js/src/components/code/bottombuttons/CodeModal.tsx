import { useState } from "react";

import AcceptButton from "./AcceptButton";
import RejectButton from "./RejectButton";

interface CodeModalProps {
  resolvedCode: string;
  originalFilePath: string;
  isVisible: boolean;
  onAccept: (filePath: string, resolvedCode: string) => Promise<void>;
  onReject: () => void;
}

export default function CodeModal({
  resolvedCode,
  originalFilePath,
  isVisible,
  onAccept,
  onReject,
}: CodeModalProps) {
  const [isLoading, setIsLoading] = useState(false);

  const handleAccept = async () => {
    setIsLoading(true);
    try {
      await onAccept(originalFilePath, resolvedCode);
    } catch (error) {
      console.error("Failed to save resolved code:", error);
    } finally {
      setIsLoading(false);
    }
  };

  if (!isVisible || !resolvedCode) {
    return null;
  }

  return (
    <div className="fixed bottom-12 right-8 bg-white dark:bg-gray-800 rounded-lg shadow-lg px-6 py-4 flex gap-6 z-[9999] border border-gray-200 dark:border-gray-700">
      <RejectButton 
        onClick={onReject}
        className="px-4 py-2"
      />
      <AcceptButton 
        onClick={handleAccept}
        isLoading={isLoading}
        className="px-4 py-2"
      />
    </div>
  );
}