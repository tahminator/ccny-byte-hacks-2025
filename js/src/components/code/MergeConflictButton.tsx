import { useState } from "react";

import type { CodeFile } from "@/lib/api/types/code";

import { Button } from "@/components/ui/button";

interface MergeConflictButtonProps {
  selectedFile: CodeFile | null;
  onResolveConflict?: (context: string) => void;
}

export default function MergeConflictButton({
  selectedFile,
  onResolveConflict,
}: MergeConflictButtonProps) {
  const [showConflictDropdown, setShowConflictDropdown] = useState(false);
  const [mergeContext, setMergeContext] = useState("");

  const handleMergeConflict = () => {
    onResolveConflict?.(mergeContext);
    setShowConflictDropdown(false);
    setMergeContext("");
  };

  if (!selectedFile?.isConflicted) {
    return null;
  }

  return (
    <div className="absolute top-4 right-4 z-50">
      <div className="relative">
        <Button
          className="bg-green-600 hover:bg-green-700 text-white"
          onClick={() => setShowConflictDropdown(!showConflictDropdown)}
        >
          Resolve Conflicts
        </Button>
        {showConflictDropdown && (
          <div className="absolute top-full right-0 mt-2 w-80 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-md shadow-lg p-4 z-50">
            <div className="mb-3">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Merge Context (Optional)
              </label>
              <textarea
                className="w-full h-24 p-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md resize-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white dark:bg-gray-700 dark:text-gray-200"
                placeholder="Describe how you want the merge to be resolved..."
                value={mergeContext}
                onChange={(e) => setMergeContext(e.target.value)}
              />
            </div>
            <div className="flex gap-2 justify-end">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowConflictDropdown(false)}
              >
                Cancel
              </Button>
              <Button
                size="sm"
                className="bg-green-600 hover:bg-green-700"
                onClick={handleMergeConflict}
              >
                Resolve
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
