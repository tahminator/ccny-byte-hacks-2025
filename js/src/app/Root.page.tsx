import { Button } from "@/components/ui/button";
import { useState } from "react";

export default function RootPage() {
  const [count, setCount] = useState(0);

  return (
    <Button
      onClick={() => {
        setCount((prev) => ++prev);
      }}
    >
      {count}
    </Button>
  );
}
