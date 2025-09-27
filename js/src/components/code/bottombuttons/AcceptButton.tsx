import { Check } from "lucide-react";

import { Button } from "@/components/ui/button";

interface AcceptButtonProps {
  onClick: () => void;
  className?: string;
  isLoading?: boolean;
}

export default function AcceptButton({ onClick, className = "", isLoading = false }: AcceptButtonProps) {
  return (
    <Button
      onClick={onClick}
      variant="default"
      size="sm"
      disabled={isLoading}
      className={`bg-green-600 hover:bg-green-700 text-white flex items-center gap-1 ${className}`}
    >
      <Check className="w-4 h-4" />
      {isLoading ? "Saving..." : "Accept"}
    </Button>
  );
}