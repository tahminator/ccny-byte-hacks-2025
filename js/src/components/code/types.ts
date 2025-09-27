import type { CodeDirectory, CodeFile } from "@/lib/api/types/code";

export type CodeEditorProps = {
  files?: (CodeDirectory | CodeFile)[];
  selectedFile?: CodeFile;
  code?: string;
  title?: string;
  className?: string;
  githubUsername?: string;
  githubRepo?: string;
  onFileSelected?: (file: CodeFile) => void;
  onChange?: (value: string | undefined, file: CodeFile | undefined) => void;
};
