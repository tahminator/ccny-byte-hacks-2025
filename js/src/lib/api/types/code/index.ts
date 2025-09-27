export type CodeExtension = "JS" | "TS" | "HTML" | "CSS" | string;

export type CodeFile = {
  type: "FILE";
  name: string;
  fullPath: string;
  extension: CodeExtension;
  isConflicted: boolean;
  code?: string;
};

export type CodeDirectory = {
  type: "DIRECTORY";
  name: string;
  fullPath: string;
  subDirectories?: (CodeDirectory | CodeFile)[];
};
