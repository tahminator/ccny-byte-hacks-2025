import { Editor } from "@monaco-editor/react";

import { cn } from "@/lib/utils";

interface DiffEditorProps {
  code: string;
}

export default function DiffEditor({ code }: DiffEditorProps) {
  return (
    <Editor
      className={cn("h-full", code ? "h-full max-w-full" : "hidden")}
      value={code}
      defaultLanguage={"python"}
      options={{
        readOnly: false,
        minimap: { enabled: false },
      }}
    />
  );
}
