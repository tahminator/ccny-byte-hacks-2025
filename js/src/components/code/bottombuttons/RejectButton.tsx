import { X } from "lucide-react";

import { Button } from "@/components/ui/button";

interface RejectButtonProps {
  onClick: () => void;
  className?: string;
}

export default function RejectButton({ onClick, className = "" }: RejectButtonProps) {
  return (
    <Button
      onClick={onClick}
      variant="destructive"
      size="sm"
      className={`bg-red-600 hover:bg-red-700 text-white flex items-center gap-1 ${className}`}
    >
      <X className="w-4 h-4" />
      Reject
    </Button>
  );
}