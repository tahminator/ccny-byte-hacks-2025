import type { CodeDirectory, CodeFile } from "@/lib/api/types/code";

import CodeEditor from "@/components/code/CodeEditor";

const files: (CodeFile | CodeDirectory)[] = [
  {
    type: "FILE",
    extension: "JS",
    name: "hello.js",
    fullPath: "hello.js",
    isConflicted: false,
  },
  {
    type: "DIRECTORY",
    name: "hello",
    fullPath: "hello",
    subDirectories: [
      {
        type: "FILE",
        extension: "JS",
        name: "hello.js",
        fullPath: "hello/hello.js",
        isConflicted: true,
      },
      {
        type: "FILE",
        extension: "TS",
        name: "hello.ts",
        fullPath: "hello/hello.ts",
        isConflicted: true,
      },
      {
        type: "FILE",
        extension: "HTML",
        name: "hello.html",
        fullPath: "hello/hello.html",
        isConflicted: true,
      },
      {
        type: "FILE",
        extension: "CSS",
        name: "hello.css",
        fullPath: "hello/hello.css",
        isConflicted: true,
      },
    ],
  },
];

export default function RootPage() {
  return (
    <div className="flex w-[100vw] h-[100vh] justify-center items-center">
      <CodeEditor
        files={files}
        title={"My Project"}
        onFileSelected={(file) => console.log(JSON.stringify(file))}
      />
      <div className="flex min-w-1/6"></div>
    </div>
  );
}
