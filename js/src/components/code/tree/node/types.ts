import type { CodeDirectory, CodeFile } from "@/lib/api/types/code";

export type FileTreeNodeProps = {
  node?: CodeDirectory | CodeFile;
  selectedFile?: CodeFile;
  onFileSelected?: (file: CodeFile) => void;
  level: number;
};
