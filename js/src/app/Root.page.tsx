import type { CodeDirectory, CodeFile } from "@/lib/api/types/code";

import CodeEditor from "@/components/code/CodeEditor";

import testFiles from "./test.json";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const files: (CodeFile | CodeDirectory)[] = testFiles as unknown as any;

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
