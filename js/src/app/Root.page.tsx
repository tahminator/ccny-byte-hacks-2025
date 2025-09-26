import { useState } from "react";

import { Button } from "@/components/ui/button";

export default function RootPage() {
  const [count, setCount] = useState(0);

  return (
    <Button
      className="bg-black text-white rounded-md p-3"
      onClick={() => {
        setCount((prev) => ++prev);
      }}
    >
      {count}
    </Button>
  );
}
