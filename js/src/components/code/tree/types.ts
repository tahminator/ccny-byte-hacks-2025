import type { CodeDirectory, CodeFile } from "@/lib/api/types/code";

export type FileTreeProps = {
  title?: string;
  files?: (CodeDirectory | CodeFile)[];
  selectedFile?: CodeFile;
  onFileSelected?: (file: CodeFile) => void;
};
