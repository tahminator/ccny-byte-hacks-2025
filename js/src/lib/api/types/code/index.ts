export type CodeExtension =
  | "JS"
  | "TS"
  | "HTML"
  | "CSS"
  | "GO"
  | "PY"
  | "JAVA"
  | "C"
  | "CPP"
  | "RS"
  | "RB"
  | "PHP"
  | "SQL"
  | "TXT"
  | "UNKNOWN"
  | string;

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
