import { Editor } from "@monaco-editor/react";

interface DiffEditorProps {
  code: string;
}

export default function DiffEditor({ code }: DiffEditorProps) {
  if (!code || code.trim() === "") {
    return null;
  }

  return (
    <div
      className="h-full border-l border-border bg-white dark:bg-gray-900"
      style={{
        width: "500px",
        minWidth: "500px",
        maxWidth: "500px",
        position: "absolute",
        right: "0",
        top: "0",
        zIndex: 10,
      }}
    >
      <Editor
        className="h-full"
        value={code}
        options={{
          readOnly: true,
          minimap: { enabled: false },
        }}
      />
    </div>
  );
}
